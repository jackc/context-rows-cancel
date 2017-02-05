// Test application demonstrating behavior of rows when ctx is canceled while reading rows.
//
// Actual output tested with lib/pq and jackc/pgx:
//
// $ DATABASE_ADAPTER=postgres DATABASE_URL=postgres://jack:jack@localhost/postgres go run main.go
// rows.Scan failed: sql: Rows are closed
// rowsRead: 28
// rows.Err(): <nil>
// $ DATABASE_ADAPTER=postgres DATABASE_URL=postgres://jack:jack@localhost/postgres go run main.go
// rowsRead: 30
// rows.Err(): <nil>
// $ DATABASE_ADAPTER=postgres DATABASE_URL=postgres://jack:jack@localhost/postgres go run main.go
// rows.Scan failed: sql: Rows are closed
// rowsRead: 20
// rows.Err(): <nil>
// $ DATABASE_ADAPTER=postgres DATABASE_URL=postgres://jack:jack@localhost/postgres go run main.go
// rowsRead: 120
// rows.Err(): <nil>
// $ DATABASE_ADAPTER=pgx DATABASE_URL=postgres://jack:jack@localhost/postgres go run main.go
// rowsRead: 204
// rows.Err(): <nil>
// $ DATABASE_ADAPTER=pgx DATABASE_URL=postgres://jack:jack@localhost/postgres go run main.go
// rowsRead: 94
// rows.Err(): <nil>
// $ DATABASE_ADAPTER=pgx DATABASE_URL=postgres://jack:jack@localhost/postgres go run main.go
// rowsRead: 135
// rows.Err(): <nil>
//
// Unexpected results:
//   * rows.Err() is always <nil> even though query was interrupted. How do I detect a partial read of result set?
//   * Undefined number of additional rows read after context cancelation
//   * Scan may or may not fail (this may be reasonable behavior)
package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	_ "github.com/lib/pq"
	"os"
)

func main() {
	db, err := sql.Open(os.Getenv("DATABASE_ADAPTER"), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql.Open failed: %v", err)
		os.Exit(1)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	rows, err := db.QueryContext(ctx, "select n from generate_series(1,10000000) n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "db.QueryContext failed: %v", err)
		os.Exit(1)
	}

	rowsRead := 0
	for rows.Next() {
		var n int64
		err = rows.Scan(&n)
		if err != nil {
			fmt.Println("rows.Scan failed:", err)
		}

		rowsRead++
		if rowsRead == 2 {
			cancelFn()
		}
	}

	fmt.Println("rowsRead:", rowsRead)
	fmt.Println("rows.Err():", rows.Err())
}
