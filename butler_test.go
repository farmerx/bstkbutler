package bstkbutler

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

var bstk BSTKButler

func init() {
	bstk = NewButler(InitOptions{
		BstkConn:            `127.0.0.1:11300`,
		InitialCap:          0,                // 连接池中拥有的最小连接数
		MaxCap:              30,               // 连接池中拥有的最大的连接数
		IdleTimeout:         15 * time.Second, // 链接最大空闲时间，超过该事件则将失效
		RestartBstkInterval: 15 * time.Second, // bstk任务轮重启频率
		ReserveTimeout:      15 * time.Second, // bstk任务轮reserve job time
	})
}

func TestNewButler(t *testing.T) {
	type args struct {
		options InitOptions
	}
	xxx := InitOptions{
		BstkConn:            `127.0.0.1:11300`,
		InitialCap:          0,                // 连接池中拥有的最小连接数
		MaxCap:              30,               // 连接池中拥有的最大的连接数
		IdleTimeout:         15 * time.Second, // 链接最大空闲时间，超过该事件则将失效
		RestartBstkInterval: 15 * time.Second, // bstk任务轮重启频率
		ReserveTimeout:      15 * time.Second, // bstk任务轮reserve job time
	}
	tests := []struct {
		name string
		args args
		want BSTKButler
	}{
		// TODO: Add test cases.
		{"bibinbin", args{xxx}, NewButler(xxx)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewButler(tt.args.options); !reflect.DeepEqual(got, got) {
				t.Errorf("NewButler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestButler_PutJob(t *testing.T) {
	if err := bstk.PutJob(`test`, `hello world!`); err != nil {
		t.Error(err)
	}
	_, body, err := bstk.ReserveJobByTube(`test`, time.Second)
	if err != nil {
		t.Error(err)
	}
	if string(body) != `hello world!` {
		t.Errorf("reserve = %v, want %v", string(body), `hello world!`)
	}
}

func TestButler_BuryJob(t *testing.T) {
	if err := bstk.PutJob(`test`, `hello world!`); err != nil {
		t.Error(err)
	}
	id, _, err := bstk.ReserveJobByTube(`test`, time.Second)
	if err != nil {
		t.Error(err)
	}
	s, err := bstk.GetStatsJobById(id)
	if err != nil {
		t.Error(err)
	}
	pri, err := strconv.Atoi(s["pri"])
	if err != nil {
		t.Error(err)
	}
	if err := bstk.BuryJob(id, uint32(pri)); err != nil {
		t.Error(err)
	}
}

func TestButler_KickJob(t *testing.T) {
	stats, err := bstk.GetStatsForTube(`test`)
	if err != nil {
		t.Error(err)
	}
	if stats.JobsBuried == `0` {
		return
	}
	bound, err := strconv.Atoi(stats.JobsBuried)
	if err != nil {
		t.Error(err)
	}
	if _, err := bstk.KickJob(`test`, bound); err != nil {
		t.Error(err)
	}
}

func TestButler_Touch(t *testing.T) {
	if err := bstk.PutJob(`test`, `hello world!`); err != nil {
		t.Error(err)
	}
	id, _, err := bstk.ReserveJobByTube(`test`, time.Second)
	if err != nil {
		t.Error(err)
	}
	if err := bstk.Touch(id); err != nil {
		t.Error(err)
	}
}

func TestButler_Peek(t *testing.T) {
	if err := bstk.PutJob(`test`, `hello world!`); err != nil {
		t.Error(err)
	}
	id, _, err := bstk.ReserveJobByTube(`test`, time.Second)
	if err != nil {
		t.Error(err)
	}
	if _, err := bstk.Peek(id); err != nil {
		t.Error(err)
	}
}

func TestButler_ReserveJobByTube(t *testing.T) {
	if err := bstk.PutJob(`test`, `hello world!`); err != nil {
		t.Error(err)
	}
	_, _, err := bstk.ReserveJobByTube(`test`, time.Second)
	if err != nil {
		t.Error(err)
	}
}

func TestButler_GetTubesName(t *testing.T) {
	if _, err := bstk.GetTubesName(); err != nil {
		t.Error(err)
	}
}

func TestButler_DeleteJobByJobstats(t *testing.T) {
	if err := bstk.PutJob(`test`, `hello world!`); err != nil {
		t.Error(err)
	}
	id, _, err := bstk.ReserveJobByTube(`test`, time.Second)
	if err != nil {
		t.Error(err)
	}
	s, err := bstk.GetStatsJobById(id)
	if err != nil {
		t.Error(err)
	}
	pri, err := strconv.Atoi(s["pri"])
	if err != nil {
		t.Error(err)
	}
	if err := bstk.BuryJob(id, uint32(pri)); err != nil {
		t.Error(err)
	}
	if err := bstk.DeleteJobByJobstats(`test`, "buried"); err != nil {
		t.Error(err)
	}
}

func TestButler_GetStatsForTube(t *testing.T) {
	if _, err := bstk.GetStatsForTube(`test`); err != nil {
		t.Error(err)
	}
}
