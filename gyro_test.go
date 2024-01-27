package gyro_test

import (
	"testing"

	"github.com/codefuentes/gyro"
)

func TestStart(t *testing.T) {
	err := gyro.NewLoop().
		SetTargetFps(60).
		Start()

	if err.Error() != gyro.ERR_NO_UPDATE_FUNC {
		t.Errorf("got %q, wanted %q", err, gyro.ERR_NO_UPDATE_FUNC)
	}
}
