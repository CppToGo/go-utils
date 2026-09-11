package double_cache

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// 创建缓存组件，每5秒刷新一次
	c := NewCache(func() (map[string]interface{}, error) {
		// 模拟从数据源加载数据
		time.Sleep(time.Duration(rand.Int64N(600)) * time.Millisecond)
		return map[string]interface{}{
			"key1": "value1",
			"key2": 12345,
			"key3": time.Now().String(),
		}, nil
	}, 5*time.Second, "", nil)

	// 启动缓存
	if err := c.Start(); err != nil {
		t.Fatalf("启动缓存失败: %v", err)
	}
	defer c.Stop()

	// 测试读取缓存
	go func() {
		for {
			testReadCache(c, t)
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			testReadCache(c, t)
			time.Sleep(50 * time.Millisecond)
		}
	}()
	go func() {
		for {
			testReadCache(c, t)
			time.Sleep(5 * time.Millisecond)
		}
	}()
	go func() {
		for {
			testReadCache(c, t)
			time.Sleep(1 * time.Millisecond)
		}
	}()
	// 再次测试读取缓存
	testReadCache(c, t)
	time.Sleep(20 * time.Second)
}

func testReadCache[T any](c *DoubleCache[T], t *testing.T) {
	// 测试读取缓存
	if val, ok := c.GetOrRefresh(); !ok {
		if c.started.Load() {
			t.Error("获取缓存失败")
		}
		hits, total := c.stats()
		fmt.Printf("当前缓存索引 :%d 获取缓存: %v state:%d %d \n", c.getCursor(), val, hits, total)
	}
	//else {
	//	hits, total := c.stats()
	//	fmt.Printf("当前缓存索引 :%d 获取缓存: %v state:%d %d \n", c.getCursor(), val, hits, total)
	//}
}

// 测试不启动直接获取是否正常
func TestCacheNotStart(t *testing.T) {
	// 创建缓存组件，每5秒刷新一次
	c := NewCache(func() (map[string]interface{}, error) {
		// 模拟从数据源加载数据
		time.Sleep(time.Duration(rand.Int64N(600)) * time.Millisecond)
		return map[string]interface{}{
			"key1": "value1",
			"key2": 12345,
			"key3": time.Now().String(),
		}, nil
	}, 5*time.Second, "", nil)

	// 测试读取缓存
	if _, ok := c.GetOrRefresh(); !ok && c.started.Load() {
		t.Error("获取缓存失败")
	}
}

// 写一个基准测试
func BenchmarkCache(t *testing.B) {
	//创建缓存组件，每5秒刷新一次
	c := NewCache(func() (map[string]interface{}, error) {
		// 模拟从数据源加载数据
		time.Sleep(time.Duration(670 * time.Millisecond))
		return map[string]interface{}{
			"key1": "value1",
			"key2": 12345,
			"key3": time.Now().String(),
		}, nil
	}, 30*time.Second, "", nil)

	// 启动缓存
	if err := c.Start(); err != nil {
		t.Fatalf("启动缓存失败: %v", err)
	}
	defer c.Stop()

	time.Sleep(1 * time.Second)
	t.ResetTimer()
	// 测试读取缓存
	for i := 0; i < t.N; i++ {
		// 测试读取缓存
		if _, ok := c.GetOrRefresh(); !ok && c.started.Load() {
			t.Error("获取缓存失败")
			c.stopChan <- struct{}{}
		} else {
			//hits, total := c.stats()
			//fmt.Printf("当前缓存索引 :%d 获取缓存: %v state:%d %d \n", c.cursor.Load(), val, hits, total)
		}
	}
}
