package linklist

import (
	"errors"
	"fmt"
)

type Object interface {
}

type Node struct {
	Data Object //定义数据域
	Next *Node  //定义地址域，指向下一个地址
}

type List struct {
	headNode *Node
	length   int64
}

func NewList() *List {
	return &List{
		headNode: nil,
		length:   0,
	}
}

//判断链表是否为空
func (l *List) IsEmpty() bool {
	if l.headNode == nil {
		return true
	} else {
		return false
	}
}

//获取链表长度
func (l *List) Length() int64 {
	return l.length
}

// 从链表头部添加元素
func (l *List) Add(data Object) *Node {
	node := &Node{Data: data}
	node.Next = l.headNode
	l.headNode = node
	l.length++
	return node
}

// 从链表尾部添加元素

func (l *List) Append(data Object) {
	node := &Node{Data: data}
	if l.IsEmpty() {
		l.headNode = node
		l.length++
	} else {
		cur := l.headNode
		for cur.Next != nil {
			cur = cur.Next
		}
		cur.Next = node
		l.length++
	}
}

//删除指定位置的元素
func (l *List) RemoveAtIndex(index int64) (err error) {
	pre := l.headNode
	if index <= 0 { //如果index为0或者小于0，则删除头节点
		l.headNode = pre.Next
		l.length--
	} else if index > l.Length() {
		return errors.New(fmt.Sprintf("超出链表长度,indext[%v], list length[%v]", index, l.Length()))
	} else {
		count := int64(0) //定义计数器
		for count != (index-1) && pre.Next != nil {
			count++
			pre = pre.Next
		}
		pre.Next = pre.Next.Next
		l.length--
	}
	return nil
}

//返回链表所有元素的切片（逆序，最新的消息在前面）
func (l *List) GetListArray() (arr []interface{}, err error) {
	arr = make([]interface{}, 0, 10)
	if !l.IsEmpty() {
		// 先收集所有元素
		var temp []interface{}
		cur := l.headNode
		for {
			temp = append(temp, cur.Data)
			if cur.Next != nil {
				cur = cur.Next
			} else {
				break
			}
		}
		// 反转切片，使最新的消息在前面
		for i := len(temp) - 1; i >= 0; i-- {
			arr = append(arr, temp[i])
		}
	}
	return arr, nil
}

func (l *List) RemoveItem(data interface{}, checkdata func(listdata interface{}, checkdata interface{}) bool) {
	pre := l.headNode
	if checkdata(pre.Data, data) {
		l.headNode = pre.Next
		l.length--
	} else {
		for pre.Next != nil {
			if checkdata(pre.Next.Data, data) {
				pre.Next = pre.Next.Next
				l.length--
			} else {
				pre = pre.Next
			}
		}
	}
}
