package gyro

import (
	"errors"
	"sync"
	"time"
)

const (
	DEFAULT_FPS = 60
)

type InputFunc func()
type UpdateFunc func(deltaTime time.Duration)
type RenderFunc func()

type Loop struct {
	// Loop Config
	targetFps  int
	msPerFrame int
	stopCh     chan struct{}

	// Flags
	isDebugMode bool
	isRunning   bool

	// Loop functions
	input  InputFunc
	update UpdateFunc
	render RenderFunc

	// Runtime values
	currentFps uint16

	once sync.Once
}

func NewLoop() *Loop {
	l := &Loop{}
	l.SetTargetFps(DEFAULT_FPS)
	return l
}

func (l *Loop) SetDebug(debug bool) *Loop {
	l.isDebugMode = debug
	return l
}

func (l *Loop) SetTargetFps(fps int) *Loop {
	l.targetFps = max(fps, 1)
	l.msPerFrame = 1 / l.targetFps * 1000
	return l
}

func (l *Loop) GetTargetFps() int {
	return l.targetFps
}

func (l *Loop) GetCurrentFps() int {
	return int(l.currentFps)
}

func (l *Loop) IsRunning() bool {
	return l.isRunning
}

func (l *Loop) SetUpdateFunc(update UpdateFunc) *Loop {
	l.update = update
	return l
}

func (l *Loop) SetInputFunc(input InputFunc) *Loop {
	l.input = input
	return l
}

func (l *Loop) SetRenderFunc(render RenderFunc) *Loop {
	l.render = render
	return l
}

// Start attempts to start the game loop.
// It requires an update function to be set and
// it will run just once for each Loop instance.
func (l *Loop) Start() error {
	if l.update == nil {
		return errors.New(ERR_NO_UPDATE_FUNC)
	}

	if l.isRunning {
		return nil
	}

	l.isRunning = true
	l.run()
	return nil
}

// Stop attempts to stop the game loop by sending a stop signal
func (l *Loop) Stop() error {
	if !l.isRunning {
		return nil
	}

	l.isRunning = false
	close(l.stopCh)
	return nil
}

func (l *Loop) run() {
	l.stopCh = make(chan struct{})
	frameCounter := uint64(0)
	lastFrame := time.Now()
	lastSecond := time.Now()

	for {
		select {
		case <-l.stopCh:
			return
		default:
			start := time.Now()

			if l.input != nil {
				l.input()
			}

			if l.update != nil {
				// Call update with delta time
				l.update(time.Since(lastFrame))
			}

			if l.render != nil {
				l.render()
			}

			// Frame finished timestamp (input, update, render are done)
			lastFrame = time.Now()
			frameCounter++

			elapsed := time.Since(lastSecond)
			if elapsed > 1000 {
				l.currentFps = uint16(frameCounter * uint64(time.Second) / uint64(elapsed))
				lastSecond = time.Now()
				frameCounter = 0
			}

			sleepTime := time.Duration(l.msPerFrame)*time.Millisecond - time.Since(start)
			if sleepTime > 0 {
				time.Sleep(sleepTime)
			}
		}

	}
}
