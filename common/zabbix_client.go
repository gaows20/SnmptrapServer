package main

import (
	"fmt"
	"os"
)

func main() {
	/*
	host := "10.187.248.15"
	port := 10051
	var metrics []*sender.Metric
	metrics = append(metrics, sender.NewMetric(host, "cqrcb.snmptrap", "hello", time.Now().Unix()))

	// Create instance of Packet class
	packet := sender.NewPacket(metrics)

	// Send packet to zabbix
	z := sender.NewSender("10.187.32.8", port)
	z.Send(packet)
	*/
	env := os.Environ()
	procAttr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}
	pid, err := os.StartProcess("zabbix_sender.exe", []string{"zabbix_sender.exe","-z", "10.187.32.8", "-s", "10.187.248.15", "-k", "snmptraper.fallback", "-o", "hello world"}, procAttr)
	if err != nil {
		fmt.Printf("Error %v starting process!", err) //
		os.Exit(1)
	}
	fmt.Printf("The process id is %v", pid)
}
