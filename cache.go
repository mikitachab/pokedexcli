package main

import (
	"sync"
	"time"
)

type Cache struct {
	cache    map[string][]byte
	mu       *sync.RWMutex
	interval time.Duration
}

func (c *Cache) Add(url string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[url] = data

	go func(url string) {
		timer := time.NewTimer(c.interval)
		<-timer.C
		c.Delete(url)
	}(url)
}

func (c *Cache) Delete(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, url)
}

func (c *Cache) Get(url string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, ok := c.cache[url]

	if !ok {
		return nil, false
	}

	return data, true

}

func NewCache(interval time.Duration) *Cache {
	return &Cache{
		cache:    make(map[string][]byte),
		mu:       &sync.RWMutex{},
		interval: interval,
	}
}
