package foo_test

import (
	"math/rand"
	"testing"
	"time"
)

func TestThatIsRacy(t *testing.T) {
	t.Parallel()

	a := 1

	for i := 0; i < 1000; i++ {
		go func() { a++ }()
	}

	time.Sleep(10 * time.Millisecond)
	a++
}

func TestThatFlakes(t *testing.T) {
	t.Parallel()

	if rand.Int()%3 == 0 { //nolint
		return
	}

	t.Fatal("oh, I'm so flaky")
}

func TestThatFlakesToo(t *testing.T) {
	t.Parallel()

	if rand.Int()%5 == 0 { //nolint
		return
	}

	t.Fatal("oh, I'm even more flaky")
}

func TestThatIsSometimesSlow(t *testing.T) {
	t.Parallel()

	if rand.Int()%3 == 0 { //nolint
		time.Sleep(1 * time.Second)
	}
}

func TestThatIsAlwaysSlow(t *testing.T) {
	t.Parallel()
	time.Sleep(1 * time.Second)
}
