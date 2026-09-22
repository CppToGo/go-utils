package tickermission

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RedisLock 简单的分布式锁实现
type RedisLock struct {
	redisAddr string
	lockKey   string
	ttl       time.Duration
}

// NewRedisLock 创建一个 Redis 分布式锁
func NewRedisLock(redisAddr, lockKey string, ttl time.Duration) *RedisLock {
	return &RedisLock{
		redisAddr: redisAddr,
		lockKey:   lockKey,
		ttl:       ttl,
	}
}

// AcquireCtx 尝试获取锁，返回是否成功
func (l *RedisLock) AcquireCtx(ctx context.Context) (bool, error) {
	// 使用 SETNX + TTL 实现
	// 这里需要 redis 客户端，暂用注释表示逻辑
	// 实际使用可配合 go-redis/redis 或标准库 net/rpc
	_ = l.redisAddr
	_ = l.lockKey
	_ = l.ttl
	_ = ctx
	// 简化：实际使用时替换为真实的 redis SETNX 调用
	return true, nil
}

// ReleaseCtx 释放锁
func (l *RedisLock) ReleaseCtx(ctx context.Context) error {
	_ = ctx
	return nil
}

type TickerMission struct {
	ctx        context.Context
	cancel     context.CancelFunc
	redislock  *RedisLock
	settleFunc func(ctx context.Context) error
	frequency  time.Duration
	wg         sync.WaitGroup
	startOnce  sync.Once
	stopOnce   sync.Once
}

// NewTickerMission 创建一个新的周期任务实例
func NewTickerMission(ctx context.Context, redislock *RedisLock, settleFunc func(ctx context.Context) error, frequency time.Duration) *TickerMission {
	myCtx, cancel := context.WithCancel(ctx)
	return &TickerMission{
		ctx:        myCtx,
		cancel:     cancel,
		redislock:  redislock,
		settleFunc: settleFunc,
		frequency:  frequency,
	}
}

// Start 启动周期任务。幂等，重复调用只会真正启动一次。
func (p *TickerMission) Start() {
	p.startOnce.Do(func() {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			ticker := time.NewTicker(p.frequency)
			defer ticker.Stop()

			if ok, _ := p.redislock.AcquireCtx(p.ctx); ok {
				// 启动时立即执行一次
				p.execute()
				p.redislock.ReleaseCtx(p.ctx)
			}

			for {
				select {
				case <-p.ctx.Done():
					return
				case <-ticker.C:
					if ok, _ := p.redislock.AcquireCtx(p.ctx); ok {
						p.execute()
						p.redislock.ReleaseCtx(p.ctx)
					}
				}
			}
		}()
	})
}

func (p *TickerMission) execute() {
	if err := p.settleFunc(p.ctx); err != nil {
		slog.Error("周期结算失败", "err", err)
	}
}

// Stop 停止周期任务：先取消 context，再等待 goroutine 完全退出。幂等，重复调用安全。
func (p *TickerMission) Stop() {
	p.stopOnce.Do(func() {
		p.cancel()
		p.wg.Wait()
	})
}
