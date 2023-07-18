//go:build imperfect
// +build imperfect

package foo_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var allGood = rand.Int()%3 == 0 //nolint

func TestThatIsRacyFoo(t *testing.T) {
	t.Parallel()

	// Sometimes passes.
	if passes(3) {
		return
	}

	a := 1

	for i := 0; i < 1000; i++ {
		go func() { a++ }()
	}

	time.Sleep(10 * time.Millisecond)
	a++
}

func TestThatFlakesFoo(t *testing.T) {
	t.Parallel()

	if passes(3) {
		return
	}

	t.Fatal("oh, I'm so flaky")
}

func TestThatFlakesTooFoo(t *testing.T) {
	t.Parallel()

	if passes(5) {
		return
	}

	t.Fatal("oh, I'm even more flaky")
}

func TestThatIsSometimesSlowFoo(t *testing.T) {
	t.Parallel()

	if passes(3) {
		time.Sleep(1 * time.Second)
	}
}

func TestThatIsAlwaysSlowFoo(t *testing.T) {
	t.Parallel()
	time.Sleep(1 * time.Second)
}

func Test_Suite(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

type TestSuite struct {
	suite.Suite
}

func (suite *TestSuite) TestThatFlakes() {
	suite.T().Parallel()

	// Sometimes passes.
	if passes(3) {
		return
	}

	suite.T().Fatal("oh, I'm so flaky")
}

func (suite *TestSuite) TestThatFlakesToo() {
	suite.T().Parallel()

	// Sometimes passes.
	if passes(5) {
		return
	}

	suite.T().Fatal("oh, I'm so flaky")
}

func passes(flakiness int) bool {
	if allGood {
		return true
	}

	// Sometimes passes.
	return rand.Int()%flakiness == 0 //nolint
}

func (suite *TestSuite) TestThatPasses() {
	time.Sleep(time.Millisecond)
}

func TestWithLongTrace(t *testing.T) {
	//if passes(15) {
	//	return
	//}

	var recursive func(depth int)
	recursive = func(depth int) {
		if depth == 0 {
			var a []string
			a[1] = "abc"
		}

		recursive(depth - 1)
	}

	recursive(50000)
}
