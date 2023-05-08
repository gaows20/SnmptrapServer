package mib_tree

import (
	"fmt"
	"testing"
)

func TestAddNode(t *testing.T) {

	mb := NewMibTree()
	mb.LoadFile("./miblist.txt")
	//mb.Print(0)

	if name, err := mb.FindNodeName(".1.3.6.1.2.1.1.3.0"); err == nil {
		fmt.Printf("name:%s\n",name)
	} else {
		fmt.Printf("err is not null:%v\n",err)
	}

}