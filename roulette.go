package bstkbutler

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/glog"
	"github.com/kr/beanstalk"
)

// BstkRouletteTubeSet set tube names
func (bt *Butler) BstkRouletteTubeSet() {
	bt.rwMu.Lock()
	tubes := []string{}
	for key, _ := range bt.BstkRouletteMap {
		tubes = append(tubes, key)
	}
	if len(tubes) > 0 {
		bt.bstkConn.TubeSet = *beanstalk.NewTubeSet(bt.bstkConn, tubes...)
	}
	bt.rwMu.Unlock()
}

// 获取job到指定方法中去执行
func (bt *Butler) BstkRoulette(restart_bstk_intval, reserve_timeout time.Duration) {
	bt.InitialBstk(restart_bstk_intval) //重试到稳定
	for {
		bt.BstkRouletteTubeSet()                             // set tube names
		id, msg, err := bt.bstkConn.Reserve(reserve_timeout) //获取一个job
		if err != nil {
			if strings.Contains(err.Error(), "reserve-with-timeout: timeout") {
				time.Sleep(restart_bstk_intval)
				continue
			} else {
				glog.Errorln("Err in Reserve Job. ", err.Error())
				bt.InitialBstk(restart_bstk_intval) //重试到稳定
				continue
			}
		}
		jobstat, err := bt.bstkConn.StatsJob(id)
		if err != nil {
			glog.Errorln("Err in StatsJob Job. ", err.Error())
			bt.InitialBstk(restart_bstk_intval) //重试到稳定
			continue
		}
		if err := bt.bstkConn.Delete(id); err != nil {
			glog.Warningln(`Err in delete job : `, err.Error())
			bt.InitialBstk(restart_bstk_intval) //重试到稳定
			continue
		}
		consumerFunc, ok := bt.BstkRouletteMap[jobstat["tube"]]
		if ok {
			go consumerFunc(msg)
		}
	}
}

// BstkRouletteMapRegistry 外部函数注册 允许外部访问
func (bt *Butler) BstkRouletteMapRegistry(registry map[string]func([]byte)) {
	bt.rwMu.Lock()
	for key, value := range registry {
		bt.BstkRouletteMap[key] = value
	}
	bt.rwMu.Unlock()
}

// InitialBstk 初始化链接beanstalk
// 当bstk挂掉时重试，重试间隔可以逐步扩大到稳定
func (bt *Butler) InitialBstk(period time.Duration) {
	bt.mu.Lock()
	stopCh := make(chan struct{})
	bt.JitterUntil(func() {
		if bstkConn, err := bt.factory(); err == nil {
			bt.bstkConn = bstkConn.(*beanstalk.Conn)
			close(stopCh)
		}
	}, period, true, stopCh)
	glog.Infoln(`initial beanstalk successful.`)
	bt.mu.Unlock()
}

// 逐步扩大到稳定
func (bt *Butler) JitterUntil(f func(), period time.Duration, sliding bool, stopCh <-chan struct{}) {
	var t *time.Timer
	times := 1
	factor := 1
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		jitteredPeriod := period
		jitteredPeriod, factor = bt.Jitter(period, times, factor)
		if !sliding {
			t = time.NewTimer(jitteredPeriod)
		}
		f()
		if sliding {
			t = time.NewTimer(jitteredPeriod)
		}
		select {
		case <-stopCh:
			return
		case <-t.C:
			times++
		}
	}
}

// 时间逐步增大到稳定
func (bt *Butler) Jitter(duration time.Duration, times int, factor int) (time.Duration, int) {
	switch {
	case times%2 == 0 && factor <= 30:
		factor++
		break
	case times%4 == 0 && factor <= 40:
		factor++
		break
	case times%10 == 0 && factor <= 50:
		factor++
		break
	case times%100 == 0:
		factor++
	}
	//fmt.Println(int(float64(factor)*math.Atan(float64(times)) + rand.Float64()))
	wait := duration + time.Duration(int(float64(factor)*math.Atan(float64(times))+rand.Float64()))*time.Second
	return wait, factor
}
