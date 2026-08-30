package main

import (
	"database/sql"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/msdev/kaptei/internal/plataforma/configuracao"
)

func main() {
	databaseURL, err := configuracao.CarregarDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}

	banco, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("abrir PostgreSQL: %v", err)
	}
	defer banco.Close()

	driver, err := postgres.WithInstance(banco, &postgres.Config{})
	if err != nil {
		log.Fatalf("preparar driver de migrations: %v", err)
	}

	migrador, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("preparar migrations: %v", err)
	}
	defer migrador.Close()

	if err := migrador.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("aplicar migrations: %v", err)
	}
	log.Println("migrations aplicadas com sucesso")
}
