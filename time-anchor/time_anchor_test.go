package time_anchor

import (
	"testing"
)

func TestTimeAnchor_GetLastDotTime(t1 *testing.T) {
	type fields struct {
		TimeAnchorPoint string
	}
	type args struct {
		intervalMinutes int64
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
				intervalMinutes: 30,
			},
		},
		{
			name: "上次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				intervalMinutes: 25,
			},
		},
		{
			name: "上次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				intervalMinutes: 3,
			},
		},
		{
			name: "上次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				intervalMinutes: 5,
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTimeAnchor(tt.fields.TimeAnchorPoint)
			got := t.GetLastDotTime(tt.args.intervalMinutes)
			t1.Logf("GetLastDotTime() got = %v", got)
		})
	}
}

func TestTimeAnchor_GetNextDotTime(t1 *testing.T) {
	type fields struct {
		TimeAnchorPoint string
	}
	type args struct {
		intervalMinutes int64
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
				intervalMinutes: 30,
			},
		},
		{
			name: "下次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				intervalMinutes: 25,
			},
		},
		{
			name: "下次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				intervalMinutes: 3,
			},
		},
		{
			name: "下次描点函数测试",
			fields: fields{
				TimeAnchorPoint: "2025-12-01 00:00:00 +0000",
			},
			args: args{
				intervalMinutes: 5,
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTimeAnchor(tt.fields.TimeAnchorPoint)
			got := t.GetNextDotTime(tt.args.intervalMinutes)
			t1.Logf("GetNextDotTime() got = %v", got)
		})
	}
}
