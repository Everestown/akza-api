package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "akza2024"
	if len(os.Args) > 1 {
		password = os.Args[1]
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Hash for %q:\n%s\n", password, string(hash))

	// If DATABASE_URL provided — update admin in DB directly
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("\nTo update DB: set DATABASE_URL and run again, or run the SQL above manually.")
		return
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	email := "admin@akza.ru"
	if len(os.Args) > 2 {
		email = os.Args[2]
	}

	res, err := db.Exec(
		"UPDATE admins SET password_hash = $1 WHERE email = $2",
		string(hash), email,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update error: %v\n", err)
		os.Exit(1)
	}
	rows, _ := res.RowsAffected()
	fmt.Printf("Updated %d row(s) for %s\n", rows, email)
}
