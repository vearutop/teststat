//go:build broken
// +build broken

package other_test

import "testing"

func TestAlwaysFails(t *testing.T) {
	t.Fail()
}
