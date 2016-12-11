package bstkbutler

import (
	"errors"
	"sync"
	"time"

	beanstalk "github.com/kr/beanstalk"
	"github.com/tcppool"
)

// InitOptions ...
type InitOptions struct {
	BstkConn            string
	InitialCap          int           // 连接池中拥有的最小连接数
	MaxCap              int           // 连接池中拥有的最大的连接数
	IdleTimeout         time.Duration // 链接最大空闲时间，超过该事件则将失效
	RestartBstkInterval time.Duration // bstk任务轮重启频率
	ReserveTimeout      time.Duration // bstk任务轮reserve job time
}

// Butler ...
type Butler struct {
	factory         func() (interface{}, error)
	close           func(interface{}) error
	pool            tcppool.TCPPool
	BstkRouletteMap map[string]func([]byte) //bstk的tube和对应消费函数的map
	mu              sync.Mutex
	rwMu            sync.RWMutex
	bstkConn        *beanstalk.Conn
}

// NewButler 初始化bstk管家
func NewButler(options InitOptions) BSTKButler {
	bt := new(Butler)
	// factory 创建连接的方法
	bt.factory = func() (interface{}, error) { return beanstalk.Dial(`tcp`, options.BstkConn) }
	// close 关闭链接的方法
	bt.close = func(c interface{}) error { return c.(*beanstalk.Conn).Close() }
	// 初始化一个beanstalk 连接池
	bt.pool, _ = tcppool.NewChannelPool(tcppool.InitOptions{
		InitialCap:  options.InitialCap,
		MaxCap:      options.MaxCap,
		Factory:     bt.factory,
		Close:       bt.close,
		IdleTimeout: options.IdleTimeout, // 链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
	})
	// 初始化bstk轮子
	bt.BstkRouletteMap = make(map[string]func([]byte))                      // bstk的tube和对应消费函数的map
	go bt.BstkRoulette(options.RestartBstkInterval, options.ReserveTimeout) // 启动任务轮
	return bt
}

// PutJob put job to bstk
func (bt *Butler) PutJob(tube string, msg string) error {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return err
	}
	Tube := beanstalk.Tube{bstkConn.(*beanstalk.Conn), tube}
	if _, err = Tube.Put([]byte(msg), 0, 0, 0); err != nil {
		bt.pool.Close(bstkConn)
		return err
	}
	bt.pool.Put(bstkConn)
	return nil
}

//Bury bury job by id
func (bt *Butler) BuryJob(id uint64, pri uint32) error {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return err
	}
	if err := bstkConn.(*beanstalk.Conn).Bury(id, pri); err != nil {
		bt.pool.Close(bstkConn)
		return err
	}
	bt.pool.Put(bstkConn)
	return nil
}

//  KickJob kick job(可以指定条数)
func (bt *Butler) KickJob(tube string, bound int) (n int, err error) {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return 0, err
	}
	Tube := &beanstalk.Tube{bstkConn.(*beanstalk.Conn), tube}
	if n, err = Tube.Kick(bound); err != nil {
		bt.pool.Close(bstkConn)
		return 0, err
	}
	bt.pool.Put(bstkConn)
	return n, err

}

// touch  job id
func (bt *Butler) Touch(id uint64) error {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return err
	}
	if err := bstkConn.(*beanstalk.Conn).Touch(id); err != nil {
		bt.pool.Close(bstkConn)
		return err
	}
	bt.pool.Put(bstkConn)
	return nil
}

// peek job id
func (bt *Butler) Peek(id uint64) (body []byte, err error) {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return nil, err
	}
	if body, err = bstkConn.(*beanstalk.Conn).Peek(id); err != nil {
		bt.pool.Close(bstkConn)
		return nil, err
	}
	bt.pool.Put(bstkConn)
	return body, nil

}

// ReserveJobByTube reserve job by tube name
func (bt *Butler) ReserveJobByTube(tube string, timeout time.Duration) (id uint64, body []byte, err error) {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return 0, nil, err
	}
	bstkConn.(*beanstalk.Conn).TubeSet = *beanstalk.NewTubeSet(bstkConn.(*beanstalk.Conn), []string{tube}...)
	if id, body, err = bstkConn.(*beanstalk.Conn).Reserve(timeout); err != nil {
		bt.pool.Close(bstkConn)
		return 0, nil, err
	}
	bt.pool.Put(bstkConn)
	return id, body, nil
}

// GetTubesName get tube list
func (bt *Butler) GetTubesName() (tubes []string, err error) {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return nil, err
	}
	if tubes, err = bstkConn.(*beanstalk.Conn).ListTubes(); err != nil {
		bt.pool.Close(bstkConn)
		return nil, err
	}
	bt.pool.Put(bstkConn)
	return tubes, err
}

// DeleteJobByJobstats delete job By job stats
func (bt *Butler) DeleteJobByJobstats(tube, stats string) (err error) {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return err
	}
	var id uint64
	Tube := &beanstalk.Tube{bstkConn.(*beanstalk.Conn), tube}
	switch stats {
	case "buried":
		id, _, err = Tube.PeekBuried()
	case "ready":
		id, _, err = Tube.PeekReady()
	case "delayed":
		id, _, err = Tube.PeekDelayed()
	}
	if err != nil {
		bt.pool.Close(bstkConn)
		return err
	}
	if err = bstkConn.(*beanstalk.Conn).Delete(id); err != nil {
		bt.pool.Close(bstkConn)
		return err
	}
	bt.pool.Put(bstkConn)
	return nil
}

// 获取tube的状态
func (bt *Butler) GetStatsForTube(tube string) (*TubeStats, error) {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return nil, err
	}
	Tube := &beanstalk.Tube{bstkConn.(*beanstalk.Conn), tube}
	stats, err := Tube.Stats()
	if err != nil {
		bt.pool.Close(bstkConn)
		return nil, err
	}
	bt.pool.Put(bstkConn)
	if name, ok := stats["name"]; !ok || name != tube {
		return nil, errors.New("Unable to retrieve tube stats")
	}
	return &TubeStats{
		JobsBuried:   stats["current-jobs-buried"],
		JobsReady:    stats["current-jobs-ready"],
		JobsDelayed:  stats["current-jobs-delayed"],
		JobsReserved: stats["current-jobs-reserved"],
		JobsUrgent:   stats["current-jobs-urgent"],
		Waiting:      stats["current-waiting"],
		TotalJobs:    stats["total-jobs"],
	}, nil
}

// get job stats by id
func (bt *Butler) GetStatsJobById(id uint64) (map[string]string, error) {
	bstkConn, err := bt.pool.Get()
	if err != nil {
		return nil, err
	}
	jobstat, err := bstkConn.(*beanstalk.Conn).StatsJob(id)
	if err != nil {
		bt.pool.Close(bstkConn)
		return nil, err
	}
	bt.pool.Put(bstkConn)
	return jobstat, nil
}
