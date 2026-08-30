package bancodados

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func AbrirPostgres(databaseURL string) (*sql.DB, error) {
	return AbrirPostgresConfigurado(databaseURL, ConfiguracaoPool{MaximoAbertas: 25, MaximoOciosas: 10, VidaMaxima: 30 * time.Minute, OciosidadeMaxima: 5 * time.Minute})
}

type ConfiguracaoPool struct {
	MaximoAbertas    int
	MaximoOciosas    int
	VidaMaxima       time.Duration
	OciosidadeMaxima time.Duration
}

func AbrirPostgresConfigurado(databaseURL string, configuracao ConfiguracaoPool) (*sql.DB, error) {
	banco, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("abrir conexão PostgreSQL: %w", err)
	}

	banco.SetMaxOpenConns(configuracao.MaximoAbertas)
	banco.SetMaxIdleConns(configuracao.MaximoOciosas)
	banco.SetConnMaxLifetime(configuracao.VidaMaxima)
	banco.SetConnMaxIdleTime(configuracao.OciosidadeMaxima)

	contexto, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	if err := banco.PingContext(contexto); err != nil {
		_ = banco.Close()
		return nil, fmt.Errorf("conectar ao PostgreSQL: %w", err)
	}

	return banco, nil
}
