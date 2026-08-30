package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/adapters/repositories"
	"github.com/msdev/kaptei/internal/plataforma/bancodados"
	"github.com/msdev/kaptei/internal/plataforma/configuracao"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := executar(); err != nil {
		fmt.Fprintln(os.Stderr, "Bootstrap recusado:", err)
		os.Exit(1)
	}
}

func executar() error {
	nome := flag.String("nome", "", "Nome completo do primeiro superadministrador")
	email := flag.String("email", "", "E-mail do primeiro superadministrador")
	flag.Parse()
	*nome = strings.TrimSpace(*nome)
	*email = strings.ToLower(strings.TrimSpace(*email))
	endereco, err := mail.ParseAddress(*email)
	if *nome == "" || len([]rune(*nome)) > 120 || err != nil || !strings.EqualFold(endereco.Address, *email) {
		return errors.New("informe nome e e-mail válidos")
	}
	senha, err := gerarSenhaTemporaria()
	if err != nil {
		return errors.New("não foi possível gerar a senha temporária")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("não foi possível proteger a senha temporária")
	}
	databaseURL, err := configuracao.CarregarDatabaseURL()
	if err != nil {
		return err
	}
	banco, err := bancodados.AbrirPostgres(databaseURL)
	if err != nil {
		return err
	}
	defer banco.Close()
	ctx, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()
	if err := repositories.NewBootstrapPostgres(banco).CriarPrimeiroSuperAdmin(ctx, *nome, *email, string(hash)); err != nil {
		return err
	}
	fmt.Println("Superadministrador criado. Guarde a senha temporária e altere-a no primeiro acesso:")
	fmt.Println(senha)
	return nil
}

func gerarSenhaTemporaria() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
