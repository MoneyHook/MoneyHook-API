package db

import (
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func getConfig() string {

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}

	dsn := mysql.Config{
		DBName:               os.Getenv("MYSQL_DATABASE"),
		User:                 os.Getenv("MYSQL_USER"),
		Passwd:               os.Getenv("MYSQL_PASSWORD"),
		Addr:                 os.Getenv("MYSQL_HOST"),
		Net:                  os.Getenv("NET"),
		ParseTime:            true,
		Collation:            "utf8mb4_unicode_ci",
		Loc:                  jst,
		AllowNativePasswords: true,
	}
	return dsn.FormatDSN()
}
