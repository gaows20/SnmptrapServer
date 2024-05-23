package mib_tree

import (
	"fmt"
	"testing"
)

func TestAddNode(t *testing.T) {
	mb := NewMibTree()
	mb.LoadFile("./miblist.txt")
	//mb.Print(0)

	if name, desc, err := mb.FindNodeName(".1.3.6.1.2.1.1.3.0"); err == nil {
		fmt.Printf("name:%s\n", name)
		fmt.Printf("desc:%s\n", desc)
	} else {
		fmt.Printf("err is not null:%v\n", err)
	}
}

// func main() {
// 	mb := NewMibTree()
// 	mb.LoadFile("../miblist.txt")
// 	// fmt.Println(black_mib_tree)
// 	if name, desc, err := mb.FindNodeName(".1.3.6.1.2.1.1.3.0"); err == nil {
// 		fmt.Printf("name:%s\n", name)
// 		fmt.Printf("desc:%s\n", desc)
// 	} else {
// 		fmt.Printf("err is not null:%v\n", err)
// 	}
// }
