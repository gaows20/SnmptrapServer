package sender

import (
	"cqrcsnmpserver/global"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

func PushMetrics(host string, msg global.PushMessage, msg_info string) (err error) {
	completionTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "machine_snmp_trap_info",
		Help: "snmp trap info from machine.contains OID, Type,Vale,Timestamp",
	})
	completionTime.SetToCurrentTime()
	completionTime.Set(1) // set可以设置任意值（float64）.BasicAuth("admin", "adgasgbsdfgas")
	pusher := push.New(global.GVA_CONFIG.Sender.PushGatewayUrl, global.GVA_CONFIG.Sender.JobName).Collector(completionTime).Grouping("instance", host)
	if err := pusher.Push(); err != nil {
		return err
	}
	return nil
}

func PushRecoverMetrics(host string) (err error) {
	completionTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "machine_snmp_trap_info",
		Help: "snmp trap info from machine.contains OID, Type,Vale,Timestamp",
	})
	completionTime.SetToCurrentTime()
	completionTime.Set(0) // set可以设置任意值（float64）
	pusher := push.New(global.GVA_CONFIG.Sender.PushGatewayUrl, global.GVA_CONFIG.Sender.JobName).Collector(completionTime).Grouping("instance", host)
	if err := pusher.Push(); err != nil {
		return err
	}
	return nil
}
