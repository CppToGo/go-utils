package timeanchor

import (
	"time"
)

// TimeAnchor 时间锚点计算器
// 用于计算基于固定锚点时间的时间点，支持按固定时间间隔划分周期
type TimeAnchor struct {
	timeAnchorPoint string        // 锚点时间的字符串格式
	anchorTime      time.Time     // 锚点时间
	interval        time.Duration // 时间间隔
}

// NewTimeAnchor 创建时间锚点计算器
// timeAnchorPoint: 锚点时间字符串，格式为 "2006-01-02 15:04:05 -0700"
// interval: 时间间隔，用于划分周期
func NewTimeAnchor(timeAnchorPoint string, interval time.Duration) *TimeAnchor {
	if len(timeAnchorPoint) == 0 {
		// 默认 2025-12-01 00:00:00 +0000
		timeAnchorPoint = "2025-12-01 00:00:00 +0000"
	}
	anchorTime, err := time.Parse("2006-01-02 15:04:05 -0700", timeAnchorPoint)
	if err != nil {
		return nil
	}
	return &TimeAnchor{
		timeAnchorPoint: timeAnchorPoint,
		anchorTime:      anchorTime,
		interval:        interval,
	}
}

// SetAnchorTime 设置锚点时间和间隔
// anchorTime: 新的锚点时间
// interval: 新的时间间隔
func (t *TimeAnchor) SetAnchorTime(anchorTime time.Time, interval time.Duration) {
	t.anchorTime = anchorTime
	t.timeAnchorPoint = anchorTime.Format("2006-01-02 15:04:05 -0700")
	t.interval = interval
}

// GetInnerDotTime 获取当前时间所属的内部分段点时间和分段编号
// current_offset: 当前的偏移量
// 返回: dotTime-分段点时间, dotNum-分段编号
// 当间隔<=0时返回锚点时间和0
// 当锚点时间在当前时间之后时返回当前时间和0
func (t *TimeAnchor) GetInnerDotTime(current_offset int64) (dotTime time.Time, dotNum int64) {
	now := time.Now()
	if t.interval <= 0 {
		return t.anchorTime, 0
	}
	anchorTime := t.anchorTime
	if anchorTime.After(now) {
		return now, 0
	}

	duration := now.Sub(anchorTime)
	offset := duration/t.interval + time.Duration(current_offset)
	return anchorTime.Add(offset * t.interval), int64(offset)
}

// GetInnerLastDotTime 获取当前时间所属分段的前一个分段点时间
// 返回最近的整分段时间点
// 当间隔<=0或锚点时间在当前时间之后时返回当前时间
func (t *TimeAnchor) GetInnerLastDotTime() time.Time {
	now := time.Now()
	if t.interval <= 0 {
		return now
	}
	anchorTime := t.anchorTime
	if anchorTime.After(now) {
		return now
	}

	duration := now.Sub(anchorTime)
	offset := duration / t.interval
	return anchorTime.Add(offset * t.interval)
}

// GetInnerNextDotTime 获取当前时间所属分段的后一个分段点时间
// 返回下一个整分段时间点
// 当间隔<=0或锚点时间在当前时间之后时返回当前时间
func (t *TimeAnchor) GetInnerNextDotTime() time.Time {
	now := time.Now()
	if t.interval <= 0 {
		return now
	}
	anchorTime := t.anchorTime
	if anchorTime.After(now) {
		return now
	}

	duration := now.Sub(anchorTime)
	offset := duration/t.interval + 1
	return anchorTime.Add(offset * t.interval)
}

// GetAnchorTime 获取锚点时间
func (t *TimeAnchor) GetAnchorTime() time.Time {
	return t.anchorTime
}

// GetInterval 获取时间间隔
func (t *TimeAnchor) GetInterval() time.Duration {
	return t.interval
}

// GetTimeAnchorPoint 获取锚点时间字符串
func (t *TimeAnchor) GetTimeAnchorPoint() string {
	return t.timeAnchorPoint
}

// GetLastDotTime 获取指定间隔下的前一个分段点时间
// interval: 自定义时间间隔
// 返回最近的整分段时间点
func (t *TimeAnchor) GetLastDotTime(interval time.Duration) time.Time {
	now := time.Now()
	if interval <= 0 {
		return now
	}
	anchorTime := t.anchorTime
	if anchorTime.After(now) {
		return now
	}

	duration := now.Sub(anchorTime)
	offset := duration / interval
	return anchorTime.Add(offset * interval)
}

// GetNextDotTime 获取指定间隔下的后一个分段点时间
// interval: 自定义时间间隔
// 返回下一个整分段时间点
func (t *TimeAnchor) GetNextDotTime(interval time.Duration) time.Time {
	now := time.Now()
	if interval <= 0 {
		return now
	}
	anchorTime := t.anchorTime
	if anchorTime.After(now) {
		return now
	}

	duration := now.Sub(anchorTime)
	offset := duration/interval + 1
	return anchorTime.Add(offset * interval)
}
