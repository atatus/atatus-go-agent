module go.atatus.com/agent/module/atsql

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jackc/pgx/v4 v4.9.0
	github.com/lib/pq v1.3.0
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/stretchr/testify v1.6.1
	go.atatus.com/agent v1.2.0
)

replace go.atatus.com/agent => ../..

go 1.13
