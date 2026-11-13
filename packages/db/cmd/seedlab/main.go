// Command seedlab applies LabSeedSQL using DATABASE_URL (default: local compose DSN).
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://bambufarmer:bambufarmer@127.0.0.1:5432/bambufarmer?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := db.Open(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	if err := db.SeedLab(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("seedlab: ok")
}
