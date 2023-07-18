//go:build broken
// +build broken

package broken_test

import "testing"

func TestDoesNotCompile(t *testing.T) {
	a := 123
}
