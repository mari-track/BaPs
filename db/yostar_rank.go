package db

import (
	"fmt"
	"log"

	"github.com/gucooing/BaPs/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type YostarRank struct {
	SeasonId int64
	Uid      int64   `gorm:"unique"`
	Score    float64 // 分数
}

func UserTable(x int64) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Table(fmt.Sprintf("raid_rank_%v", x))
	}
}

func NewYostarRank(cfg *config.DB) *gorm.DB {
	switch cfg.DbType {
	case "sqlite":
		return NewSqlite(cfg.Dsn)
	case "mysql":
		return NewMysql(cfg.Dsn)
	default:
		log.Panicln("数据库的类型只支持 'sqlite' 和 'mysql' ")
	}

	return nil
}

// UpAllYostarRank 批量覆盖保存排名数据
func UpAllYostarRank(yostarRankSql *gorm.DB, x []*YostarRank, seasonId int64) error {
	err := yostarRankSql.Clauses(clause.OnConflict{
		UpdateAll: true, // 如果主键冲突,则更新所有字段
	}).Scopes(UserTable(seasonId)).CreateInBatches(&x, 200).Error
	return err
}

// GetAllYostarRank 拉取全部排名数据
func GetAllYostarRank(yostarRankSql *gorm.DB, seasonId int64) []*YostarRank {
	r := make([]*YostarRank, 0)
	yostarRankSql.Scopes(UserTable(seasonId)).Model(&YostarRank{}).Find(&r)
	return r
}
