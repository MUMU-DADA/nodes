package nodes

// InterruptType 定义调用中断类型
type InterruptType int

const (
	// Continue 中断类型：继续执行
	Continue InterruptType = iota
	// Skip 中断类型：跳过执行
	Skip
	// Break 中断类型：中断执行
	Break
)

// NodeWaitType 节点并发执行等待类型
type NodeWaitType int

const (
	// WaitAll 并发等待类型：等待所有并发节点执行完成
	WaitAll NodeWaitType = iota
	// WaitAny 并发等待类型：等待任意一个并发节点执行完成
	WaitAny
)
