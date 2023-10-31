package test

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/swordkee/gorm-cache-v2/storage"
	"os"
	"testing"

	"gorm.io/gorm/logger"

	"github.com/swordkee/gorm-cache-v2/cache"

	"github.com/swordkee/gorm-cache-v2/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	username     = "root"
	password     = "Zcydf741205,."
	databaseName = "site_reldb"
	ip           = "localhost"
	port         = "3306"
)
var (
	searchCache  cache.Cache
	primaryCache cache.Cache
	allCache     cache.Cache

	searchDB   *gorm.DB
	primaryDB  *gorm.DB
	allDB      *gorm.DB
	originalDB *gorm.DB
)

var (
	testSize = 200 // minimum 200
)

func TestMain(m *testing.M) {
	log("test setup ...")

	var err error
	//logger.Default.LogMode(logger.Info)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, ip, port, databaseName)
	originalDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		CreateBatchSize: 1000,
		Logger:          logger.Default,
	})
	if err != nil {
		log("open db error: %v", err)
		os.Exit(-1)
	}

	searchDB, err = forkDB(originalDB)
	if err != nil {
		log("open db error: %v", err)
		os.Exit(-1)
	}

	primaryDB, err = forkDB(originalDB)
	if err != nil {
		log("open db error: %v", err)
		os.Exit(-1)
	}

	allDB, err = forkDB(originalDB)
	if err != nil {
		log("open db error: %v", err)
		os.Exit(-1)
	}

	searchCache, err = cache.NewGorm2Cache(&config.CacheConfig{
		CacheLevel:           config.CacheLevelOnlySearch,
		CacheStorage:         storage.NewGcache(gcache.New(1000)),
		InvalidateWhenUpdate: true,
		CacheTTL:             5000,
		CacheMaxItemCnt:      5000,
		DebugMode:            false,
	})
	if err != nil {
		log("setup search cache error: %v", err)
		os.Exit(-1)
	}

	primaryCache, err = cache.NewGorm2Cache(&config.CacheConfig{
		CacheLevel:           config.CacheLevelOnlyPrimary,
		CacheStorage:         storage.NewGcache(gcache.New(1000)),
		InvalidateWhenUpdate: true,
		CacheTTL:             5000,
		CacheMaxItemCnt:      5000,
		DebugMode:            false,
	})
	if err != nil {
		log("setup primary cache error: %v", err)
		os.Exit(-1)
	}

	allCache, err = cache.NewGorm2Cache(&config.CacheConfig{
		CacheLevel:           config.CacheLevelAll,
		CacheStorage:         storage.NewGcache(gcache.New(1000)),
		InvalidateWhenUpdate: true,
		CacheTTL:             5000,
		CacheMaxItemCnt:      5000,
		DebugMode:            false,
	})
	if err != nil {
		log("setup all cache error: %v", err)
		os.Exit(-1)
	}

	primaryDB.Use(primaryCache)
	searchDB.Use(searchCache)
	allDB.Use(allCache)
	// primaryCache.AttachToDB(primaryDB)+
	// searchCache.AttachToDB(searchDB)
	// allCache.AttachToDB(allDB)

	err = timer("prepare table and data", func() error {
		return PrepareTableAndData(originalDB)
	})
	if err != nil {
		log("setup table and data error: %v", err)
		os.Exit(-1)
	}

	result := m.Run()

	err = timer("clean table and data", func() error {
		return CleanTable(originalDB)
	})
	if err != nil {
		log("clean table and data error: %v", err)
		os.Exit(-1)
	}

	log("integration test end.")
	os.Exit(result)
}

func forkDB(db *gorm.DB) (newDB *gorm.DB, err error) {
	plugins := map[string]gorm.Plugin{}
	for k, v := range db.Config.Plugins {
		plugins[k] = v
	}
	newDB, err = gorm.Open(db.Dialector, &gorm.Config{
		SkipDefaultTransaction:                   db.Config.SkipDefaultTransaction,
		NamingStrategy:                           db.Config.NamingStrategy,
		FullSaveAssociations:                     db.Config.FullSaveAssociations,
		Logger:                                   db.Config.Logger,
		NowFunc:                                  db.Config.NowFunc,
		DryRun:                                   db.Config.DryRun,
		PrepareStmt:                              db.Config.PrepareStmt,
		DisableAutomaticPing:                     db.Config.DisableAutomaticPing,
		DisableForeignKeyConstraintWhenMigrating: db.Config.DisableForeignKeyConstraintWhenMigrating,
		IgnoreRelationshipsWhenMigrating:         db.Config.IgnoreRelationshipsWhenMigrating,
		DisableNestedTransaction:                 db.Config.DisableNestedTransaction,
		AllowGlobalUpdate:                        db.Config.AllowGlobalUpdate,
		QueryFields:                              db.Config.QueryFields,
		CreateBatchSize:                          db.Config.CreateBatchSize,
		ClauseBuilders:                           db.Config.ClauseBuilders,
		ConnPool:                                 db.Config.ConnPool,
		Plugins:                                  plugins,
	})
	return
}
