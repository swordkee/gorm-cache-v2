package cache

import (
	"errors"
	"fmt"
	"sync"

	"github.com/asjdf/gorm-cache/config"
	"github.com/asjdf/gorm-cache/util"
	"gorm.io/gorm"
)

func AfterQuery(cache *Gorm2Cache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context
		sqlObj, _ := db.InstanceGet("gorm:cache:sql")
		sql := sqlObj.(string)
		varObj, _ := db.InstanceGet("gorm:cache:vars")
		vars := varObj.([]interface{})

		if db.Error == nil {
			// error is nil -> cache not hit, we cache newly retrieved data
			primaryKeys, objects := getObjectsAfterLoad(db)

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()

				if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlySearch {
					// cache search data
					if int64(len(objects)) > cache.Config.CacheMaxItemCnt {
						return
					}

					cache.Logger.CtxInfo(ctx, "[AfterQuery] start to set search cache for sql: %s", sql)
					cacheBytes, err := json.Marshal(db.Statement.Dest)
					if err != nil {
						cache.Logger.CtxError(ctx, "[AfterQuery] cannot marshal cache for sql: %s, not cached", sql)
						return
					}
					cache.Logger.CtxInfo(ctx, "[AfterQuery] set cache: %v", string(cacheBytes))
					err = cache.SetSearchCache(ctx, fmt.Sprintf("%d|", db.RowsAffected)+string(cacheBytes), tableName, sql, vars...)
					if err != nil {
						cache.Logger.CtxError(ctx, "[AfterQuery] set search cache for sql: %s error: %v", sql, err)
						return
					}
					cache.Logger.CtxInfo(ctx, "[AfterQuery] sql %s cached", sql)
				}
			}()

			go func() {
				defer wg.Done()

				if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlyPrimary {
					// cache primary cache data
					if len(primaryKeys) != len(objects) {
						return
					}
					if int64(len(objects)) > cache.Config.CacheMaxItemCnt {
						cache.Logger.CtxInfo(ctx, "[AfterQuery] objects length is more than max item count, not cached")
						return
					}
					kvs := make([]util.Kv, 0, len(objects))
					for i := 0; i < len(objects); i++ {
						jsonStr, err := json.Marshal(objects[i])
						if err != nil {
							cache.Logger.CtxError(ctx, "[AfterQuery] object %v cannot marshal, not cached", objects[i])
							continue
						}
						kvs = append(kvs, util.Kv{
							Key:   primaryKeys[i],
							Value: string(jsonStr),
						})
					}
					cache.Logger.CtxInfo(ctx, "[AfterQuery] start to set primary cache for kvs: %+v", kvs)
					err := cache.BatchSetPrimaryKeyCache(ctx, tableName, kvs)
					if err != nil {
						cache.Logger.CtxError(ctx, "[AfterQuery] batch set primary key cache for key %v error: %v",
							primaryKeys, err)
					}
				}
			}()
			if !cache.Config.AsyncWrite {
				wg.Wait()
			}
			return
		}

		if !cache.Config.DisableCachePenetrationProtect {
			if errors.Is(db.Error, gorm.ErrRecordNotFound) { // 应对缓存穿透 未来可能考虑使用其他过滤器实现：如布隆过滤器
				cache.Logger.CtxInfo(ctx, "[AfterQuery] set cache: %v", "recordNotFound")
				err := cache.SetSearchCache(ctx, "recordNotFound", tableName, sql, vars...)
				if err != nil {
					cache.Logger.CtxError(ctx, "[AfterQuery] set search cache for sql: %s error: %v", sql, err)
					return
				}
				cache.Logger.CtxInfo(ctx, "[AfterQuery] sql %s cached", sql)
			}
		}
		if errors.Is(db.Error, util.RecordNotFoundCacheHit) {
			db.Error = gorm.ErrRecordNotFound
			cache.IncrHitCount()
			return
		}

		if errors.Is(db.Error, util.SearchCacheHit) {
			// search cache hit
			db.Error = nil
			cache.IncrHitCount()
			return
		}

		if errors.Is(db.Error, util.PrimaryCacheHit) {
			// primary cache hit
			db.Error = nil
			cache.IncrHitCount()
			return
		}
	}
}
