package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type executorSQL interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type outboxPostgres struct{ db *sql.DB }

func NewOutboxPostgres(db *sql.DB) ports.OutboxRepository {
	return &outboxPostgres{db: db}
}

func inserirEventoOutbox(ctx context.Context, executor executorSQL, evento *domain.EventoOutbox) error {
	if evento == nil {
		return errors.New("evento de outbox obrigatório")
	}
	err := executor.QueryRowContext(ctx, `INSERT INTO eventos_outbox
		(conta_id,tipo,payload_protegido,chave_idempotencia,status,tentativas,maximo_tentativas,disponivel_em,criado_em)
		VALUES ($1,$2,$3,$4,'PENDENTE',0,$5,$6,$7)
		ON CONFLICT (chave_idempotencia) DO NOTHING RETURNING id`,
		evento.ContaID, evento.Tipo, evento.PayloadProtegido, evento.ChaveIdempotencia,
		evento.MaximoTentativas, evento.DisponivelEm, evento.CriadoEm,
	).Scan(&evento.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("evento de outbox duplicado ou não persistido")
	}
	if err != nil {
		return fmt.Errorf("persistir evento de outbox: %w", err)
	}
	return nil
}

func (r *outboxPostgres) Reservar(
	ctx context.Context,
	trabalhadorID string,
	limite int,
	duracaoBloqueio time.Duration,
) ([]*domain.EventoOutbox, error) {
	rows, err := r.db.QueryContext(ctx, `WITH candidatos AS (
		SELECT id
		FROM eventos_outbox
		WHERE tentativas < maximo_tentativas
		  AND ((status='PENDENTE' AND disponivel_em<=NOW())
		       OR (status='PROCESSANDO' AND bloqueado_ate<NOW()))
		ORDER BY disponivel_em, criado_em
		FOR UPDATE SKIP LOCKED
		LIMIT $1
	)
	UPDATE eventos_outbox evento
	SET status='PROCESSANDO', tentativas=tentativas+1,
		bloqueado_por=$2, bloqueado_ate=NOW()+($3 * INTERVAL '1 millisecond')
	FROM candidatos
	WHERE evento.id=candidatos.id
	RETURNING evento.id,evento.conta_id,evento.tipo,evento.payload_protegido,
		evento.chave_idempotencia,evento.status,evento.tentativas,evento.maximo_tentativas,
		evento.disponivel_em,evento.bloqueado_ate,evento.bloqueado_por,evento.criado_em`,
		limite, trabalhadorID, duracaoBloqueio.Milliseconds(),
	)
	if err != nil {
		return nil, fmt.Errorf("reservar eventos de outbox: %w", err)
	}
	defer rows.Close()
	eventos := make([]*domain.EventoOutbox, 0, limite)
	for rows.Next() {
		evento := &domain.EventoOutbox{}
		if err := rows.Scan(
			&evento.ID, &evento.ContaID, &evento.Tipo, &evento.PayloadProtegido,
			&evento.ChaveIdempotencia, &evento.Status, &evento.Tentativas, &evento.MaximoTentativas,
			&evento.DisponivelEm, &evento.BloqueadoAte, &evento.BloqueadoPor, &evento.CriadoEm,
		); err != nil {
			return nil, fmt.Errorf("ler evento reservado: %w", err)
		}
		eventos = append(eventos, evento)
	}
	return eventos, rows.Err()
}

func (r *outboxPostgres) Concluir(ctx context.Context, eventoID, trabalhadorID string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE eventos_outbox
		SET status='CONCLUIDO', processado_em=NOW(), bloqueado_ate=NULL, bloqueado_por=NULL, ultimo_erro=NULL
		WHERE id=$1 AND status='PROCESSANDO' AND bloqueado_por=$2`, eventoID, trabalhadorID)
	if err != nil {
		return fmt.Errorf("concluir evento de outbox: %w", err)
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("reserva do evento de outbox não está mais ativa")
	}
	return nil
}

func (r *outboxPostgres) Falhar(
	ctx context.Context,
	eventoID, trabalhadorID, mensagem string,
	proximaTentativa time.Time,
	definitivo bool,
) error {
	status := "PENDENTE"
	if definitivo {
		status = "FALHOU"
	}
	mensagem = strings.TrimSpace(mensagem)
	if len([]rune(mensagem)) > 1000 {
		mensagem = string([]rune(mensagem)[:1000])
	}
	resultado, err := r.db.ExecContext(ctx, `UPDATE eventos_outbox
		SET status=$3, disponivel_em=$4, bloqueado_ate=NULL, bloqueado_por=NULL, ultimo_erro=$5
		WHERE id=$1 AND status='PROCESSANDO' AND bloqueado_por=$2`,
		eventoID, trabalhadorID, status, proximaTentativa, mensagem,
	)
	if err != nil {
		return fmt.Errorf("registrar falha do evento de outbox: %w", err)
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("reserva do evento de outbox não está mais ativa")
	}
	return nil
}
