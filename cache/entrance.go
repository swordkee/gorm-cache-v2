package cache

import (
	"fmt"
	"github.com/swordkee/gorm-cache-v2/storage"
	"gorm.io/gorm"

	"github.com/swordkee/gorm-cache-v2/config"
)

func NewPlugin(opts ...Option) gorm.Plugin {
	cacheConfig := newCache(opts...)
	cache := &Gorm2Cache{
		Config: cacheConfig,
	}
	err := cache.Init()
	if err != nil {
		return nil
	}
	return cache
}
func newCache(opts ...Option) *config.CacheConfig {
	opt := new(config.CacheConfig)
	for _, f := range opts {
		f(opt)
	}
	if len(opts) == 0 {
		return &config.CacheConfig{
			CacheLevel:           config.CacheLevelAll,
			CacheStorage:         storage.NewMem(),
			InvalidateWhenUpdate: true,
			CacheTTL:             5000,
			CacheMaxItemCnt:      50,
		}
	}
	return &config.CacheConfig{
		CacheLevel:                     opt.CacheLevel,
		CacheStorage:                   opt.CacheStorage,
		Tables:                         opt.Tables,
		InvalidateWhenUpdate:           opt.InvalidateWhenUpdate,
		AsyncWrite:                     false,
		CacheTTL:                       opt.CacheTTL,
		CacheMaxItemCnt:                opt.CacheMaxItemCnt,
		DisableCachePenetrationProtect: false,
		DebugMode:                      opt.DebugMode,
		DebugLogger:                    opt.DebugLogger,
	}
}

func NewGorm2Cache(cacheConfig *config.CacheConfig) (Cache, error) {
	if cacheConfig == nil {
		return nil, fmt.Errorf("you pass a nil config")
	}
	cache := &Gorm2Cache{
		Config: cacheConfig,
		stats:  &stats{},
	}
	err := cache.Init()
	if err != nil {
		return nil, err
	}
	return cache, nil
}
