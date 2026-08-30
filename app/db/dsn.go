package db

import (
	"fmt"

	common "MoneyHook/MoneyHook-API/common"
)

func getPostgresConfig() string {
	return getPostgresConfigForDatabase(common.GetEnv("POSTGRES_DATABASE", ""))
}

func getPostgresAdminConfig() string {
	return getPostgresConfigForDatabase("postgres")
}

func getPostgresConfigForDatabase(dbName string) string {
	user := common.GetEnv("POSTGRES_USER", "")
	password := common.GetEnv("POSTGRES_PASSWORD", "")
	host := common.GetEnv("POSTGRES_HOST", "")
	port := common.GetEnv("POSTGRES_PORT", "")
	sslmode := common.GetEnv("SSLMODE", "")
	timezone := "Asia/Tokyo"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s&connect_timeout=5",
		user, password, host, port, dbName, sslmode, timezone,
	)

	return dsn
}
