//go:build broken

package other_test

import (
	"testing"
)

func TestAlwaysFails(t *testing.T) {
	t.Fail()
}

func TestThatPanics(t *testing.T) {
	t.Parallel()

	panic("ouch")
}

func TestAlwaysFailsInSubtest(t *testing.T) {
	t.Run("/-&?\\[]!@#$%^*()abc123_+=", func(t *testing.T) {
		panic("can't cope")
	})
	t.Run("/pas/pas$ses", func(t *testing.T) {
		println("HELLO WORLD!")
	})
}
