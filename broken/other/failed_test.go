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
