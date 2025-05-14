//go:build broken

package tfatalf

import (
	"errors"
	"testing"
)

func TestFoo(t *testing.T) {
	err := errors.New("foo")
	t.Fatalf(err.Error())
}
