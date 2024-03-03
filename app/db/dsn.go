package db

import (
	"time"

	"github.com/go-sql-driver/mysql"
)

func getConfig() string {

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}

	dsn := mysql.Config{
		DBName:    "moneyhook",
		User:      "moneyhook",
		Passwd:    "password",
		Addr:      "sql:3306",
		Net:       "tcp",
		ParseTime: true,
		Collation: "utf8mb4_unicode_ci",
		Loc:       jst,
	}
	return dsn.FormatDSN()
}
