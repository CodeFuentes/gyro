package gyro_test

import (
	"math"
	"testing"
	"time"

	"github.com/codefuentes/gyro"
)

func TestStartWithNoUpdate(t *testing.T) {
	err := gyro.NewLoop().
		SetTargetFps(60).
		Start()

	if err.Error() != gyro.ERR_NO_UPDATE_FUNC {
		t.Errorf("got %q, wanted %q", err, gyro.ERR_NO_UPDATE_FUNC)
	}
}

func TestTargetFps(t *testing.T) {
	targetFps := 3

	loop := gyro.NewLoop().
		SetTargetFps(targetFps)

	if targetFps != loop.GetTargetFps() {
		t.Fatalf("failed to set target fps: got %v, wanted %v", loop.GetTargetFps(), targetFps)
	}
}

func TestGetCurrentFps(t *testing.T) {
	targetFps := 7
	frameCounter := 0
	testTime := 2

	loop := gyro.NewLoop().
		SetTargetFps(targetFps).
		SetUpdateFunc(func(dt time.Duration) {
			frameCounter++
		})

	go func() {
		time.Sleep(time.Duration(testTime) * time.Second)
		loop.Stop()
	}()

	err := loop.Start()
	if err != nil {
		t.Fatalf("failed to start: %q", err.Error())
	}

	if loop.GetCurrentFps() != frameCounter/testTime {
		t.Fatalf("current fps mismatch: got %v, wanted %v", loop.GetCurrentFps(), frameCounter/testTime)
	}

}

func TestFrameRate(t *testing.T) {
	// Result can be 5 frames above or below the target
	tolerance := 5
	targetFps := 45

	loop := gyro.NewLoop().
		SetTargetFps(targetFps).
		SetUpdateFunc(func(dt time.Duration) {
			time.Sleep(1 * time.Millisecond)
		})

	go func() {
		time.Sleep(2 * time.Second)
		loop.Stop()
	}()

	err := loop.Start()
	if err != nil {
		t.Fatalf("failed to start: %q", err.Error())
	}

	if targetFps != loop.GetTargetFps() {
		t.Fatalf("failed to set target fps: got %v, wanted %v", loop.GetTargetFps(), targetFps)
	}

	diff := int(math.Abs(float64(loop.GetCurrentFps() - targetFps)))
	if diff > tolerance {
		t.Fatalf("frame count (%v) exceeds tolerance (max %v, got %v) for target fps %v", loop.GetCurrentFps(), tolerance, diff, targetFps)
	}

}
