package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

var htmlString = `
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

func TestExtract(t *testing.T) {
	desc, err := extract(strings.NewReader(htmlString))
	if err != nil {
		t.Fatalf("extract failed: %s", err)
	}
	if desc != "test" {
		t.Fatalf("want %s, got %s", "test", desc)
	}
}

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

func BenchmarkExtract(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	bb := strings.NewReader(htmlString)
	for i := 0; i < b.N; i++ {
		extract(bb)
	}
}
