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
