# SNMPTrapServer

#### Description
用于接收snmp trap消息，并将消息发送到prometheus的pushgateway中,然后通过prometheus报警

#### Software Architecture
主要使用了github.com/gosnmp/gosnmp来解析snmp报文

#### Installation

1.  下载代码包
2.  go get 
3.  go build -o snmptrap-server main.go

#### Instructions

1.  收集到的snmp报文主要分三部分:
* OID: 发送到trapserver的oid， ‘.1.3.1.’这种格式的数据
* Value: 发送过来的值
* Type: SNMP数据类型
2.  接受到之后，组装了一个snmp_trap_pdu{host=“发送过来的主机IP” OID=“.x.x.x.x." value="xxxx" type="xxxxx" ts=“接收到消息的时间”}
   如果snmp_trap_pdu接收到的值为1，表示有消息，值为0表示已经恢复


#### Contribution

1.  Fork the repository
2.  Create Feat_xxx branch
3.  Commit your code
4.  Create Pull Request


#### Gitee Feature

1.  You can use Readme\_XXX.md to support different languages, such as Readme\_en.md, Readme\_zh.md
2.  Gitee blog [blog.gitee.com](https://blog.gitee.com)
3.  Explore open source project [https://gitee.com/explore](https://gitee.com/explore)
4.  The most valuable open source project [GVP](https://gitee.com/gvp)
5.  The manual of Gitee [https://gitee.com/help](https://gitee.com/help)
6.  The most popular members  [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
