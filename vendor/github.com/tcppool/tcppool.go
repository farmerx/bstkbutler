//一个优秀的连接池要能实现对池子的大小控制，线程取用安全，简单等。

package tcppool

import "errors"

var (
	// ErrClosed 连接池已经关闭Error
	ErrClosed = errors.New("pool is closed")
)

// TCPPool 基本方法
type TCPPool interface {
	Get() (interface{}, error) // 获取tcppool
	Put(interface{}) error     // 送回tcppool
	Close(interface{}) error   // 关闭conn
	Release()                  // 关闭tcpool 释放掉所有资源
	Len() int                  // 获取tcppool 池水深度
}
