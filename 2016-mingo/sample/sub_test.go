// original: https://github.com/golang/proposal/blob/master/design/12166-subtests.md
package sample

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSum(t *testing.T) {
	tests := []struct {
		A, B int
		Sum  int
	}{
		{1, 2, 3},
		{1, 1, 2},
		{2, 1, 3},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprint(tc.A, "+", tc.B), func(t *testing.T) {
			if got := tc.A + tc.B; got != tc.Sum {
				t.Errorf("got %d; want %d", got, tc.Sum)
			}
		})
	}
}

func BenchmarkX(b *testing.B) {
	b.Run("n=3", func(b *testing.B) {})
	b.Run("n=10", func(b *testing.B) {})
	b.Run("n=100", func(b *testing.B) {})
}

func cat(ss ...string) string {
	var r string
	for _, s := range ss {
		r += s
	}
	return r
}

func buf(ss ...string) string {
	var b bytes.Buffer
	for _, s := range ss {
		b.WriteString(s)
	}
	return b.String()
}

func seed(n int) []string {
	s := make([]string, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, "a")
	}
	return s
}

func bench(b *testing.B, n int, f func(...string) string) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f(seed(n)...)
	}
}

func BenchmarkConcatenate(b *testing.B) {
	benchCases := []struct {
		name string
		n    int
		f    func(...string) string
	}{
		{"Cat", 3, cat},
		{"Buf", 3, buf},
		{"Cat", 100, cat},
		{"Buf", 100, buf},
		{"Cat", 10000, cat},
		{"Buf", 10000, buf},
	}
	for _, c := range benchCases {
		b.Run(fmt.Sprintf("%s%d", c.name, c.n),
			func(b *testing.B) { bench(b, c.n, c.f) })
	}
}
