package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sync"
)

type server struct {
	cache *cache
}

func NewServer() *server {
	return &server{
		cache: NewCache(),
	}
}

type cache struct {
	sync.RWMutex
	m map[string]string
}

func NewCache() *cache {
	return &cache{
		m: make(map[string]string),
	}
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
	rawurl := r.URL.Query().Get("url")
	if rawurl == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}
	desc, err := s.get(rawurl)
	if err != nil {
		log.Printf("get error: %s", err)
		http.Error(w, "not found", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, desc)
}

func main() {
	s := server{}
	http.HandleFunc("/", s.handler)
	log.Print("http server start listening on :8080")
	http.ListenAndServe(":8080", nil)
}
