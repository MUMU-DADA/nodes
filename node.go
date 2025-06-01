package nodes

import "time"

// NodeData 定义节点数据结构
type NodeData[T any] struct {
	GetData func() T
	Name    string
}

// ReturnNodeData 返回节点数据
func ReturnNodeData[T any](node Node[T], getData func() T) *NodeData[T] {
	return &NodeData[T]{
		GetData: getData,
		Name:    node.GetName(),
	}
}

// Node 定义节点接口
type Node[T any] interface {
	GetName() string                         // 获取节点名称
	SetLoop()                                // 节点是否循环执行
	GetConcurrentType() NodeWaitType         // 获取本节点并发执行等待类型
	GetConcurrentMaxWaitTime() time.Duration // 获取并发执行最大等待时间

	AddNodes(nodes ...Node[T]) Node[T] // 添加子节点
	GetAllChildren() []Node[T]         //  获取所有子节点

	Init() error                                        // 初始化处理步骤
	Prepare(ctx *NodeContext[T]) (InterruptType, error) // 准备处理步骤
	Execute(ctx *NodeContext[T]) (*NodeData[T], error)  // 执行处理步骤
	Run() (*NodeContext[T], error)                      // 运行处理步骤
}
