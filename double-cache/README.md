# DoubleCache

泛型双缓存库，基于 two-cache 模式：始终保持两个缓存实例（active + backup），后台定时刷新 backup，active 失效后立即切换。整个过程对调用方无感知。

## 核心设计

- **双槽交替**：两个缓存槽 `caches[0]` / `caches[1]`，通过 atomic cursor 标记当前活跃槽
- **无锁读**：读操作仅访问 active 槽，不加锁；写入/切换由后台 goroutine 处理
- **后台刷新**：独立 goroutine 定时刷新 backup 槽，刷新成功且 active 失效/过期时触发原子切换
- **Redis 通知**（可选）：通过 Redis 发布/订阅同步多实例更新状态；`redis.Client` 为 nil 时跳过

## 安装

```bash
go get github.com/CppToGo/go-utils/double-cache
```

## 快速开始

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/go-utils/double-cache"
)

rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379" })

cache := double_cache.NewCache(func() (map[string]any, error) {
    // 从数据库或远端加载数据
    return fetchData()
}, 5*time.Second, "my_channel", rdb)

cache.Start()
defer cache.Stop()

data, ok := cache.GetOrRefresh()
if ok {
    // use data
}
```

## API

### `NewCache`

```go
func NewCache[T any](
    refreshFn RefreshFunc[T],   // 数据刷新函数
    refreshRate time.Duration,  // 刷新间隔
    subChannel string,          // Redis 订阅频道名（可为空）
    redisCli *redis.Client,     // Redis 客户端（可为 nil）
) *DoubleCache[T]
```

### `Start() error`

启动后台刷新循环。应在所有缓存实例创建后调用一次。

### `Stop()`

发送停止信号，优雅关闭后台 goroutine。

### `GetOrRefresh() (T, bool)`

获取缓存数据。缓存未启动或不存在返回 zero value + false。
过期时触发后台切换，下次调用返回新数据。

### `ForceRefresh() error`

强制刷新 backup 并广播更新通知到 Redis 频道。

### `Close()`

取消 context，停止所有后台协程（等同于 `Stop()`）。

### `Stats() (hits, total int64)`

缓存命中率统计。

## 线程安全

- 读操作：`get()` 无锁访问 active 槽
- 写操作：`refreshBackup` / `switchToBackup` 使用 `sync.RWMutex`
- 切换操作：`cursor` 使用 `atomic.CompareAndSwap` 保证原子切换

## Redis 依赖（可选）

Redis 仅用于多实例间同步更新信号，并非必需。传 `nil` 可完全本地运行：

```go
cache := double_cache.NewCache(fn, 5*time.Second, "", nil)
```

## 依赖

- Go 1.21+
- `github.com/redis/go-redis/v9`
- 标准库：`context`, `sync/atomic`, `time`, `crypto/sha256`, `encoding/json`
