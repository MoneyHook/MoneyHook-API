package db

import (
	"time"

	common "MoneyHook/MoneyHook-API/common"

	"github.com/go-sql-driver/mysql"
)

func getConfig() string {

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}

	dsn := mysql.Config{
		DBName:               common.GetEnv("MYSQL_DATABASE", ""),
		User:                 common.GetEnv("MYSQL_USER", ""),
		Passwd:               common.GetEnv("MYSQL_PASSWORD", ""),
		Addr:                 common.GetEnv("MYSQL_HOST", ""),
		Net:                  common.GetEnv("NET", ""),
		ParseTime:            true,
		Collation:            "utf8mb4_unicode_ci",
		Loc:                  jst,
		AllowNativePasswords: true,
	}
	return dsn.FormatDSN()
}
