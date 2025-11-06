module github.com/gookit/miglite/cmd/miglite

go 1.24.0

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/gookit/miglite v1.0.0
	github.com/lib/pq v1.10.9
	modernc.org/sqlite v1.40.0
)

replace github.com/gookit/miglite => ../../
