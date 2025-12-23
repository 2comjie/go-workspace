package server

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type BuildingState int

const (
	BuildingIdle         BuildingState = iota // 空闲，可建造
	BuildingConstructing                      // 建造中
	BuildingFinished                          // 已完成
)

type Building struct {
	ID        int
	State     BuildingState
	StartTime int64 // Unix 秒，建造开始时间
	EndTime   int64 // Unix 秒，建造完成时间
}

// 刷新状态，根据当前时间判断建筑是否完成，返回最新的状态

func (b *Building) Refresh(now int64) BuildingState {
	if b.State != BuildingConstructing {
		return b.State
	}
	if now > b.EndTime {
		b.State = BuildingFinished
	}
	return b.State
}

// 提升建造速度倍率（>1 表示加速），动态调整剩余建造时间，返回新的结束时间
func (b *Building) Accelerate(speed float64, now int64) int64 {
	if b.State != BuildingConstructing {
		return 1 << 62
	}
	remain := b.EndTime - now
	b.EndTime = int64(float64(now) + float64(remain)/speed)
	return b.EndTime
}

// 发起建造，设置开始和结束时间
func (b *Building) Start(buildSeconds int64, now int64) error {
	if b.State != BuildingIdle {
		return errors.New("not int idle state")
	}
	b.State = BuildingConstructing
	b.EndTime = now + buildSeconds
	b.StartTime = now
	return nil
}

func TestBuild(t *testing.T) {
	now := time.Now().Unix()
	b := &Building{ID: 1}
	// 发起建造（耗时 1 小时）
	b.Start(3600, now)
	// 玩家 30 分钟后上线，加速 1.5 倍

	b.Accelerate(1.5, now+1800)
	// 刷新状态
	nowState := b.Refresh(now + 3600)
	if nowState != BuildingFinished {
		fmt.Errorf("状态错误")
	} else {
		fmt.Printf("建造完成 \n")
	}
}
