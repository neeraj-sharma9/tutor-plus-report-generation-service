package utility

import "fmt"

func GetPostgresDSN(host string, port int32, dbname, user, password string) string {
	dsn := fmt.Sprintf("host=%s port=%v dbname=%s user=%s password=%s sslmode=disable", host,
		port, dbname, user, password)
	return dsn
}
