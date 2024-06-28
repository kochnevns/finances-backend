package imcache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type IMCache struct {
	*cache.Cache
}

func NewIMCache() *IMCache {
	c := IMCache{}
	c.Cache = cache.New(5*time.Minute, 10*time.Minute)

	return &c
}
