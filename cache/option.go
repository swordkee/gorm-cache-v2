package cache

import (
	"github.com/swordkee/gorm-cache-v2/config"
	"github.com/swordkee/gorm-cache-v2/storage"
	"github.com/swordkee/gorm-cache-v2/util"
)

type Option func(p *config.CacheConfig)

func WithLevel(level config.CacheLevel) Option {
	return func(p *config.CacheConfig) {
		p.CacheLevel = level
	}
}
func WithStorage(storage storage.DataStorage) Option {
	return func(p *config.CacheConfig) {
		p.CacheStorage = storage
	}
}
func WithTables(tables []string) Option {
	return func(p *config.CacheConfig) {
		p.Tables = tables
	}
}
func WithInvalidateWhenUpdate(isBool bool) Option {
	return func(p *config.CacheConfig) {
		p.InvalidateWhenUpdate = isBool
	}
}
func WithAsyncWrite(isBool bool) Option {
	return func(p *config.CacheConfig) {
		p.AsyncWrite = isBool
	}
}
func WithCacheTTL(ttl int64) Option {
	return func(p *config.CacheConfig) {
		p.CacheTTL = ttl
	}
}
func WithCacheMaxItemCnt(cnt int64) Option {
	return func(p *config.CacheConfig) {
		p.CacheMaxItemCnt = cnt
	}
}
func WithDisableCachePenetrationProtect(isBool bool) Option {
	return func(p *config.CacheConfig) {
		p.DisableCachePenetrationProtect = isBool
	}
}

func WithDebugMode(debug bool) Option {
	return func(p *config.CacheConfig) {
		p.DebugMode = debug
	}
}
func WithDebugLogger(log util.LoggerInterface) Option {
	return func(p *config.CacheConfig) {
		p.DebugLogger = log
	}
}
