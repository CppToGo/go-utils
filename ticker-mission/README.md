# ticker-mission

基于 Redis 分布式锁的周期任务调度器，支持定时执行和分布式环境下的幂等启动。

## 核心设计

- **幂等启动**：通过 `sync.Once` 保证 `Start()` 只真正执行一次
- **分布式锁**：使用 Redis SETNX 实现分布式互斥，多实例不会重复执行
- **优雅停止**：先取消 context 再等待 goroutine 退出，确保任务彻底关闭
- **启动即执行**：启动时立即执行一次任务，无需等待首次 ticker 触发

## 安装

```bash
go get github.com/CppToGo/go-utils/ticker-mission
```

## 快速开始

```go
import (
    "context"
    "time"
    "github.com/CppToGo/go-utils/ticker-mission"
)

func main() {
    ctx := context.Background()
    lock := tickermission.NewRedisLock("localhost:6379", "my-task-lock", 30*time.Second)
    
    mission := tickermission.NewTickerMission(ctx, lock, func(ctx context.Context) error {
        // 执行周期任务逻辑
        return doSettle(ctx)
    }, time.Minute)
    
    mission.Start()
    defer mission.Stop()
}
```

## API

### `RedisLock`

分布式锁实现，通过 SETNX + TTL 保证互斥。

```go
func NewRedisLock(redisAddr, lockKey string, ttl time.Duration) *RedisLock
func (l *RedisLock) AcquireCtx(ctx context.Context) (bool, error)
func (l *RedisLock) ReleaseCtx(ctx context.Context) error
```

### `TickerMission`

周期任务调度器。

```go
func NewTickerMission(ctx context.Context, redislock *RedisLock, settleFunc func(ctx context.Context) error, frequency time.Duration) *TickerMission
func (p *TickerMission) Start()
func (p *TickerMission) Stop()
```

### 核心方法

| 方法 | 说明 |
|------|------|
| `NewRedisLock` | 创建 Redis 分布式锁实例 |
| `NewTickerMission` | 创建周期任务实例 |
| `Start` | 启动任务，幂等（多次调用只启动一次） |
| `Stop` | 停止任务，幂等（多次调用安全） |

## 线程安全

- `Start()`：使用 `sync.Once` 保证幂等
- `Stop()`：使用 `sync.Once` 保证幂等
- `RedisLock.AcquireCtx` / `ReleaseCtx`：需外部 Redis 保证原子性

## 依赖

- Go 1.21+
- 标准库：`context`, `log/slog`, `sync`, `time`

## 待改进

- [ ] **RedisLock 实际实现**：当前 `AcquireCtx` / `ReleaseCtx` 为空壳实现，实际使用时需接入 `go-redis/redis` 或其他 Redis 客户端
- [ ] **锁续期机制**：分布式锁的 TTL 到期前应主动续期，防止任务执行时间超过 TTL 导致锁自动释放
- [ ] **失败重试**：任务执行失败时可考虑增加重试机制和最大重试次数配置
- [ ] **执行状态暴露**：可添加 `IsRunning()` 方法或状态回调，暴露当前任务执行状态
- [ ] **可配置日志**：当前使用固定 `slog.Error`，可改为注入 `slog.Logger` 实现自定义日志
- [ ] **启动延迟**：可增加首次执行的延迟配置，避免所有实例启动时同时竞争锁
