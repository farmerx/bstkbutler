package bstkbutler

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kr/beanstalk"
	"github.com/tcppool"
)

func TestButler_BstkRouletteMapRegistry(t *testing.T) {
	bstk.PutJob(`test`, `hello world`)
	bstk.PutJob(`test2`, `hello world`)
	// 任务轮注册
	bstk.BstkRouletteMapRegistry(map[string]func([]byte){
		"test": func(data []byte) {
			fmt.Println(string(data))
		},
		"test2": func(data []byte) {
			fmt.Println(string(data))
		},
		"hello": func(data []byte) {
			fmt.Println(string(data))
		},
	})
	time.Sleep(2 * time.Second)
}

func TestButler_JitterUntil(t *testing.T) {
	type fields struct {
		factory         func() (interface{}, error)
		close           func(interface{}) error
		pool            tcppool.TCPPool
		BstkRouletteMap map[string]func([]byte)
		mu              sync.Mutex
		rwMu            sync.RWMutex
		bstkConn        *beanstalk.Conn
	}
	type args struct {
		f       func()
		period  time.Duration
		sliding bool
		stopCh  <-chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt := &Butler{
				factory:         tt.fields.factory,
				close:           tt.fields.close,
				pool:            tt.fields.pool,
				BstkRouletteMap: tt.fields.BstkRouletteMap,
				mu:              tt.fields.mu,
				rwMu:            tt.fields.rwMu,
				bstkConn:        tt.fields.bstkConn,
			}
			bt.JitterUntil(tt.args.f, tt.args.period, tt.args.sliding, tt.args.stopCh)
		})
	}
}

func TestButler_Jitter(t *testing.T) {
	type fields struct {
		factory         func() (interface{}, error)
		close           func(interface{}) error
		pool            tcppool.TCPPool
		BstkRouletteMap map[string]func([]byte)
		mu              sync.Mutex
		rwMu            sync.RWMutex
		bstkConn        *beanstalk.Conn
	}
	type args struct {
		duration time.Duration
		times    int
		factor   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Duration
		want1  int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt := &Butler{
				factory:         tt.fields.factory,
				close:           tt.fields.close,
				pool:            tt.fields.pool,
				BstkRouletteMap: tt.fields.BstkRouletteMap,
				mu:              tt.fields.mu,
				rwMu:            tt.fields.rwMu,
				bstkConn:        tt.fields.bstkConn,
			}
			got, got1 := bt.Jitter(tt.args.duration, tt.args.times, tt.args.factor)
			if got != tt.want {
				t.Errorf("Butler.Jitter() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Butler.Jitter() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
