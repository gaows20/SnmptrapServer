// Copyright 2012 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

package main

import (
	"fmt"
	"strings"

	g "github.com/gosnmp/gosnmp"
)

func runSnmpGet(target, community, oid string) (string, error) {
	// Create a GoSNMP struct with specified target and community
	var res string = ""
	g.Default.Target = target
	g.Default.Community = community
	err := g.Default.Connect()
	if err != nil {
		fmt.Errorf("Connect() err: %v", err)
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

func main() {
	var parseOIDlist map[string]string = map[string]string{"ifIndex": "1.3.6.1.2.1.31.1.1.1.1."}
	parts := strings.Split("ifIndex.47", ".")
	res, ok := parseOIDlist[parts[0]]
	if ok {
		fmt.Println(res, "存在于 parseOIDlist 中")
		// res, err := runSnmpGet("10.254.23.47", global.GVA_CONFIG.TrapServer.Community, oid)
		// if err != nil {
		// 	log.Fatalf("querySnmp() err: %v", err)
		// }
	} else {
		fmt.Println(parts, "不存在于 parseOIDlist 中")
	}
	// target := "1.1.1.1"
	// community := "public"
	// oid := "1.3.6.1.2.1.31.1.1.1.1.47"
	// res, err := querySnmp(target, community, oid)
	// if err != nil {
	// 	log.Fatalf("querySnmp() err: %v", err)
	// }
	// fmt.Println(res)
}
