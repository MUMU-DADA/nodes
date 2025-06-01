package microapp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type BaseNode[T any] struct {
	Children []Node[T]
	IsLoop   bool
}

// SetLoop 是否为循环节点
func (b *BaseNode[T]) SetLoop() {
	b.IsLoop = true
}

// GetNodes 获取所有子节点
func (b *BaseNode[T]) GetNodes() []Node[T] {
	if len(b.Children) == 0 {
		b.Children = []Node[T]{}
	}
	return b.Children
}

// GetAllChildren 获取所有子节点
func (b *BaseNode[T]) GetAllChildren() []Node[T] {
	if len(b.Children) == 0 {
		b.Children = []Node[T]{}
	}
	return b.Children
}

// 并行执行所有子节点并等待所有子节点完成
func (b *BaseNode[T]) execAndWait(ctx *NodeContext[T], rootNode Node[T]) ([]*NodeData[T], error) {
	nodes := rootNode.GetAllChildren()
	// 并发等待所有节点执行完成
	returnLock := sync.Mutex{}              // 各节点返回数据的锁
	nodeDataList := make([]*NodeData[T], 0) // 节点返回数据
	var runErr error                        // 运行错误

	// 并发完成条件有两个 一个是超时，一个是所有节点执行完成
	timeout, cancelFunc := context.WithTimeout(context.Background(), rootNode.GetConcurrentMaxWaitTime()) // 超时时间
	wg := sync.WaitGroup{}                                                                                // 并发等待组
	wg.Add(len(nodes))                                                                                    //  添加节点数量
	waitChan := make(chan struct{})

	// 执行
	go func() {
		defer cancelFunc()
		for _, node := range nodes {
			// 添加等待函数
			waitRunFunc := func(wait Node[T]) {
				defer wg.Done()
				ans, err := wait.Execute(ctx)
				// 有结果后置入结果
				returnLock.Lock()
				defer returnLock.Unlock()
				if err != nil {
					// 节点执行错误
					runErr = err
				} else {
					// 节点执行成功
					nodeDataList = append(nodeDataList, ans)
				}
			}
			go waitRunFunc(node) // 并发执行节点
		}
		wg.Wait() // 等待所有并发执行完成
		waitChan <- struct{}{}
	}()

	// 等待执行完成条件
	select {
	case <-timeout.Done():
		//  超时
		runErr = fmt.Errorf("node %s wait timeout", rootNode.GetName())
	case <-waitChan:
		// 等待完成
	}

	// 检查错误
	if runErr != nil {
		return nil, runErr
	}

	return nodeDataList, nil
}

// 并行执行所有子节点并在任意一个完成之后继续
func (b *BaseNode[T]) execAndContinue(ctx *NodeContext[T], rootNode Node[T]) (*NodeData[T], error) {
	nodes := rootNode.GetAllChildren()
	//  并发等待任意一个节点执行完成
	returnLock := sync.Mutex{} // 各节点返回数据的锁
	var nodeData *NodeData[T]  // 节点返回的数据
	var runErr error           // 执行错误
	var nodeDone bool          // 节点是否执行完成

	// 并发完成条件有两个 一个是超时，一个是任意节点执行完成
	timeout, cancelFunc := context.WithTimeout(context.Background(), rootNode.GetConcurrentMaxWaitTime()) // 超时时间
	done := make(chan struct{})

	// 执行
	go func() {
		defer cancelFunc()
		for _, node := range nodes {
			// 添加等待函数
			waitRunFunc := func(wait Node[T]) {
				defer func() {
					returnLock.Lock()
					if !nodeDone { // 如果节点已经执行完成，则不再执行
						nodeDone = true    //  标记节点已经执行完成
						done <- struct{}{} // 通知执行完成
					}
					returnLock.Unlock()
				}()
				ans, err := wait.Execute(ctx)
				// 有结果后置入结果
				returnLock.Lock()
				defer returnLock.Unlock()
				if !nodeDone { // 如果节点已经执行完成，则不再执行
					if err != nil {
						// 节点执行错误
						runErr = err
					} else {
						// 节点执行成功
						nodeData = ans
					}
					nodeDone = true    //  标记节点执行完成
					done <- struct{}{} //  通知执行完成
				}
			}
			go waitRunFunc(node) // 并发执行节点
		}
	}()

	// 等待执行完成条件
	select {
	case <-timeout.Done():
		//  超时
		runErr = fmt.Errorf("node %s wait timeout", rootNode.GetName())
	case <-done:
		// 等待完成
	}

	// 检查错误
	if runErr != nil {
		return nil, runErr
	}
	return nodeData, nil
}

// GetName 获取节点名称
func (b *BaseNode[T]) GetName() string {
	// 默认值是 BaseNode
	return "BaseNode"
}

// GetConcurrentType 获取 并行类型
func (b *BaseNode[T]) GetConcurrentType() NodeWaitType {
	// 默认值是 等待所有
	return WaitAll
}

// GetConcurrentMaxWaitTime 获取 并行等待时间
func (b *BaseNode[T]) GetConcurrentMaxWaitTime() time.Duration {
	//  默认值是 10 分钟
	return 10 * time.Minute
}

// Init 初始化
func (b *BaseNode[T]) Init() error {
	return nil
}

// Prepare 准备
func (b *BaseNode[T]) Prepare(ctx *NodeContext[T]) (InterruptType, error) {
	// 空节点的话不做任何处理
	return Continue, nil
}

// Execute 执行
func (b *BaseNode[T]) Execute(ctx *NodeContext[T]) (*NodeData[T], error) {
	// 空节点的话不做任何处理
	return ReturnNodeData[T](b, nil), nil
}

// AddNodes 添加节点
func (b *BaseNode[T]) AddNodes(nodes ...Node[T]) Node[T] {
	b.Children = append(b.Children, nodes...)
	return b
}

func (b *BaseNode[T]) baseRun(ctx *NodeContext[T]) (*NodeContext[T], error) {
	for _, v := range b.Children {
		// 判断本节点是否为并发节点
		vChildren := v.GetAllChildren()
		if len(vChildren) > 1 {
			// 子节点下存在多个节点,视为并发执行
			// 获取并发类型
			nodeWaitType := v.GetConcurrentType()
			switch nodeWaitType {
			case WaitAll:
			case WaitAny:
			default:
				return nil, fmt.Errorf("concurrent type error, node name: %s, concurrent type: %d", v.GetName(), nodeWaitType)
			}

			// 子节点中任意前提准备出现了终止或者跳过均执行对应逻辑, 优先终止
			var childContinue bool
			var childBreak bool
			for _, v2 := range vChildren {
				interruptType, err := v2.Prepare(ctx)
				if err != nil {
					return nil, err
				}
				if interruptType == Break {
					childBreak = true
					break
				}
				if interruptType == Continue {
					childContinue = true
				}
			}

			if childBreak {
				// 中断执行
				return nil, nil
			}
			if childContinue {
				// 继续执行下一个处理步骤
				continue
			}

			if nodeWaitType == WaitAny {
				data, err := b.execAndContinue(ctx, v)
				if err != nil {
					return nil, err
				}
				ctx.Set(data)
			}

			if nodeWaitType == WaitAll {
				data, err := b.execAndWait(ctx, v)
				if err != nil {
					return nil, err
				}
				for _, v := range data {
					ctx.Set(v)
				}
			}
		} else {
			// 单节点
			// 单节点准备
			interruptType, err := v.Prepare(ctx)
			if err != nil {
				return nil, err
			}

			// 根据中断类型处理
			switch interruptType {
			case Continue:
				// 继续执行下一个处理步骤
			case Skip:
				// 跳过执行下一个处理步骤
				continue
			case Break:
				// 中断执行
				return nil, nil
			}

			// 执行处理步骤
			nodeData, err := v.Execute(ctx)
			if err != nil {
				return nil, err
			}

			ctx.Set(nodeData)
		}
	}
	return ctx, nil
}

// Run 执行
func (b *BaseNode[T]) Run() (*NodeContext[T], error) {
	// 首先执行本届点的Init
	if err := b.Init(); err != nil {
		return nil, err
	}

	// 执行子节点的Init
	if len(b.Children) > 0 {
		for _, v := range b.Children {
			// 递归执行子节点
			if err := v.Init(); err != nil {
				return nil, err
			}
		}
	}

	if b.IsLoop {
		// 循环执行
		for {
			ctx := NewContext[T]()
			return b.baseRun(ctx)
		}
	} else {
		// 执行节点
		ctx := NewContext[T]()
		return b.baseRun(ctx)
	}
}
