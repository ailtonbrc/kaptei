.PHONY: run build migrateup migratedown tidy validar-database-url

# Variáveis do banco de dados (pode ser sobrescrito pelo .env)
validar-database-url:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL nao configurada" && exit 1)

# Baixar pacotes e organizar o go.mod
tidy:
	go mod tidy

# Executa as migrations de criação
migrateup: validar-database-url
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose up

# Executa as migrations de reversão
migratedown: validar-database-url
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose down

# Roda o projeto
run:
	go run cmd/api/main.go

# Compila o projeto
build:
	go build -o bin/api cmd/api/main.go
