package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Func func(key string) (interface{}, error)

type result struct {
	value interface{}
	err error
}

type entry struct {
	res result
	ready chan struct {}
}

type Demo struct {
	f Func
	mu sync.Mutex
	cache map[string]*entry
}

func NewDemo(f Func) *Demo {
	return &Demo {
		f: f,
		cache : make(map[string]*entry),
	}
}

func (demo *Demo) Get (key string) (interface{}, error) {
	demo.mu.Lock()
	e := demo.cache[key]

	if e == nil {
		e = &entry{
			ready: make(chan struct{}),
		}
		demo.cache[key] = e
		demo.mu.Unlock()

		e.res.value, e.res.err = demo.f(key)
		close(e.ready)
	} else {
		demo.mu.Unlock()
		<- e.ready
	}
	return e.res.value, e.res.err
}

func httpGetBody(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println(url , " has some thing wrong.")
	}
	return ioutil.ReadAll(resp.Body)
}

func main() {
	m := NewDemo(httpGetBody)
	urls := []string{"https://www.zhihu.com", "https://www.zhihu.com", "https://www.baidu.com", "http://www.windsky.com", "https://www.baidu.com"}
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			start := time.Now()
			_, err := m.Get(url)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("%s, %s, \n",
				url, time.Since(start))
			wg.Done()
		} (url)
	}
	time.Sleep(3 * time.Second)
	fmt.Println("\nTRY AGAIN\n") // 能够更加清楚的了解缓存之后查询速度的优化
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			start := time.Now()
			_, err := m.Get(url)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("%s, %s, \n", url, time.Since(start))
			wg.Done()
		} (url)
	}
	wg.Wait()
}