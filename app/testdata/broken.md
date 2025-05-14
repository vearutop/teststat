### Failures
<details>
<summary>Failed builds</summary>

```
# github.com/vearutop/teststat/broken_test [github.com/vearutop/teststat/broken.test]
../../broken/build_failed_test.go:8:2: declared and not used: a
# github.com/vearutop/teststat/broken/deeper_test [github.com/vearutop/teststat/broken/deeper.test]
../../broken/deeper/build_failed_test.go:8:2: declared and not used: a
FAIL	github.com/vearutop/teststat/broken [build failed]
FAIL	github.com/vearutop/teststat/broken/deeper [build failed]
# github.com/vearutop/teststat/broken/tfatalf
# [github.com/vearutop/teststat/broken/tfatalf]
../../broken/tfatalf/t_test.go:10:11: non-constant format string in call to (*testing.common).Fatalf
FAIL	github.com/vearutop/teststat/broken/tfatalf [build failed]
```

</details>

<details>
<summary>Failed tests (including flaky): 3</summary>

<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestAlwaysFails</code></summary>

```
=== RUN   TestAlwaysFails
--- FAIL: TestAlwaysFails (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestAlwaysFailsInSubtest</code></summary>

```
=== RUN   TestAlwaysFailsInSubtest
--- FAIL: TestAlwaysFailsInSubtest (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestAlwaysFailsInSubtest//-&?\[]!@#$%^*()abc123_+=</code></summary>

```
=== RUN   TestAlwaysFailsInSubtest//-&?\[]!@#$%^*()abc123_+=
    --- FAIL: TestAlwaysFailsInSubtest//-&?\[]!@#$%^*()abc123_+= (0.00s)
panic: can't cope [recovered]
	panic: can't cope

goroutine 21 [running]:
testing.tRunner.func1.2({0x1030b4a80, 0x1030e5ba0})
	/opt/homebrew/opt/go/libexec/src/testing/testing.go:1734 +0x2bc
testing.tRunner.func1()
	/opt/homebrew/opt/go/libexec/src/testing/testing.go:1737 +0x47c
panic({0x1030b4a80?, 0x1030e5ba0?})
	/opt/homebrew/opt/go/libexec/src/runtime/panic.go:787 +0x124
github.com/vearutop/teststat/broken/other_test.TestAlwaysFailsInSubtest.func1(0xc00008b340?)
	/Users/vearutop/dev/teststat/broken/other/failed_test.go:21 +0x34
testing.tRunner(0xc00008b340, 0x1030e5168)
	/opt/homebrew/opt/go/libexec/src/testing/testing.go:1792 +0x184
created by testing.(*T).Run in goroutine 20
	/opt/homebrew/opt/go/libexec/src/testing/testing.go:1851 +0x688

```
</details>
</details>

<details>
<summary>Unfinished tests: 2</summary>

<details>
<summary><code>github.com/vearutop/teststat/broken/goroutine.TestThatPanicsInAGoroutine</code></summary>

```
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 4 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x34
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 3
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x44
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 7 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x34
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 6
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x44
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 22 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x34
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 21
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x44
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 22 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x34
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 21
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x44

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestThatPanics</code></summary>

```
=== RUN   TestThatPanics
=== PAUSE TestThatPanics
=== RUN   TestThatPanics
=== PAUSE TestThatPanics
=== RUN   TestThatPanics
=== PAUSE TestThatPanics
=== RUN   TestThatPanics
=== PAUSE TestThatPanics

```
</details>
</details>

### Metrics

```
pass: 0, fail: 12, unfinished: 2, total pkg: 5
```

Elapsed: 0s
Slow: 0s

### Test time distribution (seconds)
```
[ min  max] cnt total%  sum (12 events)
[0.00 0.00] 12 100.00% 0.00 ....................................................................................................

```
### Slowest test packages
<details>
<summary>Total packages with tests: 5</summary>

| Duration | Package |
| - | - |
| 1.179s | github.com/vearutop/teststat/broken/goroutine |
| 286ms | github.com/vearutop/teststat/broken/other |
| 0s | github.com/vearutop/teststat/broken |
| 0s | github.com/vearutop/teststat/broken/deeper |
| 0s | github.com/vearutop/teststat/broken/tfatalf |
</details>

