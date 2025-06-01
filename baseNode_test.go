package nodes

import "testing"

type MyNode struct {
	BaseNode[string]
}

func TestNode(t *testing.T) {
	newNode := MyNode{}

	newNode.AddNodes(&MyNode{}).
		AddNodes(&MyNode{}).
		AddNodes(&MyNode{}).
		AddNodes(&MyNode{}).
		AddNodes(&MyNode{}).
		Run()
}
