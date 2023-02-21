package main_test

import (
	"math/rand"
	"testing"
	"time"
)

func TestThatIsRacy(t *testing.T) {
	a := 1

	go func() { a++ }()
	go func() { a++ }()
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
		time.Sleep(5 * time.Second)
	}
}

func TestThatIsAlwaysSlow(t *testing.T) {
	t.Parallel()
	time.Sleep(5 * time.Second)
}
