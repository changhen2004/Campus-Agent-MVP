package mysql

import (
	"fmt"

	"campus-agent/internal/config"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func BuildDSN(cfg config.MySQLConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}

func Open(cfg config.MySQLConfig) (*gorm.DB, error) {
	return gorm.Open(gormmysql.Open(BuildDSN(cfg)), &gorm.Config{})
}
