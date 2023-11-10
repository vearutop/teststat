//go:build broken

package goroutine_test

import (
	"testing"
	"time"
)

func TestThatPanicsInAGoroutine(t *testing.T) {
	t.Parallel()

	time.Sleep(time.Second)

	go func() {
		panic("ouch2")
	}()

	time.Sleep(time.Second)
}
