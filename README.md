# Bstkbutler

* bstk butler beanstalk管家
* 支持获取当前bstk队列状态,且能指定具体的tube名称进行查询
* 支持在bstk挂掉时重试，重试间隔可以逐步扩大到稳定
* 支持创建tube,并在tube中投放消息
* 支持...

## Install and Usage


Import it with:

```go
import "github.com/bibinbin/bstkbutler"
```
and use `pool` as the package name inside the code.

```go
import (
  bstk "github.com/bibinbin/bstkbutler"
)
```

## Example
```vim
package main

import (
	"time"

	"fmt"

	"github.com/bibinbin/bstkbutler"
)

func main() {

	bstk := bstkbutler.NewButler(bstkbutler.InitOptions{
		BstkConn:            `127.0.0.1:11300`,
		InitialCap:          5,                // 连接池中拥有的最小连接数
		MaxCap:              30,               // 连接池中拥有的最大的连接数
		IdleTimeout:         15 * time.Second, // 链接最大空闲时间，超过该事件则将失效
		RestartBstkInterval: 15 * time.Second, // bstk任务轮重启频率
		ReserveTimeout:      15 * time.Second, // bstk任务轮reserve job time
	})
	// put job
	bstk.PutJob(`test`, `hello world`)
	// 任务轮注册
	bstk.BstkRouletteMapRegistry(map[string]func([]byte){
		"test": func(data []byte) {
			fmt.Println(string(data))
		},
		"test2": func(data []byte) {
			fmt.Println(string(data))
		},
	})
	// 。。。

}

```

