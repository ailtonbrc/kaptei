package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type privacidadePostgres struct{ db *sql.DB }

func NewPrivacidadePostgres(db *sql.DB) ports.PrivacidadeRepository {
	return &privacidadePostgres{db: db}
}

func (r *privacidadePostgres) ObterContaPorSlug(ctx context.Context, slug string) (string, error) {
	var contaID string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM contas_saas WHERE LOWER(slug_publico)=LOWER($1) AND site_publicado=TRUE`, slug).Scan(&contaID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return contaID, err
}

func (r *privacidadePostgres) Criar(ctx context.Context, solicitacao *domain.SolicitacaoTitular) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar solicitação do titular: %w", err)
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `INSERT INTO solicitacoes_titular
		(conta_id,protocolo,tipo,nome_protegido,email_hash,telefone_hash,contato_protegido,detalhes_protegidos,status,prazo_resposta_em)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'RECEBIDA',$9)
		RETURNING id,criado_em,atualizado_em`, solicitacao.ContaID, solicitacao.Protocolo, solicitacao.Tipo,
		solicitacao.NomeProtegido, solicitacao.EmailHash, solicitacao.TelefoneHash, solicitacao.ContatoProtegido,
		solicitacao.DetalhesProtegidos, solicitacao.PrazoRespostaEm,
	).Scan(&solicitacao.ID, &solicitacao.CriadoEm, &solicitacao.AtualizadoEm)
	if err != nil {
		return fmt.Errorf("criar solicitação do titular: %w", err)
	}
	if err := registrarEventoTitular(ctx, tx, solicitacao.ContaID, solicitacao.ID, nil, "RECEBIDA", "Solicitação recebida pelo canal público"); err != nil {
		return err
	}
	return tx.Commit()
}

const camposSolicitacaoTitular = `id,conta_id,protocolo,tipo,nome_protegido,email_hash,telefone_hash,
	contato_protegido,detalhes_protegidos,status,prazo_resposta_em,identidade_verificada_em,metodo_verificacao,
	decisao,fundamento_legal,observacao_decisao,decidida_em,executada_em,criado_em,atualizado_em`

func escanearSolicitacaoTitular(scanner interface{ Scan(...any) error }) (*domain.SolicitacaoTitular, error) {
	var s domain.SolicitacaoTitular
	err := scanner.Scan(&s.ID, &s.ContaID, &s.Protocolo, &s.Tipo, &s.NomeProtegido, &s.EmailHash, &s.TelefoneHash,
		&s.ContatoProtegido, &s.DetalhesProtegidos, &s.Status, &s.PrazoRespostaEm, &s.IdentidadeVerificadaEm,
		&s.MetodoVerificacao, &s.Decisao, &s.FundamentoLegal, &s.ObservacaoDecisao, &s.DecididaEm, &s.ExecutadaEm,
		&s.CriadoEm, &s.AtualizadoEm)
	return &s, err
}

func (r *privacidadePostgres) Listar(ctx context.Context, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.SolicitacaoTitular], error) {
	deslocamento := (filtro.Pagina - 1) * filtro.Limite
	condicao := `conta_id=$1 AND ($2='' OR protocolo ILIKE '%'||$2||'%') AND ($3='' OR status=$3) AND ($4='' OR tipo=$4)`
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM solicitacoes_titular WHERE `+condicao, contaID, filtro.Busca, filtro.Status, filtro.Tipo).Scan(&total); err != nil {
		return nil, fmt.Errorf("contar solicitações do titular: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+camposSolicitacaoTitular+` FROM solicitacoes_titular WHERE `+condicao+`
		ORDER BY criado_em DESC,id DESC LIMIT $5 OFFSET $6`, contaID, filtro.Busca, filtro.Status, filtro.Tipo, filtro.Limite, deslocamento)
	if err != nil {
		return nil, fmt.Errorf("listar solicitações do titular: %w", err)
	}
	defer rows.Close()
	dados := make([]*domain.SolicitacaoTitular, 0)
	for rows.Next() {
		s, err := escanearSolicitacaoTitular(rows)
		if err != nil {
			return nil, fmt.Errorf("ler solicitação do titular: %w", err)
		}
		dados = append(dados, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &domain.ListaPaginada[*domain.SolicitacaoTitular]{Dados: dados, Total: total, Pagina: filtro.Pagina, Limite: filtro.Limite}, nil
}

func (r *privacidadePostgres) Obter(ctx context.Context, id, contaID string) (*domain.SolicitacaoTitular, error) {
	s, err := escanearSolicitacaoTitular(r.db.QueryRowContext(ctx, `SELECT `+camposSolicitacaoTitular+` FROM solicitacoes_titular WHERE id=$1 AND conta_id=$2`, id, contaID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,tipo,descricao,usuario_id,criado_em FROM eventos_solicitacao_titular
		WHERE solicitacao_id=$1 AND conta_id=$2 ORDER BY criado_em,id`, id, contaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	s.Eventos = make([]domain.EventoSolicitacaoTitular, 0)
	for rows.Next() {
		var evento domain.EventoSolicitacaoTitular
		if err := rows.Scan(&evento.ID, &evento.Tipo, &evento.Descricao, &evento.UsuarioID, &evento.CriadoEm); err != nil {
			return nil, err
		}
		s.Eventos = append(s.Eventos, evento)
	}
	return s, rows.Err()
}

func (r *privacidadePostgres) VerificarIdentidade(ctx context.Context, id, contaID, usuarioID, metodo, evidenciaProtegida string) error {
	return r.alterarComEvento(ctx, contaID, id, &usuarioID, "IDENTIDADE_VERIFICADA", "Identidade do titular verificada", func(tx *sql.Tx) error {
		resultado, err := tx.ExecContext(ctx, `UPDATE solicitacoes_titular SET identidade_verificada_em=now(),identidade_verificada_por=$1,
			metodo_verificacao=$2,evidencia_verificacao_protegida=$3,status='EM_ANALISE',atualizado_em=now()
			WHERE id=$4 AND conta_id=$5 AND status IN ('RECEBIDA','VALIDANDO_IDENTIDADE','EM_ANALISE')`, usuarioID, metodo, evidenciaProtegida, id, contaID)
		return exigirLinhaAlterada(resultado, err, "solicitação não está disponível para verificação")
	})
}

func (r *privacidadePostgres) Decidir(ctx context.Context, id, contaID, usuarioID, decisao, fundamento, observacao string) error {
	return r.alterarComEvento(ctx, contaID, id, &usuarioID, "DECISAO_REGISTRADA", "Decisão "+decisao+" registrada", func(tx *sql.Tx) error {
		resultado, err := tx.ExecContext(ctx, `UPDATE solicitacoes_titular SET decisao=$1,fundamento_legal=$2,observacao_decisao=$3,
			decidida_em=now(),decidida_por=$4,status=$1,atualizado_em=now()
			WHERE id=$5 AND conta_id=$6 AND identidade_verificada_em IS NOT NULL AND status='EM_ANALISE'`, decisao, fundamento, observacao, usuarioID, id, contaID)
		return exigirLinhaAlterada(resultado, err, "solicitação precisa ter identidade verificada e estar em análise")
	})
}

func (r *privacidadePostgres) GerarDadosExportacao(ctx context.Context, contaID, email, telefone string) (*domain.DadosTitularPersistidos, error) {
	dados := &domain.DadosTitularPersistidos{}
	consultas := []struct {
		destino *[]map[string]any
		query   string
	}{
		{&dados.Clientes, `SELECT COALESCE(jsonb_agg(to_jsonb(c) ORDER BY c.criado_em),'[]'::jsonb) FROM clientes c WHERE c.conta_id=$1 AND (($2<>'' AND LOWER(COALESCE(c.email,''))=$2) OR ($3<>'' AND regexp_replace(COALESCE(c.telefone,''),'\D','','g')=$3))`},
		{&dados.Leads, `SELECT COALESCE(jsonb_agg(to_jsonb(l) ORDER BY l.criado_em),'[]'::jsonb) FROM leads l WHERE l.conta_id=$1 AND (($2<>'' AND LOWER(COALESCE(l.email,''))=$2) OR ($3<>'' AND regexp_replace(COALESCE(l.telefone,''),'\D','','g')=$3))`},
		{&dados.Interacoes, `SELECT COALESCE(jsonb_agg(to_jsonb(i) ORDER BY i.criado_em),'[]'::jsonb) FROM interacoes i WHERE i.conta_id=$1 AND i.cliente_id IN (SELECT c.id FROM clientes c WHERE c.conta_id=$1 AND (($2<>'' AND LOWER(COALESCE(c.email,''))=$2) OR ($3<>'' AND regexp_replace(COALESCE(c.telefone,''),'\D','','g')=$3)))`},
		{&dados.Agendamentos, `SELECT COALESCE(jsonb_agg(to_jsonb(a) ORDER BY a.created_at),'[]'::jsonb) FROM agendamentos a WHERE a.conta_id=$1 AND a.cliente_id IN (SELECT c.id FROM clientes c WHERE c.conta_id=$1 AND (($2<>'' AND LOWER(COALESCE(c.email,''))=$2) OR ($3<>'' AND regexp_replace(COALESCE(c.telefone,''),'\D','','g')=$3)))`},
		{&dados.Conversas, `SELECT COALESCE(jsonb_agg(to_jsonb(c) ORDER BY c.criado_em),'[]'::jsonb) FROM conversas_whatsapp c WHERE c.conta_id=$1 AND (($3<>'' AND regexp_replace(c.numero_contato,'\D','','g')=$3) OR c.lead_id IN (SELECT l.id FROM leads l WHERE l.conta_id=$1 AND (($2<>'' AND LOWER(COALESCE(l.email,''))=$2) OR ($3<>'' AND regexp_replace(COALESCE(l.telefone,''),'\D','','g')=$3))))`},
		{&dados.Mensagens, `SELECT COALESCE(jsonb_agg(to_jsonb(m) ORDER BY m.ocorrida_em),'[]'::jsonb) FROM mensagens_whatsapp m WHERE m.conta_id=$1 AND m.conversa_id IN (SELECT c.id FROM conversas_whatsapp c WHERE c.conta_id=$1 AND (($3<>'' AND regexp_replace(c.numero_contato,'\D','','g')=$3) OR c.lead_id IN (SELECT l.id FROM leads l WHERE l.conta_id=$1 AND (($2<>'' AND LOWER(COALESCE(l.email,''))=$2) OR ($3<>'' AND regexp_replace(COALESCE(l.telefone,''),'\D','','g')=$3)))))`},
	}
	for _, consulta := range consultas {
		var bruto []byte
		if err := r.db.QueryRowContext(ctx, consulta.query, contaID, email, telefone).Scan(&bruto); err != nil {
			return nil, fmt.Errorf("consultar dados do titular: %w", err)
		}
		if err := json.Unmarshal(bruto, consulta.destino); err != nil {
			return nil, fmt.Errorf("decodificar dados do titular: %w", err)
		}
	}
	return dados, nil
}

func (r *privacidadePostgres) ConcluirExportacao(ctx context.Context, id, contaID, usuarioID string) error {
	return r.alterarComEvento(ctx, contaID, id, &usuarioID, "EXPORTACAO_GERADA", "Exportação segura gerada para entrega ao titular", func(tx *sql.Tx) error {
		resultado, err := tx.ExecContext(ctx, `UPDATE solicitacoes_titular SET status='CONCLUIDA',executada_em=now(),executada_por=$1,atualizado_em=now()
			WHERE id=$2 AND conta_id=$3 AND status='APROVADA'`, usuarioID, id, contaID)
		return exigirLinhaAlterada(resultado, err, "solicitação não está aprovada")
	})
}

func (r *privacidadePostgres) ExecutarDireito(ctx context.Context, id, contaID, usuarioID, email, telefone string, tipo domain.TipoSolicitacaoTitular) error {
	return r.alterarComEvento(ctx, contaID, id, &usuarioID, "DIREITO_EXECUTADO", "Tratamento aprovado foi executado", func(tx *sql.Tx) error {
		var tipoPersistido domain.TipoSolicitacaoTitular
		if err := tx.QueryRowContext(ctx, `SELECT tipo FROM solicitacoes_titular WHERE id=$1 AND conta_id=$2 AND status='APROVADA' AND identidade_verificada_em IS NOT NULL FOR UPDATE`, id, contaID).Scan(&tipoPersistido); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("solicitação não está aprovada")
			}
			return err
		}
		if tipoPersistido != tipo {
			return errors.New("tipo da solicitação diverge da operação")
		}
		predicadoCliente := `conta_id=$1 AND (($2<>'' AND LOWER(COALESCE(email,''))=$2) OR ($3<>'' AND regexp_replace(COALESCE(telefone,''),'\D','','g')=$3))`
		predicadoLead := predicadoCliente
		switch tipo {
		case domain.TipoExclusao:
			if _, err := tx.ExecContext(ctx, `DELETE FROM conversas_whatsapp WHERE conta_id=$1 AND (($3<>'' AND regexp_replace(numero_contato,'\D','','g')=$3) OR lead_id IN (SELECT id FROM leads WHERE `+predicadoLead+`))`, contaID, email, telefone); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM agendamentos WHERE conta_id=$1 AND cliente_id IN (SELECT id FROM clientes WHERE `+predicadoCliente+`)`, contaID, email, telefone); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM leads WHERE `+predicadoLead, contaID, email, telefone); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM clientes WHERE `+predicadoCliente, contaID, email, telefone); err != nil {
				return err
			}
		case domain.TipoAnonimizacao:
			if _, err := tx.ExecContext(ctx, `DELETE FROM conversas_whatsapp WHERE conta_id=$1 AND (($3<>'' AND regexp_replace(numero_contato,'\D','','g')=$3) OR lead_id IN (SELECT id FROM leads WHERE `+predicadoLead+`))`, contaID, email, telefone); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE agendamentos SET titulo='Atendimento anonimizado',descricao=NULL,updated_at=now() WHERE conta_id=$1 AND cliente_id IN (SELECT id FROM clientes WHERE `+predicadoCliente+`)`, contaID, email, telefone); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE leads SET nome='Titular anonimizado',email=NULL,telefone=NULL,mensagem=NULL,motivo_descarte=NULL,consentimento_lgpd=FALSE,consentimento_lgpd_em=NULL,consentimento_lgpd_versao=NULL,atualizado_em=now() WHERE `+predicadoLead, contaID, email, telefone); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE clientes SET nome='Titular anonimizado',email=NULL,telefone=NULL,notas=NULL,cpf=NULL,data_nascimento=NULL,estado_civil=NULL,financeiro='{}'::jsonb,origem_utm='{}'::jsonb,tags='[]'::jsonb,preferencias='{}'::jsonb,atualizado_em=now() WHERE `+predicadoCliente, contaID, email, telefone); err != nil {
				return err
			}
		case domain.TipoRevogacao:
			if _, err := tx.ExecContext(ctx, `UPDATE conversas_whatsapp SET consentimento_marketing=FALSE,consentimento_marketing_em=NULL,consentimento_marketing_origem='REVOGACAO_TITULAR',consentimento_marketing_evidencia=NULL,atualizado_em=now() WHERE conta_id=$1 AND (($3<>'' AND regexp_replace(numero_contato,'\D','','g')=$3) OR lead_id IN (SELECT id FROM leads WHERE `+predicadoLead+`))`, contaID, email, telefone); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE leads SET consentimento_lgpd=FALSE,consentimento_lgpd_em=NULL,consentimento_lgpd_versao=NULL,atualizado_em=now() WHERE `+predicadoLead, contaID, email, telefone); err != nil {
				return err
			}
		case domain.TipoCorrecao, domain.TipoBloqueio:
			// A execução confirma que o operador concluiu o tratamento manual no CRM.
		default:
			return errors.New("tipo não permite execução por esta operação")
		}
		resultado, err := tx.ExecContext(ctx, `UPDATE solicitacoes_titular SET status='CONCLUIDA',executada_em=now(),executada_por=$1,atualizado_em=now() WHERE id=$2 AND conta_id=$3 AND status='APROVADA'`, usuarioID, id, contaID)
		return exigirLinhaAlterada(resultado, err, "solicitação não está aprovada")
	})
}

func (r *privacidadePostgres) alterarComEvento(ctx context.Context, contaID, id string, usuarioID *string, tipo, descricao string, alterar func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := alterar(tx); err != nil {
		return err
	}
	if err := registrarEventoTitular(ctx, tx, contaID, id, usuarioID, tipo, descricao); err != nil {
		return err
	}
	return tx.Commit()
}

func registrarEventoTitular(ctx context.Context, tx *sql.Tx, contaID, solicitacaoID string, usuarioID *string, tipo, descricao string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO eventos_solicitacao_titular (conta_id,solicitacao_id,usuario_id,tipo,descricao) VALUES ($1,$2,$3,$4,$5)`, contaID, solicitacaoID, usuarioID, tipo, descricao)
	return err
}

func exigirLinhaAlterada(resultado sql.Result, err error, mensagem string) error {
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return errors.New(mensagem)
	}
	return nil
}
