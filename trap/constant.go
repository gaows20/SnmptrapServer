package trap

import (
	"cqrcsnmpserver/linklist"
	"cqrcsnmpserver/mib_tree"
)

var (
	TrapMap         map[string]*linklist.List = make(map[string]*linklist.List)
	global_mib_tree                           = mib_tree.NewMibTree()
	// black_mib_tree  map[string]string         = make(map[string]string)
)

/*
type trapPdu struct {
	OID string
	Value interface{}
	Type interface{}
}
*/
