//go:build ferramentas_manuais

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL nÃ£o configurada")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("ALTER TABLE imoveis ADD COLUMN IF NOT EXISTS usuario_id UUID REFERENCES usuarios(id) ON DELETE CASCADE;")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Success: usuario_id column added to imoveis table.")
}
