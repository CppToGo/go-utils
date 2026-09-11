package double_cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	cacheSize              = 2
	maxConsecutiveFailures = 5

	redisKeyPrefix = "double_cache:refresh_signal:%s"
)

// DoubleCache 泛型双缓存结构
type DoubleCache[T any] struct {
	caches              []*cacheInstance[T]
	cursor              atomic.Uint32
	refreshMu           sync.RWMutex // 防止并发刷新
	refreshFn           RefreshFunc[T]
	refreshRate         time.Duration
	lastRefreshTime     atomic.Int64
	ticker              *time.Ticker
	stopChan            chan struct{}
	started             atomic.Bool
	consecutiveFailures atomic.Int32
	maxFailures         int32
	changeSignal        chan uint32
	hits                atomic.Int64 // 新增：缓存命中计数
	total               atomic.Int64 // 新增：总访问计数
	redisClient         *redis.Client
	subChannel          string // 订阅频道名称
	key                 string // 自己独立的key
	ctx                 context.Context
	cancel              context.CancelFunc
	version             atomic.Value
	syncSignal          chan string
}

// cacheInstance 缓存实例
type cacheInstance[T any] struct {
	data   T
	expiry time.Time
	hash   string
}

func genHash[T any](ctx context.Context, data T) string {
	// 更通用的空值判断
	if reflect.ValueOf(data).IsZero() {
		return ""
	}

	// 如果已有哈希值且数据未变，可考虑直接返回（需要额外的修改标记）
	// 这里简单处理为每次调用都重新计算
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal data", "error", err, "ctx", ctx)
		return ""
	}

	// 计算哈希
	hashBytes := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hashBytes)
}

// RefreshFunc 刷新缓存的函数类型
type RefreshFunc[T any] func() (T, error)

// NewCache 创建新的缓存实例
func NewCache[T any](refreshFn RefreshFunc[T], refreshRate time.Duration, subChannel string, redisCli *redis.Client) *DoubleCache[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &DoubleCache[T]{
		caches:              make([]*cacheInstance[T], cacheSize),
		cursor:              atomic.Uint32{},
		refreshMu:           sync.RWMutex{},
		refreshFn:           refreshFn,
		refreshRate:         refreshRate,
		lastRefreshTime:     atomic.Int64{},
		ticker:              nil,
		stopChan:            make(chan struct{}),
		started:             atomic.Bool{},
		consecutiveFailures: atomic.Int32{},
		maxFailures:         maxConsecutiveFailures,
		changeSignal:        make(chan uint32, 1),
		hits:                atomic.Int64{},
		total:               atomic.Int64{},
		redisClient:         redisCli,
		subChannel:          subChannel,
		key:                 uuid.New().String(),
		ctx:                 ctx,
		cancel:              cancel,
		version:             atomic.Value{},
		syncSignal:          make(chan string, 1),
	}
}

// Start 启动缓存刷新
func (c *DoubleCache[T]) Start() error {
	if c.started.Load() {
		return nil
	}
	if c.ticker != nil {
		return nil // 已启动
	}

	// 确保初始缓存已加载
	data, err := c.refreshFn()
	if err != nil {
		slog.Error("初始化缓存失败 刷新方法错误", "error", err, "ctx", c.ctx)
		return errors.New("初始化缓存失败")
	}

	// 初始化缓存
	freshCache := &cacheInstance[T]{
		data:   data,
		expiry: time.Now().Add(c.refreshRate * 177 / 1e2),
		hash:   genHash(c.ctx, data),
	}
	c.refreshMu.Lock()
	cursor := c.getCursor()
	c.caches[cursor] = freshCache
	slog.InfoContext(c.ctx, "初始化 刷新备用缓存完毕", "subChannel", c.subChannel, "key", c.key, "time", time.Now(), "expiry", freshCache.expiry, "hash", freshCache.hash)
	c.refreshMu.Unlock()

	// 启动定时刷新
	c.ticker = time.NewTicker(c.refreshRate)
	c.started.Store(true)

	go c.refreshLoop()
	go c.dealSignal()
	// 启动订阅协程，监听更新通知
	go c.subscribeToUpdates()

	return nil
}

// Stop 停止缓存刷新
func (c *DoubleCache[T]) Stop() {
	if !c.started.Load() {
		return
	}
	c.stopChan <- struct{}{}
}

// refreshLoop 后台刷新循环
func (c *DoubleCache[T]) refreshLoop() {
	backoffDuration := c.refreshRate
	for range c.ticker.C {
		if !c.started.Load() {
			return
		}
		// 定期刷新备用缓存
		//logc.Infof(c.ctx,"key %s ticker.C 刷新 %v", c.key, time.Now())
		if _, err := c.refreshBackup(false); err != nil {
			//logc.Infof(c.ctx,"subChannel %s key %s 缓存刷新失败: %v", c.subChannel, c.key, err)
			failures := c.consecutiveFailures.Add(1)
			if failures >= c.maxFailures {
				//logc.Infof(c.ctx,"subChannel %s key %s 连续失败次数达到阈值 %d，暂停刷新 %v", c.subChannel, c.key, c.maxFailures, backoffDuration)
				time.Sleep(backoffDuration)
				backoffDuration = min(backoffDuration<<1, 1*time.Minute) // 指数退避
			}
		} else {
			c.consecutiveFailures.Store(0)
			backoffDuration = c.refreshRate // 重置退避时间
		}
	}
}

// 处理切换信号和stop信号,把缓存刷新信号单独处理
func (c *DoubleCache[T]) dealSignal() {
	for {
		select {
		case oldCursor := <-c.changeSignal:
			// 收到切换信号, 准备切换缓存
			swaped, backupHash := c.switchToBackup(oldCursor)
			slog.InfoContext(c.ctx, "changeSignal 准备切换缓存", "subChannel", c.subChannel, "key", c.key, "time", time.Now(), "oldCursor", oldCursor, "swaped", swaped, "backupHash", backupHash)
		case keyTimestamp := <-c.syncSignal:
			slog.InfoContext(c.ctx, "syncSignal 缓存刷新信号", "subChannel", c.subChannel, "key", c.key, "keyTimestamp", keyTimestamp)
			if c.redisClient != nil {
				if err := c.redisClient.Set(c.ctx, fmt.Sprintf(redisKeyPrefix, c.subChannel), keyTimestamp, 60*time.Second); err != nil {
					slog.Error("缓存刷新信号发送失败", "subChannel", c.subChannel, "key", c.key, "error", err, "ctx", c.ctx)
				}
			}
		case <-c.stopChan:
			c.refreshMu.Lock()
			close(c.stopChan)
			c.started.Store(false)

			// 清理资源
			for i := range c.caches {
				c.caches[i] = nil
			}
			c.caches = nil
			c.cursor.Store(0)
			c.consecutiveFailures.Store(0)
			if c.ticker != nil {
				c.ticker.Stop()
			}
			c.refreshMu.Unlock()

			// 清理redis缓存
			if c.redisClient != nil {
				if _, err := c.redisClient.Del(c.ctx, fmt.Sprintf(redisKeyPrefix, c.subChannel)).Result(); err != nil {
					slog.Error("缓存清理失败", "subChannel", c.subChannel, "key", c.key, "error", err, "ctx", c.ctx)
				}
			}
			return
		}
	}
}

// refreshBackup 刷新备用缓存
func (c *DoubleCache[T]) refreshBackup(isForce bool) (string, error) {
	lastRefreshTime := c.lastRefreshTime.Load()
	// 如果距离上次刷新时间不到三分之二的刷新时间，跳过刷新
	if !isForce && lastRefreshTime > 0 && time.Since(time.Unix(0, lastRefreshTime)) < c.refreshRate/3*2 {
		//logc.Infof(context.Background(),"key %s 距离上次刷新时间%v，当前时间 %v 跳过刷新 %v", c.key, time.Unix(0, lastRefreshTime), time.Now(), time.Since(time.Unix(0, lastRefreshTime)))
		return "", nil
	}

	if !c.started.Load() {
		return "", nil
	}
	data, err := c.refreshFn()
	if err != nil {
		return "", err
	}
	// 更新备用缓存
	freshCache := &cacheInstance[T]{
		data:   data,
		expiry: time.Now().Add(c.refreshRate * 177 / 1e2),
		hash:   genHash(c.ctx, data),
	}
	c.refreshMu.Lock()
	cursor := c.getCursor()
	backupIdx := (cursor + 1) % 2
	c.caches[backupIdx] = freshCache
	slog.InfoContext(c.ctx, "刷新备用缓存完毕", "subChannel", c.subChannel, "key", c.key, "time", time.Now(), "expiry", freshCache.expiry, "backupIdx", backupIdx, "hash", freshCache.hash)
	activeCache := c.caches[cursor]
	c.refreshMu.Unlock()
	// 仅当强制刷新 或者 活跃缓失效 或 存过期时，才触发切换
	if isForce || activeCache == nil || time.Now().After(activeCache.expiry) {
		// 直接尝试切换（无需发送信号）
		swaped, backupHash := c.switchToBackup(cursor)
		slog.InfoContext(c.ctx, "直接切换备用缓存", "subChannel", c.subChannel, "key", c.key, "cursor", cursor, "isForce", isForce, "activeCacheNil", activeCache == nil, "swaped", swaped, "backupHash", backupHash)
	}
	c.lastRefreshTime.Store(time.Now().UnixNano())
	return freshCache.hash, nil
}

// Get 获取当前活跃缓存
func (c *DoubleCache[T]) get() (T, bool) {
	if !c.started.Load() {
		var zero T
		return zero, false
	}
	c.total.Add(1)
	cursor := c.getCursor()
	c.refreshMu.RLock() // 注意：这里仍需读锁保护 c.caches 切片本身的访问
	cacheInst := c.caches[cursor]
	c.refreshMu.RUnlock()
	if cacheInst == nil {
		// 极端情况：可能是初始化未完成，尝试加锁重试一次
		c.refreshMu.RLock()
		cacheInst = c.caches[cursor]
		c.refreshMu.RUnlock()
		if cacheInst == nil {
			var zero T
			return zero, false
		}
	}

	// 检查过期
	if time.Now().After(cacheInst.expiry) {
		// 即使过期，仍返回缓存数据
		select {
		case c.changeSignal <- cursor:
			// 发送切换信号
			slog.InfoContext(c.ctx, "Get缓存已过期，已发送切换信号", "subChannel", c.subChannel, "key", c.key, "time", time.Now(), "expiry", cacheInst.expiry, "cursor", cursor)
		default:
			slog.Warn("Get缓存已过期，但无法发送切换信号", "subChannel", c.subChannel, "key", c.key, "time", time.Now(), "expiry", cacheInst.expiry, "cursor", cursor, "ctx", c.ctx)
		}
	}
	c.hits.Add(1)
	return cacheInst.data, true
}

// GetOrRefresh 获取缓存，如果过期则尝试刷新
func (c *DoubleCache[T]) GetOrRefresh() (T, bool) {
	if !c.started.Load() {
		var zero T
		return zero, false
	}
	// 先尝试获取现有缓存
	if data, ok := c.get(); ok {
		return data, true
	}

	// 缓存无效时，直接刷新备用缓存并尝试获取
	if _, err := c.refreshBackup(true); err != nil {
		var zero T
		return zero, false
	}

	return c.get()
}

// 手动刷新方法
func (c *DoubleCache[T]) ForceRefresh() error {
	// 2. 发布更新通知到Redis频道
	if !c.started.Load() {
		return nil
	}
	hash, err := c.refreshBackup(true)
	if err != nil {
		return err
	}
	c.syncSignal <- fmt.Sprintf("%s_%d_%s", c.key, time.Now().UnixMilli(), hash)
	return err
}

// 切换到备用实例（原子操作）
func (c *DoubleCache[T]) switchToBackup(old uint32) (bool, string) {
	//c.cursor.Add(1) // 原子切换索引
	// 先检查备用缓存是否有效
	backupIdx := (old + 1) & 0x1
	c.refreshMu.RLock()
	backUpData := c.caches[backupIdx]
	c.refreshMu.RUnlock()
	backupValid := backUpData != nil && !time.Now().After(backUpData.expiry)
	backUpHash := ""
	// 仅当备用缓存有效时才切换
	swaped := false
	if backupValid {
		backUpHash = backUpData.hash
		swaped = c.cursor.CompareAndSwap(old, backupIdx)
	}
	slog.InfoContext(c.ctx, "switchToBackup 缓存切换", "subChannel", c.subChannel, "key", c.key, "backupValid", backupValid, "old", old, "backupIdx", backupIdx, "cursor", c.getCursor(), "swaped", swaped, "hash", backUpHash)
	return swaped, backUpHash
}

func (c *DoubleCache[T]) getCursor() uint32 {
	return c.cursor.Load() & 0x1
}

// 新增：获取缓存统计信息
func (c *DoubleCache[T]) stats() (hits, total int64) {
	return c.hits.Load(), c.total.Load()
}

// 订阅Redis通知频道，接收更新消息
func (c *DoubleCache[T]) subscribeToUpdates() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	slog.InfoContext(c.ctx, "已订阅更新频道", "subChannel", c.subChannel, "key", c.key, "channel", fmt.Sprintf(redisKeyPrefix, c.subChannel))

	for {
		select {
		case <-c.ctx.Done():
			return // 退出信号
		case _, ok := <-ticker.C:
			if !ok {
				// 频道关闭，重试订阅
				time.Sleep(1 * time.Second)
				continue
			}
			// 检查是否有新的版本
			if c.redisClient == nil {
				continue
			}
			keyTimestampHash, err := c.redisClient.Get(c.ctx, fmt.Sprintf(redisKeyPrefix, c.subChannel)).Result()
			if err != nil || err == redis.Nil {
				continue
			}
			if len(keyTimestampHash) == 0 {
				c.version.Store("")
				continue
			}
			var key string
			parts := strings.Split(keyTimestampHash, "_")
			if len(parts) != 3 {
				continue
			}
			key = parts[0]
			hash := parts[2]
			var myVersion = c.version.Load().(string)
			slog.InfoContext(c.ctx, "收到更新通知", "subChannel", c.subChannel, "key", c.key, "keyTimestampHash", keyTimestampHash, "localVersion", myVersion)
			if len(myVersion) > 0 {
				// 如果是版本一致则不需要更新
				if keyTimestampHash == myVersion {
					continue
				}
			}

			// 检查是不是自己发出的跟新通知才更新
			if key != c.key {
				slog.InfoContext(c.ctx, "收到其他key的更新通知", "subChannel", c.subChannel, "key", c.key, "otherKey", key, "hash", hash)
				NewHash, err := c.refreshBackup(true)
				if err != nil {
					slog.Error("更新配置失败", "subChannel", c.subChannel, "key", c.key, "error", err, "ctx", c.ctx)
				}
				if hash != NewHash {
					slog.InfoContext(c.ctx, "收到其他key的更新通知", "subChannel", c.subChannel, "key", c.key, "otherKey", key, "hash", hash, "localHash", NewHash)
					newKeyTimestampHash := fmt.Sprintf("%s_%d_%s", c.key, time.Now().UnixMilli(), NewHash)
					c.syncSignal <- newKeyTimestampHash
					c.version.Store(newKeyTimestampHash)
				} else {
					slog.InfoContext(c.ctx, "hash一致", "subChannel", c.subChannel, "key", c.key, "equal", hash == NewHash)
					c.version.Store(keyTimestampHash)
				}
			}
		}
	}
}

// 关闭缓存实例
func (c *DoubleCache[T]) Close() {
	c.cancel()
}
