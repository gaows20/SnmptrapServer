package main

import (
	"cqrcsnmpserver/string_express"

	g "github.com/gosnmp/gosnmp"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Default is a pointer to a GoSNMP struct that contains sensible defaults
	// eg port 161, community public, etc
	g.Default.Target = "10.254.23.100"
	g.Default.Port = 10162
	g.Default.Version = g.Version2c
	g.Default.Community = "public"
	g.Default.Logger = g.NewLogger(log.New())
	err := g.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer g.Default.Conn.Close()

	pdu := g.SnmpPDU{
		Name:  ".1.3.6.1.6.3.1.1.4.1.0",
		Type:  g.ObjectIdentifier,
		Value: ".1.3.6.1.6.3.1.1.5.1",
	}
	gb2312string, _ := string_express.UTF82GB2312([]byte("你好世界"))
	pdustr := g.SnmpPDU{
		Name:  ".1.3.6.1.6.3.1.1.4.1.0",
		Type:  g.OctetString,
		Value: gb2312string,
	}

	trap := g.SnmpTrap{
		Variables: []g.SnmpPDU{pdu, pdustr},
	}

	_, err = g.Default.SendTrap(trap)
	if err != nil {
		log.Fatalf("SendTrap() err: %v", err)
	}
}
