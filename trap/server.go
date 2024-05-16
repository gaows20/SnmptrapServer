package trap

import (
	"cqrcsnmpserver/global"
	"time"

	g "github.com/gosnmp/gosnmp"
	log "github.com/sirupsen/logrus"
)

type TrapPDU struct {
	Id         int64       `json:"id"`
	OID        string      `json:"oid"`
	Type       interface{} `json:"type"`
	Value      interface{} `json:"value"`
	Ts         string      `json:"ts"`
	RawOID     string      `json:"raw_oid"`
	ParseValue string      `json:"parse_value"`
	Desc       string      `json:"desc"`
}

type TrapServer struct {
	listener *g.TrapListener
	ip       string
	port     string
	trapmap  map[string]*TrapPDU
}

func NewTrapServer(ip, port string) (*TrapServer, error) {
	// 关闭gosnmp的debug输出
	//g.Default.Logger = gosnmp.NewLogger(log.New(os.Stdout, "", 0))
	// load mib map file in mibtree
	if err := global_mib_tree.LoadFile(global.GVA_CONFIG.TrapServer.MibMapFile); err != nil {
		return nil, err
	}
	// if tmp_black, err := readBlacklist(global.GVA_CONFIG.TrapServer.BlackMibMapFile); err != nil {
	// 	return nil, err
	// } else {
	// 	black_mib_tree = tmp_black
	// }
	tl := g.NewTrapListener()
	tl.OnNewTrap = BaseTrapHandler
	var version g.SnmpVersion = g.Version2c
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat:           "2006-01-02 15:04:05",
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		// FullTimestamp:true,
		// DisableLevelTruncation:true,
	})
	switch global.GVA_CONFIG.TrapServer.Version { //初始化配置文件的Level
	case "v1":
		version = g.Version1
	case "v2c":
		version = g.Version2c
	case "v3":
		version = g.Version3
	}
	SNMPCONFG := &g.GoSNMP{
		Port:               uint16(global.GVA_CONFIG.TrapServer.Port),
		Transport:          "udp",
		Community:          global.GVA_CONFIG.TrapServer.Community,
		Version:            version,
		Timeout:            time.Duration(global.GVA_CONFIG.TrapServer.Timeout) * time.Second,
		Retries:            3,
		ExponentialTimeout: true,
		MaxOids:            g.MaxOids,
	}
	tl.Params = SNMPCONFG
	tl.Params.Logger = g.NewLogger(log.New())
	return &TrapServer{listener: tl,
		ip:      ip,
		port:    port,
		trapmap: make(map[string]*TrapPDU)}, nil
}

func (t *TrapServer) Run() (err error) {
	listenaddr := t.ip + ":" + t.port
	log.WithField("address", listenaddr).Info("set trapserver address ")
	return t.listener.Listen(listenaddr)
}
