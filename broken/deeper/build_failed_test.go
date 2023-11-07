//go:build broken
// +build broken

package deeper_test

import "testing"

func TestDoesNotCompile(t *testing.T) {
	a := 123
}
