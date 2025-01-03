package db

import (
	"fmt"
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

func getPostgresConfig() string{
	dbName := common.GetEnv("POSTGRES_DATABASE", "")
	user := common.GetEnv("POSTGRES_USER", "")
	password := common.GetEnv("POSTGRES_PASSWORD", "")
	host := common.GetEnv("POSTGRES_HOST", "")
	port := common.GetEnv("POSTGRES_PORT", "5432")
	timezone := "Asia/Tokyo"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=%s",
		user, password, host, port, dbName,timezone,
	)

	return dsn
}
