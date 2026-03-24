package main

import (
	"cqrcsnmpserver/string_express"

	g "github.com/gosnmp/gosnmp"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Default is a pointer to a GoSNMP struct that contains sensible defaults
	// eg port 161, community public, etc
	g.Default.Target = "10.103.0.116"
	g.Default.Port = 162
	g.Default.Version = g.Version2c
	g.Default.Community = "public"
	g.Default.Logger = g.NewLogger(log.New())
	err := g.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer g.Default.Conn.Close()

	// 华三交换机SNMP陷阱模拟
	// 企业OID: .1.3.6.1.4.1.25506 (H3C)
	// 链路状态变化陷阱
	pdu := g.SnmpPDU{
		Name:  ".1.3.6.1.6.3.1.1.4.1.0",
		Type:  g.ObjectIdentifier,
		Value: ".1.3.6.1.4.1.25506.2.6.1.1.1.0", // H3C链路状态变化陷阱
	}
	
	// 接口索引
	ifIndex := g.SnmpPDU{
		Name:  ".1.3.6.1.2.1.2.2.1.1.1",
		Type:  g.Integer,
		Value: int(1),
	}
	
	// 接口描述
	ifDescr := g.SnmpPDU{
		Name:  ".1.3.6.1.2.1.2.2.1.2.1",
		Type:  g.OctetString,
		Value: "GigabitEthernet1/0/1",
	}
	
	// 接口状态
	ifOperStatus := g.SnmpPDU{
		Name:  ".1.3.6.1.2.1.2.2.1.8.1",
		Type:  g.Integer,
		Value: int(2), // 2表示down
	}

	trap := g.SnmpTrap{
		Variables: []g.SnmpPDU{pdu, ifIndex, ifDescr, ifOperStatus},
	}

	_, err = g.Default.SendTrap(trap)
	if err != nil {
		log.Fatalf("SendTrap() err: %v", err)
	}
	log.Info("SNMP trap sent successfully to 10.103.0.116:162")
}
