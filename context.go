package nodes

// NodeContext 定义上下文结构
type NodeContext[T any] struct {
	FirstNodeData *NodeData[T]            // 第一个节点数据
	nodeData      map[string]*NodeData[T] // 节点数据
}

// NewContext 新建一个上下文
func NewContext[T any]() *NodeContext[T] {
	return &NodeContext[T]{
		nodeData: make(map[string]*NodeData[T]),
	}
}

func (c *NodeContext[T]) SetContext(ctx *NodeContext[T]) {
	for _, v := range ctx.nodeData {
		c.Set(v)
	}
}

// Set 设置上下文数据
func (c *NodeContext[T]) Set(value *NodeData[T]) {
	if c.FirstNodeData == nil {
		c.FirstNodeData = value
	}
	c.nodeData[value.Name] = value
}

// Get 获取上下文数据
func (c *NodeContext[T]) Get(key string) T {
	return c.nodeData[key].GetData()
}
