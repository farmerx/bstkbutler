// bstk butler beanstalk管家
// 支持获取当前bstk队列状态,且能指定具体的tube名称进行查询
// 支持在bstk挂掉时重试，重试间隔可以逐步扩大到稳定
// 支持创建tube,并在tube中投放消息
// 支持清空某个tube
//
// printStats 打印全部的beanstalk的状态
// +---------+--------+---------+-------+----------+--------+---------+-------+
// |  NAME   | BURIED | DELAYED | READY | RESERVED | URGENT | WAITING | TOTAL |
// +---------+--------+---------+-------+----------+--------+---------+-------+
// | default |      0 |       0 |     0 |        0 |      0 |       0 |     0 |
// | test    |      0 |       0 |     1 |        0 |      1 |       0 |     1 |
// +---------+--------+---------+-------+----------+--------+---------+-------+
//
//  外部函数注册
//  bstk.PutJobToBstk(`test`, `hello world!`)
// 	bstk.BstkRouletteMapRegistry(map[string]func([]byte){
// 		"test": func(data []byte) {
// 			fmt.Println(string(data))
// 		},
// 	})
// 	go bstk.BstkRoulette(10 * time.Second)
// 	time.Sleep(1 * time.Second)
// 	bstk.PrintStats()
//  重试时间逐步递增:
//  [1,2,2,4,4,5,6,7,7,9,9,10,12,14,13,15,16,17,17,18,......30..........31,.........32...........33......]

package bstkbutler

import "time"

// BSTKButler  必须实现的方法
type BSTKButler interface {
	PutJob(tube string, msg string) error
	BuryJob(id uint64, pri uint32) error
	KickJob(tube string, bound int) (n int, err error)
	Touch(id uint64) error
	Peek(id uint64) (body []byte, err error)
	ReserveJobByTube(tube string, timeout time.Duration) (id uint64, body []byte, err error)
	GetTubesName() ([]string, error)
	DeleteJobByJobstats(tube, stats string) (err error)
	GetStatsForTube(tube string) (*TubeStats, error)
	GetStatsJobById(id uint64) (map[string]string, error)
	BstkRoulette(restart_bstk_intval, reserve_timeout time.Duration)
	BstkRouletteMapRegistry(registry map[string]func([]byte))
	InitialBstk(period time.Duration)
}
