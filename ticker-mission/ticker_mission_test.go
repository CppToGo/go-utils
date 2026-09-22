package tickermission

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTickerMission(t *testing.T) {
	ctx := context.Background()
	lock := NewRedisLock("localhost:6379", "test-lock", time.Second)
	settleFunc := func(ctx context.Context) error { return nil }

	tm := NewTickerMission(ctx, lock, settleFunc, time.Second)

	if tm == nil {
		t.Fatal("NewTickerMission returned nil")
	}
	if tm.ctx == nil {
		t.Error("ctx should not be nil")
	}
	if tm.cancel == nil {
		t.Error("cancel should not be nil")
	}
	if tm.redislock == nil {
		t.Error("redislock should not be nil")
	}
	if tm.settleFunc == nil {
		t.Error("settleFunc should not be nil")
	}
	if tm.frequency != time.Second {
		t.Errorf("frequency = %v, want %v", tm.frequency, time.Second)
	}
}

func TestTickerMission_StartOnce(t *testing.T) {
	ctx := context.Background()
	lock := NewRedisLock("localhost:6379", "test-lock", time.Second)
	callCount := atomic.Int32{}

	settleFunc := func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	}

	tm := NewTickerMission(ctx, lock, settleFunc, 10*time.Millisecond)

	// Start multiple times - should only execute once due to startOnce
	tm.Start()
	tm.Start()
	tm.Start()

	time.Sleep(50 * time.Millisecond)
	tm.Stop()

	if count := callCount.Load(); count < 1 {
		t.Errorf("settleFunc called %d times, want at least 1", count)
	}
}

func TestTickerMission_StopOnce(t *testing.T) {
	ctx := context.Background()
	lock := NewRedisLock("localhost:6379", "test-lock", time.Second)
	callCount := atomic.Int32{}

	settleFunc := func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	}

	tm := NewTickerMission(ctx, lock, settleFunc, 10*time.Millisecond)

	tm.Start()
	time.Sleep(20 * time.Millisecond)

	// Stop multiple times - should not panic due to stopOnce
	tm.Stop()
	tm.Stop()
	tm.Stop()
}

func TestTickerMission_StopWithoutStart(t *testing.T) {
	ctx := context.Background()
	lock := NewRedisLock("localhost:6379", "test-lock", time.Second)

	settleFunc := func(ctx context.Context) error { return nil }

	tm := NewTickerMission(ctx, lock, settleFunc, time.Second)

	// Stop without Start should not panic
	tm.Stop()
}

func TestRedisLock_AcquireCtx(t *testing.T) {
	lock := NewRedisLock("localhost:6379", "test-key", time.Second)

	ok, err := lock.AcquireCtx(context.Background())
	if err != nil {
		t.Errorf("AcquireCtx returned error: %v", err)
	}
	if !ok {
		t.Error("AcquireCtx returned false, want true")
	}
}

func TestRedisLock_ReleaseCtx(t *testing.T) {
	lock := NewRedisLock("localhost:6379", "test-key", time.Second)

	err := lock.ReleaseCtx(context.Background())
	if err != nil {
		t.Errorf("ReleaseCtx returned error: %v", err)
	}
}
