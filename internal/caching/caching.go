package caching

import (
	"time"
)

type cacher interface {
	Set(k string, x interface{}, d time.Duration)
	Get(k string) (interface{}, bool)
	Delete(k string)
}

type Caching struct {
	c                 cacher
	defaultExpiration time.Duration
}

func New(c cacher) *Caching {
	return &Caching{
		c: c,
	}
}

func (c *Caching) Set(key string, value interface{}) {
	c.c.Set(key, value, c.defaultExpiration)
}

func (c *Caching) Get(key string) (interface{}, bool) {
	return c.c.Get(key)
}

func (c *Caching) Delete(key string) {
	c.c.Delete(key)
}

func (c *Caching) GenerateKey(parts ...string) string {
	key := ""
	for _, p := range parts {
		key += p + ":"
	}

	return key
}
