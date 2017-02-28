# Go live coding: testing, profiling

fetch url

```go
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

func get(rawurl string) ([]byte, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func main() {
	var (
		rawurl = flag.String("url", "http://example.com", "url to get")
	)
	flag.Parse()
	b, err := get(*rawurl)
	if err != nil {
		panic(err)
	}
	fmt.Printf("b = %s\n", string(b))
}
```

Let's scrape.

```go
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

func extract(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	re, err := regexp.Compile(`<meta name="description" content="([^"]*)" />`)
	if err != nil {
		return "", err
	}
	bb := re.FindSubmatch(b)
	return string(bb[1]), nil
}

func get(rawurl string) (string, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return extract(resp.Body)
}

func main() {
	var (
		rawurl = flag.String("url", "http://example.com", "url to get")
	)
	flag.Parse()
	desc, err := get(*rawurl)
	if err != nil {
		panic(err)
	}
	fmt.Printf("desc = %s\n", desc)
}
```

OK, add test

```go
package main

import (
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="description" content="test" />
		<title></title>
	</head>
	<body>
	</body>
	</html>
	`

	desc, err := extract(strings.NewReader(html))
	if err != nil {
		t.Fatalf("extract failed: %s", err)
	}
	if desc != "test" {
		t.Fatalf("want %s, got %s", "test", desc)
	}
}
```

regexp still failed ? make more flexible.

```go
	re, err := regexp.Compile(`<meta\s+name="description"\s+content="([^"]*)"\s*/>`)
```

```
-> % go test -v
=== RUN   TestExtract
--- PASS: TestExtract (0.00s)
PASS
ok      github.com/suzuken/talks/2017-livecoding        0.011s

```

## benchmark

isn't it slow? benchmark!

```go
var html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="description" content="test" />
	<title></title>
</head>
<body>
</body>
</html>
`

func BenchmarkExtract(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	bb := strings.NewReader(html)
	for i := 0; i < b.N; i++ {
		extract(bb)
	}
}
```

```
-> % go test -v -bench .
=== RUN   TestExtract
--- PASS: TestExtract (0.00s)
goos: darwin
goarch: amd64
pkg: github.com/suzuken/talks/2017-livecoding
BenchmarkExtract-4        100000             14330 ns/op           50208 B/op         66 allocs/op
PASS
ok      github.com/suzuken/talks/2017-livecoding        1.607s
```

profiling

    $ go test -v -bench . -cpuprofile prof.cpu


```
-> % go tool pprof 2017-livecoding.test prof.cpu
Entering interactive mode (type "help" for commands)
(pprof) top
2.11s of 2.21s total (95.48%)
Dropped 24 nodes (cum <= 0.01s)
Showing top 10 nodes out of 75 (cum >= 0.02s)
      flat  flat%   sum%        cum   cum%
     1.68s 76.02% 76.02%      1.68s 76.02%  runtime.kevent
     0.14s  6.33% 82.35%      0.14s  6.33%  runtime.mach_semaphore_signal
     0.13s  5.88% 88.24%      0.13s  5.88%  runtime.mach_semaphore_timedwait
     0.04s  1.81% 90.05%      0.04s  1.81%  runtime.usleep
     0.03s  1.36% 91.40%      0.03s  1.36%  runtime.mach_semaphore_wait
     0.02s   0.9% 92.31%      1.52s 68.78%  runtime.mallocgc
     0.02s   0.9% 93.21%      0.02s   0.9%  runtime.memclrNoHeapPointers
     0.02s   0.9% 94.12%      0.02s   0.9%  runtime.scanblock
     0.02s   0.9% 95.02%      0.02s   0.9%  runtime.stkbucket
     0.01s  0.45% 95.48%      0.02s   0.9%  regexp/syntax.(*compiler).cat
(pprof) list extract
Total: 2.21s
ROUTINE ======================== github.com/suzuken/talks/2017-livecoding.extract in /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo.go
         0      1.57s (flat, cum) 71.04% of Total
         .          .     12:func extract(r io.Reader) (string, error) {
         .          .     13:   b, err := ioutil.ReadAll(r)
         .          .     14:   if err != nil {
         .          .     15:           return "", err
         .          .     16:   }
         .      400ms     17:   re, err := regexp.Compile(`<meta\s+name="description"\s+content="([^"]*)"\s*/>`)
         .          .     18:   if err != nil {
         .          .     19:           return "", err
         .          .     20:   }
         .      1.17s     21:   bb := re.FindSubmatch(b)
         .          .     22:   if len(bb) <= 1 {
         .          .     23:           // not found
         .          .     24:           return "", nil
         .          .     25:   }
         .          .     26:   return string(bb[1]), nil

```


memory usage? memory profiling.

```
-> % go tool pprof --alloc_space 2017-livecoding.test prof.mem
Entering interactive mode (type "help" for commands)
(pprof) top10
5176.41MB of 5250.79MB total (98.58%)
Dropped 26 nodes (cum <= 26.25MB)
Showing top 10 nodes out of 23 (cum >= 4321.89MB)
      flat  flat%   sum%        cum   cum%
 4106.77MB 78.21% 78.21%  4106.77MB 78.21%  regexp.(*bitState).reset
  559.96MB 10.66% 88.88%   559.96MB 10.66%  regexp/syntax.(*compiler).rune
  205.62MB  3.92% 92.79%   205.62MB  3.92%  regexp.progMachine
  137.01MB  2.61% 95.40%   137.01MB  2.61%  regexp/syntax.(*parser).newLiteral
   57.03MB  1.09% 96.49%    57.03MB  1.09%  io/ioutil.readAll
   42.50MB  0.81% 97.30%    42.50MB  0.81%  regexp/syntax.(*parser).maybeConcat
   22.50MB  0.43% 97.73%       65MB  1.24%  regexp/syntax.(*parser).push
      22MB  0.42% 98.15%   272.02MB  5.18%  regexp/syntax.Parse
   13.50MB  0.26% 98.40%   870.99MB 16.59%  regexp.compile
    9.50MB  0.18% 98.58%  4321.89MB 82.31%  regexp.(*Regexp).doExecute

```

regexp.Compile use many memory. and FindSubmatch has more.

compile just once.

```
-> % go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/suzuken/talks/2017-livecoding
BenchmarkExtract-4       5000000               251 ns/op             512 B/op          1 allocs/op
PASS
ok      github.com/suzuken/talks/2017-livecoding        1.536s
```

20x faster!

mem profile again

```
-> % go test -run=^$ -bench . -cpuprofile prof.cpu -memprofile prof.mem
goos: darwin
goarch: amd64
pkg: github.com/suzuken/talks/2017-livecoding
BenchmarkExtract-4       5000000               250 ns/op             512 B/op          1 allocs/op
PASS
ok      github.com/suzuken/talks/2017-livecoding        1.537s

```

```
-> % go tool pprof --alloc_space 2017-livecoding.test prof.mem
Entering interactive mode (type "help" for commands)
(pprof) top --cum 30
2.86GB of 2.86GB total (  100%)
Dropped 4 nodes (cum <= 0.01GB)
      flat  flat%   sum%        cum   cum%
         0     0%     0%     2.86GB   100%  runtime.goexit
         0     0%     0%     2.86GB   100%  github.com/suzuken/talks/2017-livecoding.BenchmarkExtract
         0     0%     0%     2.86GB   100%  github.com/suzuken/talks/2017-livecoding.extract
         0     0%     0%     2.86GB   100%  io/ioutil.ReadAll
    2.86GB   100%   100%     2.86GB   100%  io/ioutil.readAll
         0     0%   100%     2.86GB   100%  testing.(*B).launch
         0     0%   100%     2.86GB   100%  testing.(*B).runN

```

## optimize

* avoid ioutil.ReadAll

use bufio.Scanner.

```go
func extract(r io.Reader) (string, error) {
	var desc []byte
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		bb := descRE.FindSubmatch(scanner.Bytes())
		if len(bb) <= 1 {
			continue
		}
		desc = bb[1]
		break
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return string(desc), nil
}
```


```
-> % go test -run=^$ -bench . -cpuprofile prof.cpu -memprofile prof.mem
goos: darwin
goarch: amd64
pkg: github.com/suzuken/talks/2017-livecoding
BenchmarkExtract-4       2000000               767 ns/op            4096 B/op          1 allocs/op
PASS
ok      github.com/suzuken/talks/2017-livecoding        2.310s

```

3x slower..

because,  `(*regexp).FindSubmatch` is slow by line.


how about using https://godoc.org/golang.org/x/net/html

like that.

```go
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"golang.org/x/net/html"
)

var descRE = regexp.MustCompile(`<meta\s+name="description"\s+content="([^"]*)"\s*/>`)

func extract(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}
	var desc string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			if isDescription(n.Attr) {
				for _, attr := range n.Attr {
					if attr.Key == "content" {
						desc = attr.Val
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return desc, nil
}

func isDescription(attrs []html.Attribute) bool {
	for _, attr := range attrs {
		if attr.Key == "name" && attr.Val == "description" {
			return true
		}
	}
	return false
}
```

test it.

```
-> % go test -v -bench . -cpuprofile prof.cpu -memprofile prof.mem
=== RUN   TestExtract
--- PASS: TestExtract (0.00s)
goos: darwin
goarch: amd64
pkg: github.com/suzuken/talks/2017-livecoding
BenchmarkExtract-4       1000000              1444 ns/op            4952 B/op          9 allocs/op
PASS
ok      github.com/suzuken/talks/2017-livecoding        1.481s
```

still slow?

```
-> % go tool pprof --alloc_space 2017-livecoding.test prof.mem
Entering interactive mode (type "help" for commands)
(pprof) to10
Error: unrecognized command: to10
(pprof) top10
4.61GB of 4.61GB total (  100%)
Dropped 5 nodes (cum <= 0.02GB)
Showing top 10 nodes out of 16 (cum >= 3.99GB)
      flat  flat%   sum%        cum   cum%
    3.99GB 86.62% 86.62%     3.99GB 86.62%  golang.org/x/net/html.NewTokenizerFragment
    0.32GB  6.86% 93.48%     0.34GB  7.46%  golang.org/x/net/html.(*parser).addElement
    0.27GB  5.90% 99.38%     4.61GB   100%  golang.org/x/net/html.Parse
    0.03GB  0.59%   100%     0.03GB  0.59%  golang.org/x/net/html.(*parser).addChild
         0     0%   100%     4.61GB   100%  github.com/suzuken/talks/2017-livecoding.BenchmarkExtract
         0     0%   100%     4.61GB   100%  github.com/suzuken/talks/2017-livecoding.extract
         0     0%   100%     0.34GB  7.46%  golang.org/x/net/html.(*parser).parse
         0     0%   100%     0.34GB  7.46%  golang.org/x/net/html.(*parser).parseCurrentToken
         0     0%   100%     0.34GB  7.46%  golang.org/x/net/html.(*parser).parseImpliedToken
         0     0%   100%     3.99GB 86.62%  golang.org/x/net/html.NewTokenizer
(pprof) list extract
Total: 4.61GB
ROUTINE ======================== github.com/suzuken/talks/2017-livecoding.extract in /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo.go
         0     4.61GB (flat, cum)   100% of Total
         .          .     11:)
         .          .     12:
         .          .     13:var descRE = regexp.MustCompile(`<meta\s+name="description"\s+content="([^"]*)"\s*/>`)
         .          .     14:
         .          .     15:func extract(r io.Reader) (string, error) {
         .     4.61GB     16:   doc, err := html.Parse(r)
         .          .     17:   if err != nil {
         .          .     18:           return "", err
         .          .     19:   }
         .          .     20:   var desc string
         .          .     21:   var f func(*html.Node)

```

html.Parse parses whole html string and build nodes tree. it's not fast.

## after all

for retriving description only, regexp is fast enough.

```go
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

var descRE = regexp.MustCompile(`<meta\s+name="description"\s+content="([^"]*)"\s*/>`)

func extract(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	bb := descRE.FindSubmatch(b)
	if len(bb) <= 1 {
		return "", nil
	}
	return string(bb[1]), nil
}

func get(rawurl string) (string, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return extract(resp.Body)
}

func main() {
	var (
		rawurl = flag.String("url", "http://example.com", "url to get")
	)
	flag.Parse()
	desc, err := get(*rawurl)
	if err != nil {
		panic(err)
	}
	fmt.Printf("desc = %s\n", desc)
}
```

consider

* how about retriving any other meta tag content ?

## http server / race detection

OK, let's create scraper as a service.

```go
func handler(w http.ResponseWriter, r *http.Request) {
	rawurl := r.URL.Query().Get("url")
	if rawurl == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}
	desc, err := get(rawurl)
	if err != nil {
		log.Printf("get error: %s", err)
		http.Error(w, "not found", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, desc)
}

func main() {
	http.HandleFunc("/", handler)
	log.Print("http server start listening on :8080")
	http.ListenAndServe(":8080", nil)
}
```

```
-> % go run demo.go
2017/02/28 14:29:07 http server start listening on :8080
2017/02/28 14:29:08 get error: Get aaa: unsupported protocol scheme ""

-> % curl "localhost:8080?url=https://voyagegroup.com"
        ALL    IRニュース    プレスリリース    パブリシティ    勉強会/登壇情報                 経営理念創業時からの想い「SOUL」と、価値観で&hellip;%
```

It works fine. Once scraped, you can cache it.

```go
var cache map[string]string

func init() {
	cache = make(map[string]string)
}

func get(rawurl string) (string, error) {
	if desc, ok := cache[rawurl]; ok {
		return desc, nil
	}
	resp, err := http.Get(rawurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	desc, err := extract(resp.Body)
	if err != nil {
		return "", err
	}
	cache[rawurl] = desc
	return desc, nil
}
```

It's safe? No. Go's map is not concurrent safe.

    go test -race

OK? Doesn't fail. Go's race detector has no false positive.
Write concurren tests.


```go
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlString)
}

func TestRace(t *testing.T) {
	// dummy server
	ts := httptest.NewServer(http.HandlerFunc(dummyHandler))
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := get(ts.URL); err != nil {
				t.Errorf("get failted: %s", err)
			}
		}()
	}
	wg.Wait()
}
```

it's failed.

```
-> % go test -race
==================
WARNING: DATA RACE
Write at 0x00c42006f290 by goroutine 10:
  runtime.mapassign()
      /Users/ke-suzuki/src/github.com/golang/go/src/runtime/hashmap.go:485 +0x0
  github.com/suzuken/talks/2017-livecoding.get()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo.go:45 +0x22a
  github.com/suzuken/talks/2017-livecoding.TestRace.func1()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:47 +0x83

Previous write at 0x00c42006f290 by goroutine 9:
  runtime.mapassign()
      /Users/ke-suzuki/src/github.com/golang/go/src/runtime/hashmap.go:485 +0x0
  github.com/suzuken/talks/2017-livecoding.get()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo.go:45 +0x22a
  github.com/suzuken/talks/2017-livecoding.TestRace.func1()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:47 +0x83

Goroutine 10 (running) created at:
  github.com/suzuken/talks/2017-livecoding.TestRace()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:45 +0xec
  testing.tRunner()
      /Users/ke-suzuki/src/github.com/golang/go/src/testing/testing.go:659 +0x10b

Goroutine 9 (finished) created at:
  github.com/suzuken/talks/2017-livecoding.TestRace()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:45 +0xec
  testing.tRunner()
      /Users/ke-suzuki/src/github.com/golang/go/src/testing/testing.go:659 +0x10b
==================
==================
WARNING: DATA RACE
Write at 0x00c4200c8f28 by goroutine 10:
  github.com/suzuken/talks/2017-livecoding.get()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo.go:45 +0x240
  github.com/suzuken/talks/2017-livecoding.TestRace.func1()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:47 +0x83

Previous write at 0x00c4200c8f28 by goroutine 9:
  github.com/suzuken/talks/2017-livecoding.get()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo.go:45 +0x240
  github.com/suzuken/talks/2017-livecoding.TestRace.func1()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:47 +0x83

Goroutine 10 (running) created at:
  github.com/suzuken/talks/2017-livecoding.TestRace()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:45 +0xec
  testing.tRunner()
      /Users/ke-suzuki/src/github.com/golang/go/src/testing/testing.go:659 +0x10b

Goroutine 9 (finished) created at:
  github.com/suzuken/talks/2017-livecoding.TestRace()
      /Users/ke-suzuki/src/github.com/suzuken/talks/2017-livecoding/demo_test.go:45 +0xec
  testing.tRunner()
      /Users/ke-suzuki/src/github.com/golang/go/src/testing/testing.go:659 +0x10b
==================
--- FAIL: TestRace (0.00s)
        testing.go:612: race detected during execution of test
FAIL
exit status 1
FAIL    github.com/suzuken/talks/2017-livecoding        0.022s

```

let's fix them.

```go
type cache struct {
	sync.RWMutex
	m map[string]string
}

func (c *cache) Add(k, v string) {
	defer c.Unlock()
	c.Lock()
	c.m[k] = v
}

func (c *cache) Get(k string) string {
	defer c.RUnlock()
	c.RLock()
	return c.m[k]
}

var globalCache cache

func init() {
	globalCache = cache{
		m: make(map[string]string),
	}
}

func get(rawurl string) (string, error) {
	if d := globalCache.Get(rawurl); d != "" {
		return d, nil
	}
	resp, err := http.Get(rawurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	desc, err := extract(resp.Body)
	if err != nil {
		return "", err
	}
	globalCache.Add(rawurl, desc)
	return desc, nil
}
```

passed!

```
-> % go test -v -race
=== RUN   TestExtract
--- PASS: TestExtract (0.00s)
=== RUN   TestRace
--- PASS: TestRace (0.00s)
PASS
ok      github.com/suzuken/talks/2017-livecoding        1.024s
```

## refactor

Next, avoid global variables `globalCache`.

```go
type server struct {
	cache *cache
}

func NewServer() *server {
	return &server{
		cache: NewCache(),
	}
}

func (s *server) get(rawurl string) (string, error) {
	if d := s.cache.Get(rawurl); d != "" {
		return d, nil
	}
	resp, err := http.Get(rawurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	desc, err := extract(resp.Body)
	if err != nil {
		return "", err
	}
	s.cache.Add(rawurl, desc)
	return desc, nil
}

func (s *server) handler(w http.ResponseWriter, r *http.Request) {
    ///
	desc, err := s.get(rawurl)
    ///...
}
```

and then,

```
-> % go test -v
=== RUN   TestExtract
--- PASS: TestExtract (0.00s)
=== RUN   TestRace
--- PASS: TestRace (0.00s)
PASS
ok      github.com/suzuken/talks/2017-livecoding        0.013s

```

passed!

## consider..

- when `<meta name="description"` commented, how does it works in `regexp` ?
