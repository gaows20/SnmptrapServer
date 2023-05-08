# SNMPTrapServer

#### Description

用于接收 snmp trap 消息，并将消息发送到 prometheus 的 pushgateway 中和 zabbix 中，,然后通过 prometheus 或者 zabbix 报警

#### Software Architecture

主要使用了 github.com/gosnmp/gosnmp 来解析 snmp 报文

#### Installation

1.  下载代码包
2.  go get
3.  go build -o snmptrap-server main.go

#### Instructions

1.  收集到的 snmp 报文主要分三部分:

- OID: 发送到 trapserver 的 oid， ‘.1.3.1.’这种格式的数据
- Value: 发送过来的值
- Type: SNMP 数据类型

2.  接受到之后，组装了一个 snmp_trap_pdu{host=“发送过来的主机 IP” OID=“.x.x.x.x." value="xxxx" type="xxxxx" ts=“接收到消息的时间”}
    如果 snmp_trap_pdu 接收到的值为 1，表示有消息，值为 0 表示已经恢复

#### 启动项目

开发 go run main.go
手动发送 trap 消息 cd client go run test_client.go

#### prometheus

查询语法
machine_snmp_trap_info{job="snmp_trap"}
