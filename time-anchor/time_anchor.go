package time_anchor

import (
	"log/slog"
	"time"
)

type TimeAnchor struct {
	timeAnchorPoint string
	anchorTime      time.Time
}

func NewTimeAnchor(timeAnchorPoint string) *TimeAnchor {
	if len(timeAnchorPoint) == 0 {
		// 默认 2025-12-01 00:00:00 +0000
		timeAnchorPoint = "2025-12-01 00:00:00 +0000"
	}
	anchorTime, err := time.Parse("2006-01-02 15:04:05 -0700", timeAnchorPoint)
	if err != nil {
		slog.Error("NewTimeAnchor() error", "error", err.Error())
		return nil
	}
	return &TimeAnchor{
		timeAnchorPoint: timeAnchorPoint,
		anchorTime:      anchorTime,
	}
}

func (t *TimeAnchor) GetAnchorTime() time.Time {
	return t.anchorTime
}

func (t *TimeAnchor) GetLastDotTime(intervalMinutes int64) time.Time {
	now := time.Now()
	if intervalMinutes <= 0 {
		return now
	}
	anchorTime := t.anchorTime
	if anchorTime.After(now) {
		return now
	}
	interval := time.Duration(intervalMinutes) * time.Minute

	duration := now.Sub(anchorTime)
	offset := duration / interval
	return anchorTime.Add(offset * interval)
}

func (t *TimeAnchor) GetNextDotTime(intervalMinutes int64) time.Time {
	now := time.Now()
	if intervalMinutes <= 0 {
		return now
	}
	anchorTime := t.anchorTime
	if anchorTime.After(now) {
		return now
	}
	interval := time.Duration(intervalMinutes) * time.Minute

	duration := now.Sub(anchorTime)
	offset := duration/interval + 1
	return anchorTime.Add(offset * interval)
}
