# time-anchor

基于时间锚点的周期刻度计算工具，支持按固定时间间隔划分周期。

## 安装

```bash
go get github.com/CppToGo/go-utils/time-anchor
```

## 使用

```go
import "github.com/CppToGo/go-utils/time-anchor"

// 创建锚点，默认为 2025-12-01 00:00:00 +0000
ta := time_anchor.NewTimeAnchor("2025-12-01 00:00:00 +0000", 30*time.Minute)

// 获取最近刻度（当前分段的前一个整点）
last := ta.GetLastDotTime(30 * time.Minute)

// 获取下一刻度（当前分段的后一个整点）
next := ta.GetNextDotTime(30 * time.Minute)

// 获取内部分段点时间和分段编号
dotTime, dotNum := ta.GetInnerDotTime(currentOffset)

// 获取当前分段的前一个整分段时间点
innerLast := ta.GetInnerLastDotTime()

// 获取当前分段的后一个整分段时间点
innerNext := ta.GetInnerNextDotTime()
```

## 核心 API

| 方法 | 说明 |
|------|------|
| `NewTimeAnchor(timeAnchorPoint string, interval time.Duration)` | 创建时间锚点，空字符串使用默认锚点 |
| `SetAnchorTime(anchorTime time.Time, interval time.Duration)` | 设置锚点时间和间隔 |
| `GetAnchorTime() time.Time` | 获取锚点时间 |
| `GetInterval() time.Duration` | 获取时间间隔 |
| `GetLastDotTime(interval time.Duration) time.Time` | 获取指定间隔下的前一个分段点时间 |
| `GetNextDotTime(interval time.Duration) time.Time` | 获取指定间隔下的后一个分段点时间 |
| `GetInnerDotTime(currentOffset int64) (time.Time, int64)` | 获取内部分段点时间和分段编号 |
| `GetInnerLastDotTime() time.Time` | 获取当前分段的前一个整分段时间点 |
| `GetInnerNextDotTime() time.Time` | 获取当前分段的后一个整分段时间点 |
