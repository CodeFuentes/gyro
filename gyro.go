package gyro

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	DEFAULT_FPS       = 60
	MS_PER_CLOCK_TICK = 10
)

type inputFunc func()
type updateFunc func(time.Duration)
type renderFunc func()

type loop struct {
	// Loop Config
	targetFps  uint16
	msPerFrame float32
	quitCh     chan struct{}

	// Flags
	isDebugMode bool
	isSafeMode  bool
	isRunning   bool

	// Loop functions
	input  inputFunc
	update updateFunc
	render renderFunc

	// Runtime values
	currentFps uint16

	once sync.Once
}

func NewLoop() *loop {
	l := &loop{}
	l.SetTargetFps(DEFAULT_FPS)
	return l
}

func (l *loop) SetDebug(debug bool) *loop {
	l.isDebugMode = debug
	return l
}

func (l *loop) SetSafeMode(safe bool) *loop {
	if l.shouldPreventSensitiveChanges() {
		return l
	}
	l.isSafeMode = safe
	return l
}

func (l *loop) SetTargetFps(fps int) *loop {
	// FPS must be between 1 and 65535 (to fit uint16)
	l.targetFps = uint16(max(min(fps, 65535), 1))
	l.msPerFrame = 1000.0 / float32(l.targetFps)
	return l
}

func (l *loop) GetTargetFps() uint16 {
	return l.targetFps
}

func (l *loop) SetUpdateFunc(update updateFunc) *loop {
	if l.shouldPreventSensitiveChanges() {
		return l
	}

	l.update = update
	return l
}

func (l *loop) SetInputFunc(input inputFunc) *loop {
	if l.shouldPreventSensitiveChanges() {
		return l
	}

	l.input = input
	return l
}

func (l *loop) SetRenderFunc(render renderFunc) *loop {
	if l.shouldPreventSensitiveChanges() {
		return l
	}

	l.render = render
	return l
}

// Start attempts to start the game loop.
// It requires an update function to be set and
// it will run just once for each loop instance.
func (l *loop) Start() error {
	if l.update == nil {
		return errors.New(ERR_NO_UPDATE_FUNC)
	}
	l.isRunning = true
	l.once.Do(l.start)
	return nil
}

// Quit attempts to stop the game loop by sending a quit signal
func (l *loop) Quit() error {
	select {
	case l.quitCh <- struct{}{}:
		return nil
	default:
		return errors.New(ERR_QUIT_CHAN_BLOCKED)
	}
}

func (l *loop) IsRunning() bool {
	return l.isRunning
}

func (l *loop) GetCurrentFps() int {
	return int(l.currentFps)
}

func (l *loop) shouldPreventSensitiveChanges() bool {
	return l.isSafeMode && l.isRunning
}

func (l *loop) start() {
	var frameCounter int

	lastFrame := time.Now()
	lastSecond := time.Now()

	// Ticker to limit CPU usage
	gameClock := time.NewTicker(MS_PER_CLOCK_TICK * time.Millisecond)
	prevUpdateDone := true

	for {
		// Block game loop until quit signal or clock tick
		select {
		case <-l.quitCh:
			return
		case <-gameClock.C:
			msSinceLastFrame := float32(time.Since(lastFrame).Milliseconds())
			// Wait until next tick if we can't update yet
			if !prevUpdateDone || msSinceLastFrame < l.msPerFrame {
				continue
			}
		}

		// Frame logic
		prevUpdateDone = false
		go func() {
			// Call update with delta time
			l.update(time.Since(lastFrame))
			lastFrame = time.Now()
			prevUpdateDone = true

			frameCounter++
			msSinceLastSecond := time.Since(lastSecond).Milliseconds()
			l.currentFps = uint16(float32(frameCounter) / float32(msSinceLastSecond) * 1000)
			if msSinceLastSecond > 1000 {
				lastSecond = time.Now()
				frameCounter = 0
			}
			if l.isDebugMode {
				fmt.Println("fps", l.currentFps)
			}
		}()
	}
}
