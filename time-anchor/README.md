# time-anchor

基于时间锚点的周期刻度计算工具。

## 功能

根据给定的锚点时间，计算指定时间间隔内的最近刻度和下一刻度。

## 安装

```bash
go get github.com/CppToGo/go-utils/time-anchor
```

## 使用

```go
import "github.com/CppToGo/go-utils/time-anchor"

// 创建锚点，默认为 2025-12-01 00:00:00 +0000
ta := time_anchor.NewTimeAnchor("2025-12-01 00:00:00 +0000")

// 获取最近刻度（给定时间间隔内的前一个整点）
last := ta.GetLastDotTime(30) // 30 分钟间隔

// 获取下一刻度（给定时间间隔内的后一个整点）
next := ta.GetNextDotTime(30) // 30 分钟间隔
```

## 核心 API

| 方法 | 说明 |
|------|------|
| `NewTimeAnchor(timeAnchorPoint string)` | 创建时间锚点，空字符串使用默认锚点 |
| `GetAnchorTime() time.Time` | 获取锚点时间 |
| `GetLastDotTime(intervalMinutes int64) time.Time` | 获取最近刻度（向前对齐） |
| `GetNextDotTime(intervalMinutes int64) time.Time` | 获取下一刻度（向后对齐） |
