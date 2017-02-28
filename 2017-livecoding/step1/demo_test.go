package main

import (
	"strings"
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

func BenchmarkExtract(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	bb := strings.NewReader(htmlString)
	for i := 0; i < b.N; i++ {
		extract(bb)
	}
}
