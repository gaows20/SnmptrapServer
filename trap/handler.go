package trap

import (
	"cqrcsnmpserver/common/sender"
	"cqrcsnmpserver/global"
	"cqrcsnmpserver/linklist"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	g "github.com/gosnmp/gosnmp"
	log "github.com/sirupsen/logrus"
)

var parseOIDlist map[string]string = map[string]string{"ifIndex": "1.3.6.1.2.1.31.1.1.1.1."}
var valueMap map[string]map[string]string = map[string]map[string]string{
	"ifOperStatus":  {"1": "up", "2": "down", "3": "testing"},
	"ifAdminStatus": {"1": "up", "2": "down", "3": "testing"},
}

// snmp get实现
func runSnmpGet(target, community, oid string) (string, error) {
	// Create a GoSNMP struct with specified target and community
	var res string = ""
	g.Default.Target = target
	g.Default.Community = community
	// g.Default.Timeout = 5
	err := g.Default.Connect()
	if err != nil {
		log.WithField("err", err).Error("runSnmpGet Connect() err:")
		return "Get() err: %v", err
	}
	defer g.Default.Conn.Close()

	oids := []string{oid}
	result, err2 := g.Default.Get(oids) // Get() accepts up to g.M***_OIDS
	if err2 != nil {
		// return fmt.Errorf("Get() err: %v", err2)
		return "Get() err: %v", err2
	}

	for i, variable := range result.Variables {
		fmt.Printf("%d: oid: %s ", i, variable.Name)

		switch variable.Type {
		case g.OctetString:
			// fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
			res = string(variable.Value.([]byte))
			return res, nil
		default:
			// fmt.Printf("number: %d\n", g.ToBigInt(variable.Value))
			res = string(variable.Value.([]byte))
			return res, nil
		}
	}

	return res, nil
}

func genMsgHeader(hostip string, packet *g.SnmpPacket) (msg string) {
	msg = ""
	msg = msg + fmt.Sprintf("%s CQRCB_SYSTEM %s\n", time.Now().Format("2006-01-02 15:04:05"), hostip)
	msg = msg + "PDU INFO:\n"
	msg = msg + "  MESSAGE TYPE: SNMP TRAP \n"
	msg = msg + fmt.Sprintf("  VERSION:  %s\n", packet.Version)
	msg = msg + fmt.Sprintf("  FROM: [%s]\n", hostip)
	msg = msg + fmt.Sprintf("  STATUS: [%s]\n", packet.Error)
	msg = msg + fmt.Sprintf("  MESSAGE ID: [%v]\n", packet.MsgID)
	msg = msg + fmt.Sprintf("  COMMUNITY: [%s]\n", packet.Community)
	msg = msg + fmt.Sprintf("  INDEX: [%v]\n", packet.ErrorIndex)
	msg = msg + fmt.Sprintf("  REQUEST ID: [%v]\n", packet.RequestID)
	msg = msg + "TRAP VARIABLES:\n"
	return msg
}

// var once sync.Once
func parseSnmpPack(hostip string, list *linklist.List, packet *g.SnmpPacket) {

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 错误处理
		return
	}
	bjTime := time.Now().In(loc)
	msg := genMsgHeader(hostip, packet)

	pdus := make([]*TrapPDU, 0)

	// 迭代解析每一个snmp报文
	for _, v := range packet.Variables {
		oidName := ""
		if name, err := global_mib_tree.FindNodeName(v.Name); err != nil {
			log.WithField("err", err).Error("没有找到OID解析")
			oidName = v.Name
		} else {
			oidName = name
		}
		switch v.Type {
		case g.OctetString:
			b := v.Value.([]byte)
			log.WithField("OID", v.Name).WithField("string", fmt.Sprintf("%s", b)).WithField("Type", v.Type).Info()
			pdu := TrapPDU{
				OID:    oidName,
				RawOID: v.Name,
				Type:   v.Type,
				Value:  fmt.Sprintf("%s", b),
				// Ts:    time.Now().Format("2006-01-02 15:04:05"),
				Ts:         bjTime.Format("2006-01-02 15:04:05"),
				ParseValue: "",
			}
			pdus = append(pdus, &pdu)
		// 嵌套OID
		case g.ObjectIdentifier:
			obj_id := fmt.Sprintf("%s", v.Value)
			obj_name := ""
			parse_value := ""
			if name, err := global_mib_tree.FindNodeName(obj_id); err != nil {
				log.WithField("err", err).Error("trans oid to name error")
				obj_name = obj_id
			} else {
				obj_name = name
			}
			pdu := TrapPDU{
				OID:    oidName,
				RawOID: obj_id,
				Type:   v.Type,
				Value:  obj_name,
				// Ts:    time.Now().Format("2006-01-02 :04:05"),
				Ts:         bjTime.Format("2006-01-02 15:04:05"),
				ParseValue: parse_value,
			}
			pdus = append(pdus, &pdu)
		default:
			// 额外解析字段，这里将把ifIndex字段翻译成ifName最后填入ParseValue
			value := v.Value
			parse_value := ""
			parts := strings.Split(oidName, ".")
			index_v, ok := parseOIDlist[parts[0]]
			if ok {
				// fmt.Println(parts, "存在于 parseOIDlist 中")
				get_value, err := runSnmpGet(hostip, global.GVA_CONFIG.TrapServer.ReadCommunity, index_v+parts[1])
				if err != nil {
					fmt.Printf("querySnmp() err: %v", err)
					// log.Fatalf("querySnmp() err: %v", err)
					parse_value = fmt.Sprintf("%v, community: %v", err, global.GVA_CONFIG.TrapServer.ReadCommunity)
				} else {
					parse_value = get_value
				}
			}
			// 值映射 value map
			map_v, ok := valueMap[parts[0]]
			if ok {
				value = map_v[fmt.Sprintf("%v", v.Value)]
			}

			pdu := TrapPDU{
				OID:    oidName,
				RawOID: v.Name,
				Type:   v.Type,
				Value:  value,
				// Ts:    time.Now().Format("2006-01-02 :04:05"),
				Ts:         bjTime.Format("2006-01-02 15:04:05"),
				ParseValue: parse_value,
			}
			pdus = append(pdus, &pdu)
		}
	}
	tags := paesePdusToListMap(pdus)
	for _, tag := range tags {
		item := fmt.Sprintf("  SNMP-MIB::%v  value=%v: [%v]\n", tag["oid"], tag["type"], tag["value"])
		msg = msg + item
	}

	list.Append(msg)
	var push_msg global.PushMessage
	push_msg.Host = hostip
	push_msg.Message = tags
	push_msg.Version = fmt.Sprintf("%v", packet.Version)
	push_msg.Status = fmt.Sprintf("%v", packet.Error)
	push_msg.MessageID = fmt.Sprintf("%v", packet.MsgID)
	push_msg.Index = fmt.Sprintf("%v", packet.ErrorIndex)
	// sender.Sends(hostip, msg)
	sender.Sends(hostip, push_msg)
}

func paesePdusToListMap(pdus []*TrapPDU) []map[string]string {
	res := make([]map[string]string, 0)
	for _, v := range pdus {
		res = append(res, map[string]string{"oid": v.OID, "type": fmt.Sprintf("%v", v.Type), "value": fmt.Sprintf("%v", v.Value), "ts": v.Ts, "raw_oid": v.RawOID, "parse_value": v.ParseValue})
	}
	return res
}

func BaseTrapHandler(packet *g.SnmpPacket, addr *net.UDPAddr) {
	log.WithField("addr", addr.IP.String()).Info("got trap package from")
	curList, ok := TrapMap[addr.IP.String()]
	if ok {
		log.WithField("address", addr.IP.String()).Info("address is exits.")
		parseSnmpPack(addr.IP.String(), curList, packet)
	} else {
		newList := linklist.NewList()
		TrapMap[addr.IP.String()] = newList
		parseSnmpPack(addr.IP.String(), newList, packet)
	}
}

func DelItem(ip string, index int64) error {
	curlist, ok := TrapMap[ip]
	if ok {
		curlist.RemoveAtIndex(index)
		if curlist.Length() <= 0 {
			delete(TrapMap, ip)
		}
	} else {
		return errors.New(fmt.Sprintf("该IP【%s】没有在trap数据库中", ip))
	}
	return nil
}

func checkItem(data interface{}, pdu interface{}) bool {
	d := data.(TrapPDU)
	p := pdu.(*TrapPDU)
	if d.OID == p.OID && d.Ts == p.Ts && fmt.Sprintf("%v", d.Value) == fmt.Sprintf("%v", p.Value) && fmt.Sprintf("%v", d.Type) == fmt.Sprintf("%v", p.Type) {
		return true
	} else {
		return false
	}
}
