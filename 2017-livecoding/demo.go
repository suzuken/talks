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
