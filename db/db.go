package db

import (
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gucooing/BaPs/config"
	"github.com/gucooing/BaPs/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gromlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var SQL *gorm.DB

func NewPE(cfg *config.DB) {
	switch cfg.DbType {
	case "sqlite":
		SQL = NewSqlite(cfg.Dsn)
	case "mysql":
		SQL = NewMysql(cfg.Dsn)
	default:
		log.Panicln("数据库的类型只支持 'sqlite' 和 'mysql' ")
		return
	}

	SQL.AutoMigrate(
		&YostarAccount{},
		&YostarUser{},
		&YostarUserLogin{},
		&BlackDevice{},
		&YostarGame{},
		&YostarMail{},
		&YostarFriend{},
		&YostarClan{},
	)
	logger.Info("数据库连接成功")
}

func NewMysql(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gromlogger.Default.LogMode(gromlogger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		logger.Error("mysql connect err:", err)
		panic(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("mysql connect err:", err)
		panic(err.Error())
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(100)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(10000)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(100 * time.Millisecond) // 0.1 秒

	return db
}

func NewSqlite(dsn string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gromlogger.Default.LogMode(gromlogger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		logger.Error("mysql connect err:", err)
		panic(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("mysql connect err:", err)
		panic(err.Error())
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(100)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(10000)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(100 * time.Millisecond) // 0.1 秒

	return db
}
