package timeanchor

import (
	"testing"
	"time"
)

func TestTimeAnchor_GetLastDotTime(t1 *testing.T) {
	type fields struct {
		TimeAnchorPoint string
	}
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "上次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 30 * time.Minute,
			},
		},
		{
			name: "上次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 25 * time.Minute,
			},
		},
		{
			name: "上次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 3 * time.Minute,
			},
		},
		{
			name: "上次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 5 * time.Minute,
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTimeAnchor(tt.fields.TimeAnchorPoint, 0)
			got := t.GetLastDotTime(tt.args.interval)
			t1.Logf("GetLastDotTime() got = %v", got)
		})
	}
}

func TestTimeAnchor_GetNextDotTime(t1 *testing.T) {
	type fields struct {
		TimeAnchorPoint string
	}
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "下次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 30 * time.Minute,
			},
		},
		{
			name: "下次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 25 * time.Minute,
			},
		},
		{
			name: "下次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 3 * time.Minute,
			},
		},
		{
			name: "下次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				interval: 5 * time.Minute,
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTimeAnchor(tt.fields.TimeAnchorPoint, 0)
			got := t.GetNextDotTime(tt.args.interval)
			t1.Logf("GetNextDotTime() got = %v", got)
		})
	}
}

func TestTimeAnchor_SetAnchorTime(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		timeAnchorPoint string
		// Named input parameters for target function.
		anchorTime time.Time
		// 描点时间间隔
		interval time.Duration
	}{
		// TODO: Add test cases.
		{
			name:            "设置锚点时间测试",
			timeAnchorPoint: "2025-12-01 00:00:00 +0000",
			anchorTime:      time.Now(),
			interval:        5 * time.Minute,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := NewTimeAnchor(tt.timeAnchorPoint, 0)
			ti.SetAnchorTime(tt.anchorTime, tt.interval)
			got := ti.GetAnchorTime()
			lastDotTime := ti.GetInnerLastDotTime()
			nextDotTime := ti.GetInnerNextDotTime()
			t.Logf("GetAnchorTime() got = %v, lastDotTime = %v, nextDotTime = %v", got, lastDotTime, nextDotTime)
		})
	}
}

func TestTimeAnchor_GetInnerDotTime(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		timeAnchorPoint string

		interval time.Duration
		// Named input parameters for target function.
		current_offset int64
		want           time.Time
		want2          int64
	}{
		// TODO: Add test cases.
		{
			name:            "获取当前描点时间测试",
			timeAnchorPoint: "2025-12-01 00:00:00 +0000",
			interval:        5 * time.Minute,
			current_offset:  0,
			want:            time.Now(),
			want2:           0,
		},
		{
			name:            "获取下次描点时间测试",
			timeAnchorPoint: "2025-12-01 00:00:00 +0000",
			interval:        5 * time.Minute,
			current_offset:  1,
			want:            time.Now(),
			want2:           1,
		},
		{
			name:            "获取上次描点时间测试",
			timeAnchorPoint: "2025-12-01 00:00:00 +0000",
			interval:        5 * time.Minute,
			current_offset:  -1,
			want:            time.Now(),
			want2:           -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := NewTimeAnchor(tt.timeAnchorPoint, tt.interval)
			got, got2 := ti.GetInnerDotTime(tt.current_offset)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetInnerDotTime() = %v, want %v", got, tt.want)
			}
			if true {
				t.Errorf("GetInnerDotTime() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
