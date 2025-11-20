package trap

import (
	"bufio"
	"cqrcsnmpserver/common/sender"
	"cqrcsnmpserver/global"
	"cqrcsnmpserver/linklist"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	g "github.com/gosnmp/gosnmp"
	log "github.com/sirupsen/logrus"
)

var parseOIDlist map[string]string = map[string]string{
	"ifIndex":          "1.3.6.1.2.1.31.1.1.1.1.",
	"hh3cAggPortIndex": "1.3.6.1.2.1.31.1.1.1.1.",
	// "hh3cAggPortIndex": "1.3.6.1.4.1.25506.8.25.1.2.1.1.",
}
var valueMap map[string]map[string]string = map[string]map[string]string{
	"ifOperStatus":             {"1": "up", "2": "down", "3": "testing"},
	"ipv6IfOperStatus":         {"1": "up", "2": "down", "3": "noIfIdentifier", "4": "unknown", "5": "notPresent"},
	"ifAdminStatus":            {"1": "up", "2": "down", "3": "testing"},
	"hh3cEntityExtAdminStatus": {"1": "notSupported", "2": "locked", "3": "shuttingDown", "4": "unlocked"},
	"hh3cEntityExtAlarmLight":  {"0": "notSupported", "1": "underRepair", "2": "critical", "3": "major", "4": "minor", "5": "alarmOutstanding", "6": "warning", "7": "indeterminate"},
	"bgpPeerState":             {"1": "idle", "2": "connect", "3": "active", "4": "opensent", "5": "openconfirm", "6": "established"},
	"hh3cBgp4v2PeerState":      {"1": "idle", "2": "connect", "3": "active", "4": "opensent", "5": "openconfirm", "6": "established"},
	"hh3cBgpPeerState":         {"1": "idle", "2": "connect", "3": "active", "4": "opensent", "5": "openconfirm", "6": "established"},
	"hh3cBfdSessState":         {"0": "adminDown", "1": "down", "2": "init", "3": "up"},
	"hwBgpPeerState":           {"1": "idle", "2": "connect", "3": "active", "4": "opensent", "5": "openconfirm", "6": "established", "9": "Noneg"},
	// "hwBfdSessDiag":            {"0": "无诊断", "1": "控制检测时间超时", "2": "echo功能故障", "3": "邻居会话信号衰落", "4": "转发平面复位", "5": "路径Down", "6": "接路径Down", "7": "管理Down", "8": "反向连接路径Down", "9": "邻居会话信号衰落(接收admindown)"},
	// "hwBfdSessType":            {"1": "static(1)-静态会话", "2": "dynamic(2)-动态会话", "3": "entireDynamic(3)-全部动态会话", "4": "auto(4)-自动会话"},
	// "hwBfdSessDefaultIp":       {"1": "no", "2": "yes"},
	// "hwBfdSessBindType":        {"1": "interfaceIp(1) -BFD for IP绑定接口和对端IP", "2": "peerIp(2) -BFD for IP仅有对端IP", "3": "sourceIp(3) -BFD for IP绑定对端IP和源IP", "4": "ifAndSourceIp(4) -BFD for IP绑定接口、对端IP、和源IP", "5": "fec(5) -BFD for FEC(当前不支持)", "6": "tunnelIf(6) -BFD for Tunnel interface(当前不支持)", "7": "ospf(7) -BFD for OSPF", "8": "isis(8) -BFD for ISIS", "9": "ldpLsp(9) -BFD for LDP-LSP", "10": "staticLsp(10) -BFD for static LSP", "11": "teLsp(11) -BFD for TE-LSP", "12": "teTunnel(12) -BFD for TE-Tunnel", "13": " pw(13) -BFD for PW", "15": "vsiPw(15) -BFD for VSI PW", "21": "ldpTunnel(21) -BFD for LDP-Tunnel", "22": "bgpTunnel(22) -BFD for BGP-Tunnel"},
	// "hwBfdSessPWSecondaryFlag": {"1": "flagMasterPW(主PW)", "2": "flagSecondaryPW(备PW)", "3": "flagNoPW(没有绑定PW)"},
	// "hwBfdSessDiscrAuto":       {"1": "enabled(标识符可以自动分配)", "2": "disabled(标识符不可以自动分配)"},
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

// oid黑名单匹配
func dropOID(OID, Blacklist string) bool {
	file, err := os.Open(Blacklist)
	if err != nil {
		log.WithField("err", err).Error("打开blackmiblist.txt失败")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		log.WithField("err", err).Error("读取blackmiblist.txt发生错误")
	}

	matchlen := 0
	matchmib := ""

	OID = strings.TrimPrefix(OID, ".")
	OIDS := strings.Split(OID, ".")

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		mib := parts[1][1 : len(parts[1])-1]

		if strings.HasPrefix(OID, mib) {
			mibcount := strings.Count(mib, ".") + 1
			nowid := strings.Join(OIDS[:mibcount], ".")
			if len(nowid) > len(mib) {
				continue
			}
			matchlen = len(mib)
			if matchlen > len(matchmib) {
				matchmib = mib
			}
		}
	}
	return matchlen > 0
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
		oidDesc := ""
		if name, desc, err := global_mib_tree.FindNodeName(v.Name); err != nil {
			log.WithField("err", err).Error("没有找到OID解析")
			oidName = v.Name
		} else {
			oidName = name
			oidDesc = desc
		}
		switch v.Type {
		case g.Integer:
			// 额外解析字段，这里将把ifIndex字段翻译成ifName最后填入ParseValue
			value := v.Value
			parse_value := ""
			parts := strings.Split(oidName, ".")
			// 值映射 value map
			map_v, ok := valueMap[parts[0]]
			if ok {
				value = map_v[fmt.Sprintf("%v", v.Value)]
			}
			if handler, exists := SpeedValueMap[parts[0]]; exists {
				parse_value = handler(fmt.Sprintf("%v", v.Value))
			}
			pdu := TrapPDU{
				OID:        oidName,
				RawOID:     v.Name,
				Type:       v.Type,
				Value:      value,
				Ts:         bjTime.Format("2006-01-02 15:04:05"),
				ParseValue: parse_value,
				Desc:       oidDesc,
			}
			if drop := dropOID(pdu.RawOID, global.GVA_CONFIG.TrapServer.BlackMibMapFile); drop {
				return
			}
			pdus = append(pdus, &pdu)
		case g.OctetString:
			b := v.Value.([]byte)
			parse_value := parseOctetStringToIP(v.Value.([]byte))
			// log.WithField("OID", v.Name).WithField("string", fmt.Sprintf("%s", b)).WithField("Type", v.Type).Info()
			pdu := TrapPDU{
				OID:    oidName,
				RawOID: v.Name,
				Type:   v.Type,
				// Value:  fmt.Sprintf("%s", b),
				Value: string(b),
				// Ts:    time.Now().Format("2006-01-02 15:04:05"),
				Ts:         bjTime.Format("2006-01-02 15:04:05"),
				ParseValue: parse_value,
				Desc:       oidDesc,
			}
			if drop := dropOID(pdu.RawOID, global.GVA_CONFIG.TrapServer.BlackMibMapFile); drop {
				return
			}
			pdus = append(pdus, &pdu)
		// 嵌套OID
		case g.ObjectIdentifier:
			obj_id := fmt.Sprintf("%s", v.Value)
			obj_name := ""
			oidDesc := ""
			parse_value := ""
			if name, desc, err := global_mib_tree.FindNodeName(obj_id); err != nil {
				log.WithField("err", err).Error("trans oid to name error")
				obj_name = obj_id
			} else {
				obj_name = name
				oidDesc = desc
			}
			pdu := TrapPDU{
				OID:    oidName,
				RawOID: obj_id,
				Type:   v.Type,
				Value:  obj_name,
				// Ts:    time.Now().Format("2006-01-02 :04:05"),
				Ts:         bjTime.Format("2006-01-02 15:04:05"),
				ParseValue: parse_value,
				Desc:       oidDesc,
			}
			if drop := dropOID(pdu.RawOID, global.GVA_CONFIG.TrapServer.BlackMibMapFile); drop {
				return
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
				Desc:       oidDesc,
			}
			if drop := dropOID(pdu.RawOID, global.GVA_CONFIG.TrapServer.BlackMibMapFile); drop {
				return
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
	sender.Sends(hostip, push_msg, msg)
}

func paesePdusToListMap(pdus []*TrapPDU) []map[string]string {
	res := make([]map[string]string, 0)
	for _, v := range pdus {
		res = append(res, map[string]string{"oid": v.OID, "type": fmt.Sprintf("%v", v.Type), "value": fmt.Sprintf("%v", v.Value), "ts": v.Ts, "raw_oid": v.RawOID, "parse_value": v.ParseValue, "desc": v.Desc})
	}
	return res
}

func BaseTrapHandler(packet *g.SnmpPacket, addr *net.UDPAddr) {
	log.WithField("addr", addr.IP.String()).Info("got trap package from")
	curList, ok := TrapMap[addr.IP.String()]
	if ok {
		// log.WithField("address", addr.IP.String()).Info("address is exits.")
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
		return fmt.Errorf("%v", fmt.Sprintf("该IP【%s】没有在trap数据库中", ip))
		// return errors.New(fmt.Sprintf("该IP【%s】没有在trap数据库中", ip))
	}
	return nil
}

// func checkItem(data interface{}, pdu interface{}) bool {
// 	d := data.(TrapPDU)
// 	p := pdu.(*TrapPDU)
// 	if d.OID == p.OID && d.Ts == p.Ts && fmt.Sprintf("%v", d.Value) == fmt.Sprintf("%v", p.Value) && fmt.Sprintf("%v", d.Type) == fmt.Sprintf("%v", p.Type) {
// 		return true
// 	} else {
// 		return false
// 	}
// }
// parseAsIP 尝试解析为IP地址
func parseAsIP(data []byte) string {
	// IPv4: 4字节
	if len(data) == 4 {
		ip := net.IPv4(data[0], data[1], data[2], data[3])
		// 验证IP有效性（排除全零等特殊情况）
		if !ip.IsUnspecified() && !ip.IsMulticast() {
			return ip.String()
		}
	}

	// IPv6: 16字节
	if len(data) == 16 {
		ip := net.IP(data)
		if ip.To4() == nil && !ip.IsUnspecified() && !ip.IsMulticast() {
			return ip.String()
		}
	}

	return ""
}

// parseAsMAC 尝试解析为MAC地址
func parseAsMAC(data []byte) string {
	// MAC地址: 6字节
	if len(data) == 6 {
		// 构建MAC地址格式
		hexParts := make([]string, len(data))
		for i, b := range data {
			hexParts[i] = fmt.Sprintf("%02x", b)
		}
		return strings.Join(hexParts, ":")
	}

	return ""
}

// formatAsHexString 将字节数组格式化为十六进制字符串
// func formatAsHexString(data []byte) string {
// 	if len(data) == 0 {
// 		return ""
// 	}

// 	hexParts := make([]string, len(data))
// 	for i, b := range data {
// 		hexParts[i] = fmt.Sprintf("%02x", b)
// 	}
// 	return strings.Join(hexParts, ":")
// }

// 保持向后兼容的函数
func parseOctetStringToIP(data []byte) string {
	return ParseOctetString(data)
}

// ParseOctetString 通用解析函数，自动识别IP和MAC地址，解析失败返回原有内容
func ParseOctetString(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// 尝试解析为IP地址 (IPv4: 4字节, IPv6: 16字节)
	if ip := parseAsIP(data); ip != "" {
		return ip
	}

	// 尝试解析为MAC地址 (6字节)
	if mac := parseAsMAC(data); mac != "" {
		return mac
	}

	// 如果都无法解析，返回原始内容的十六进制表示
	// return formatAsHexString(data)
	return string(data)
}

// 定义处理函数类型（现在接收字符串参数）
type ValueHandler func(string) string

// 全局value map 处理kbps
var SpeedValueMap map[string]ValueHandler = map[string]ValueHandler{
	"hwCurrentStatisticalPeriodRate": convertRateSimple,
	"hwLastStatisticalPeriodRate":    convertRateSimple,
}

// convertRateSimple 简单速率转换（直接向上取整）
func convertRateSimple(kbpsStr string) string {
	// 将字符串转换为整数
	kbps, err := strconv.ParseFloat(kbpsStr, 64)
	if err != nil {
		// 如果转换失败，返回原始字符串
		return kbpsStr
	}

	switch {
	case kbps >= 1000000: // 转换为Gbps
		gbps := kbps / 1000000
		return fmt.Sprintf("%.0fGbps", math.Ceil(gbps))
	case kbps >= 1000: // 转换为Mbps
		mbps := kbps / 1000
		return fmt.Sprintf("%.0fMbps", math.Ceil(mbps))
	default: // 保持kbps
		return fmt.Sprintf("%.0fkbps", math.Ceil(kbps))
	}
}
