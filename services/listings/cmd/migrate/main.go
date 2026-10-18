// Command migrate applies pending auth database migrations and exits.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/hacuba/listings/internal/dbmigrate"
)

func main() {
	directory := flag.String("dir", "/migrations", "directory containing ordered .sql migration files")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	applied, err := dbmigrate.Apply(ctx, os.Getenv("DATABASE_URL"), *directory)
	if err != nil {
		log.Fatalf("listings migrate: %v", err)
	}
	log.Printf("listings migrate: applied %d migration(s)", applied)
}
