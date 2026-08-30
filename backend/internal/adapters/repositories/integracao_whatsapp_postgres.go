package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type integracaoWhatsAppPostgres struct{ db *sql.DB }

func NewIntegracaoWhatsAppPostgres(db *sql.DB) ports.IntegracaoWhatsAppRepository {
	return &integracaoWhatsAppPostgres{db: db}
}

func (r *integracaoWhatsAppPostgres) ObterPorConta(ctx context.Context, contaID string) (*domain.ConfiguracaoWhatsApp, error) {
	return r.obter(ctx, `SELECT id,conta_id,waba_id,numero_telefone_id,numero_exibicao,token_acesso_protegido,ativa,criado_em,atualizado_em
		FROM integracoes_whatsapp_cloud WHERE conta_id=$1`, contaID)
}

func (r *integracaoWhatsAppPostgres) ObterPorNumeroTelefone(ctx context.Context, numeroTelefoneID string) (*domain.ConfiguracaoWhatsApp, error) {
	return r.obter(ctx, `SELECT id,conta_id,waba_id,numero_telefone_id,numero_exibicao,token_acesso_protegido,ativa,criado_em,atualizado_em
		FROM integracoes_whatsapp_cloud WHERE numero_telefone_id=$1 AND ativa=true`, numeroTelefoneID)
}

func (r *integracaoWhatsAppPostgres) obter(ctx context.Context, consulta, parametro string) (*domain.ConfiguracaoWhatsApp, error) {
	configuracao := &domain.ConfiguracaoWhatsApp{}
	err := r.db.QueryRowContext(ctx, consulta, parametro).Scan(
		&configuracao.ID, &configuracao.ContaID, &configuracao.WABAID,
		&configuracao.NumeroTelefoneID, &configuracao.NumeroExibicao,
		&configuracao.TokenAcessoProtegido, &configuracao.Ativa,
		&configuracao.CriadoEm, &configuracao.AtualizadoEm,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("obter configuração WhatsApp: %w", err)
	}
	configuracao.TokenAcessoConfigurado = configuracao.TokenAcessoProtegido != ""
	return configuracao, nil
}

func (r *integracaoWhatsAppPostgres) Salvar(ctx context.Context, configuracao *domain.ConfiguracaoWhatsApp) error {
	err := r.db.QueryRowContext(ctx, `INSERT INTO integracoes_whatsapp_cloud
		(conta_id,waba_id,numero_telefone_id,numero_exibicao,token_acesso_protegido,ativa,criado_em,atualizado_em)
		VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW())
		ON CONFLICT (conta_id) DO UPDATE SET waba_id=EXCLUDED.waba_id,
			numero_telefone_id=EXCLUDED.numero_telefone_id,numero_exibicao=EXCLUDED.numero_exibicao,
			token_acesso_protegido=EXCLUDED.token_acesso_protegido,ativa=EXCLUDED.ativa,atualizado_em=NOW()
		RETURNING id,criado_em,atualizado_em`, configuracao.ContaID, configuracao.WABAID,
		configuracao.NumeroTelefoneID, configuracao.NumeroExibicao,
		configuracao.TokenAcessoProtegido, configuracao.Ativa,
	).Scan(&configuracao.ID, &configuracao.CriadoEm, &configuracao.AtualizadoEm)
	if err != nil {
		var erroPostgres *pq.Error
		if errors.As(err, &erroPostgres) && erroPostgres.Code == "23505" {
			return errors.New("este número do WhatsApp já está vinculado a outra conta")
		}
		return fmt.Errorf("salvar configuração WhatsApp: %w", err)
	}
	configuracao.TokenAcessoConfigurado = true
	return nil
}

func (r *integracaoWhatsAppPostgres) Enfileirar(ctx context.Context, eventos []*domain.EventoIntegracao) error {
	if len(eventos) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar fila WhatsApp: %w", err)
	}
	defer tx.Rollback()
	for _, evento := range eventos {
		if evento == nil {
			continue
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO eventos_integracao
			(conta_id,provedor,tipo,identificador_externo,pagina_id,payload_protegido,status,tentativas,
			 maximo_tentativas,disponivel_em,criado_em)
			VALUES ($1,$2,$3,$4,$5,$6,'PENDENTE',0,$7,$8,$9)
			ON CONFLICT (provedor,identificador_externo) DO NOTHING`, evento.ContaID, evento.Provedor,
			evento.Tipo, evento.IdentificadorExterno, evento.PaginaID, evento.PayloadProtegido,
			evento.MaximoTentativas, evento.DisponivelEm, evento.CriadoEm)
		if err != nil {
			return fmt.Errorf("enfileirar mensagem WhatsApp: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar fila WhatsApp: %w", err)
	}
	return nil
}

func (r *integracaoWhatsAppPostgres) Reservar(ctx context.Context, trabalhadorID string, limite int, duracaoBloqueio time.Duration) ([]*domain.EventoIntegracao, error) {
	rows, err := r.db.QueryContext(ctx, `WITH candidatos AS (
		SELECT id FROM eventos_integracao
		WHERE provedor='WHATSAPP' AND tipo='MENSAGEM_RECEBIDA' AND tentativas < maximo_tentativas
		  AND ((status='PENDENTE' AND disponivel_em<=NOW()) OR (status='PROCESSANDO' AND bloqueado_ate<NOW()))
		ORDER BY disponivel_em,criado_em FOR UPDATE SKIP LOCKED LIMIT $1
	)
	UPDATE eventos_integracao evento SET status='PROCESSANDO',tentativas=tentativas+1,
		bloqueado_por=$2,bloqueado_ate=NOW()+($3 * INTERVAL '1 millisecond')
	FROM candidatos WHERE evento.id=candidatos.id
	RETURNING evento.id,evento.conta_id,evento.provedor,evento.tipo,evento.identificador_externo,
		evento.pagina_id,evento.status,evento.tentativas,evento.maximo_tentativas,evento.disponivel_em,
		evento.bloqueado_ate,evento.bloqueado_por,evento.criado_em,evento.payload_protegido`,
		limite, trabalhadorID, duracaoBloqueio.Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("reservar mensagens WhatsApp: %w", err)
	}
	defer rows.Close()
	eventos := make([]*domain.EventoIntegracao, 0, limite)
	for rows.Next() {
		evento := &domain.EventoIntegracao{}
		if err := rows.Scan(&evento.ID, &evento.ContaID, &evento.Provedor, &evento.Tipo,
			&evento.IdentificadorExterno, &evento.PaginaID, &evento.Status, &evento.Tentativas,
			&evento.MaximoTentativas, &evento.DisponivelEm, &evento.BloqueadoAte,
			&evento.BloqueadoPor, &evento.CriadoEm, &evento.PayloadProtegido); err != nil {
			return nil, fmt.Errorf("ler mensagem WhatsApp reservada: %w", err)
		}
		eventos = append(eventos, evento)
	}
	return eventos, rows.Err()
}

func (r *integracaoWhatsAppPostgres) RegistrarMensagem(ctx context.Context, contaID string, mensagem *domain.MensagemWhatsAppRecebida, conteudoProtegido, chaveLead string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar registro da mensagem WhatsApp: %w", err)
	}
	defer tx.Rollback()
	janelaAte := mensagem.OcorridaEm.Add(24 * time.Hour)
	var conversaID string
	err = tx.QueryRowContext(ctx, `INSERT INTO conversas_whatsapp
		(conta_id,lead_id,numero_contato,nome_contato,janela_atendimento_ate,ultima_mensagem_em,criado_em,atualizado_em)
		VALUES ($1,(SELECT id FROM leads WHERE conta_id=$1 AND chave_idempotencia=$2::uuid),$3,NULLIF($4,''),$5,$6,NOW(),NOW())
		ON CONFLICT (conta_id,numero_contato) DO UPDATE SET
			lead_id=COALESCE(conversas_whatsapp.lead_id,EXCLUDED.lead_id),
			nome_contato=COALESCE(EXCLUDED.nome_contato,conversas_whatsapp.nome_contato),
			janela_atendimento_ate=GREATEST(conversas_whatsapp.janela_atendimento_ate,EXCLUDED.janela_atendimento_ate),
			ultima_mensagem_em=GREATEST(conversas_whatsapp.ultima_mensagem_em,EXCLUDED.ultima_mensagem_em),atualizado_em=NOW()
		RETURNING id`, contaID, chaveLead, mensagem.NumeroContato,
		mensagem.NomeContato, janelaAte, mensagem.OcorridaEm).Scan(&conversaID)
	if err != nil {
		return fmt.Errorf("registrar conversa WhatsApp: %w", err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO mensagens_whatsapp
		(conta_id,conversa_id,identificador_externo,direcao,tipo,conteudo_protegido,status,ocorrida_em,criado_em,atualizado_em)
		VALUES ((SELECT conta_id FROM conversas_whatsapp WHERE id=$1),$1,$2,'ENTRADA',$3,$4,'RECEBIDA',$5,NOW(),NOW())
		ON CONFLICT (identificador_externo) DO NOTHING`, conversaID, mensagem.IdentificadorExterno,
		mensagem.Tipo, conteudoProtegido, mensagem.OcorridaEm)
	if err != nil {
		return fmt.Errorf("registrar mensagem WhatsApp: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar mensagem WhatsApp: %w", err)
	}
	return nil
}

func (r *integracaoWhatsAppPostgres) Concluir(ctx context.Context, eventoID, trabalhadorID string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE eventos_integracao SET status='CONCLUIDO',processado_em=NOW(),
		bloqueado_ate=NULL,bloqueado_por=NULL,ultimo_erro=NULL WHERE id=$1 AND status='PROCESSANDO' AND bloqueado_por=$2`, eventoID, trabalhadorID)
	return validarAtualizacaoEventoIntegracao(resultado, err, "concluir")
}

func (r *integracaoWhatsAppPostgres) Falhar(ctx context.Context, eventoID, trabalhadorID, mensagem string, proximaTentativa time.Time, definitivo bool) error {
	status := "PENDENTE"
	if definitivo {
		status = "FALHOU"
	}
	mensagem = strings.TrimSpace(mensagem)
	if len([]rune(mensagem)) > 1000 {
		mensagem = string([]rune(mensagem)[:1000])
	}
	resultado, err := r.db.ExecContext(ctx, `UPDATE eventos_integracao SET status=$3,disponivel_em=$4,
		bloqueado_ate=NULL,bloqueado_por=NULL,ultimo_erro=$5 WHERE id=$1 AND status='PROCESSANDO' AND bloqueado_por=$2`,
		eventoID, trabalhadorID, status, proximaTentativa, mensagem)
	return validarAtualizacaoEventoIntegracao(resultado, err, "registrar falha de")
}

func (r *integracaoWhatsAppPostgres) ObterConversa(ctx context.Context, conversaID, contaID string) (*domain.ConversaWhatsApp, error) {
	conversa := &domain.ConversaWhatsApp{}
	err := r.db.QueryRowContext(ctx, `SELECT id,conta_id,lead_id,numero_contato,nome_contato,
		consentimento_marketing,janela_atendimento_ate,ultima_mensagem_em
		FROM conversas_whatsapp WHERE id=$1 AND conta_id=$2`, conversaID, contaID).Scan(
		&conversa.ID, &conversa.ContaID, &conversa.LeadID, &conversa.NumeroContato, &conversa.NomeContato,
		&conversa.ConsentimentoMarketing, &conversa.JanelaAtendimentoAte, &conversa.UltimaMensagemEm)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("obter conversa WhatsApp: %w", err)
	}
	return conversa, nil
}

func (r *integracaoWhatsAppPostgres) ListarConversas(ctx context.Context, contaID string, usuarioID *string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.ConversaWhatsApp], error) {
	busca := "%" + filtro.Busca + "%"
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM conversas_whatsapp conversa
		LEFT JOIN leads lead ON lead.id=conversa.lead_id AND lead.conta_id=conversa.conta_id
		WHERE conversa.conta_id=$1 AND ($2::uuid IS NULL OR lead.usuario_id=$2)
		  AND ($3='' OR conversa.numero_contato ILIKE $4 OR COALESCE(conversa.nome_contato,'') ILIKE $4)`,
		contaID, usuarioID, filtro.Busca, busca).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("contar conversas WhatsApp: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT conversa.id,conversa.conta_id,conversa.lead_id,conversa.numero_contato,
		conversa.nome_contato,conversa.consentimento_marketing,conversa.janela_atendimento_ate,conversa.ultima_mensagem_em
		FROM conversas_whatsapp conversa
		LEFT JOIN leads lead ON lead.id=conversa.lead_id AND lead.conta_id=conversa.conta_id
		WHERE conversa.conta_id=$1 AND ($2::uuid IS NULL OR lead.usuario_id=$2)
		  AND ($3='' OR conversa.numero_contato ILIKE $4 OR COALESCE(conversa.nome_contato,'') ILIKE $4)
		ORDER BY conversa.ultima_mensagem_em DESC LIMIT $5 OFFSET $6`,
		contaID, usuarioID, filtro.Busca, busca, filtro.Limite, (filtro.Pagina-1)*filtro.Limite)
	if err != nil {
		return nil, fmt.Errorf("listar conversas WhatsApp: %w", err)
	}
	defer rows.Close()
	dados := make([]*domain.ConversaWhatsApp, 0, filtro.Limite)
	for rows.Next() {
		conversa := &domain.ConversaWhatsApp{}
		if err := rows.Scan(&conversa.ID, &conversa.ContaID, &conversa.LeadID, &conversa.NumeroContato,
			&conversa.NomeContato, &conversa.ConsentimentoMarketing, &conversa.JanelaAtendimentoAte, &conversa.UltimaMensagemEm); err != nil {
			return nil, fmt.Errorf("ler conversa WhatsApp: %w", err)
		}
		dados = append(dados, conversa)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer conversas WhatsApp: %w", err)
	}
	return &domain.ListaPaginada[*domain.ConversaWhatsApp]{Dados: dados, Total: total, Pagina: filtro.Pagina, Limite: filtro.Limite}, nil
}

func (r *integracaoWhatsAppPostgres) ListarMensagens(ctx context.Context, conversaID, contaID string, filtro domain.FiltroPaginacao) (*domain.ListaPaginada[*domain.MensagemWhatsApp], error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM mensagens_whatsapp WHERE conversa_id=$1 AND conta_id=$2`, conversaID, contaID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("contar mensagens WhatsApp: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,conversa_id,identificador_externo,direcao,tipo,conteudo_protegido,status,
		ocorrida_em,enviada_em,entregue_em,lida_em,falhou_em,erro_codigo,erro_detalhe
		FROM mensagens_whatsapp WHERE conversa_id=$1 AND conta_id=$2
		ORDER BY ocorrida_em DESC LIMIT $3 OFFSET $4`, conversaID, contaID, filtro.Limite, (filtro.Pagina-1)*filtro.Limite)
	if err != nil {
		return nil, fmt.Errorf("listar mensagens WhatsApp: %w", err)
	}
	defer rows.Close()
	dados := make([]*domain.MensagemWhatsApp, 0, filtro.Limite)
	for rows.Next() {
		mensagem := &domain.MensagemWhatsApp{}
		if err := rows.Scan(&mensagem.ID, &mensagem.ConversaID, &mensagem.IdentificadorExterno, &mensagem.Direcao,
			&mensagem.Tipo, &mensagem.ConteudoProtegido, &mensagem.Status, &mensagem.OcorridaEm, &mensagem.EnviadaEm,
			&mensagem.EntregueEm, &mensagem.LidaEm, &mensagem.FalhouEm, &mensagem.ErroCodigo, &mensagem.ErroDetalhe); err != nil {
			return nil, fmt.Errorf("ler mensagem WhatsApp: %w", err)
		}
		dados = append(dados, mensagem)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("percorrer mensagens WhatsApp: %w", err)
	}
	return &domain.ListaPaginada[*domain.MensagemWhatsApp]{Dados: dados, Total: total, Pagina: filtro.Pagina, Limite: filtro.Limite}, nil
}

func (r *integracaoWhatsAppPostgres) CriarMensagemSaida(ctx context.Context, solicitacao *domain.SolicitacaoEnvioWhatsApp, conteudoProtegido string, evento *domain.EventoOutbox) error {
	if solicitacao == nil || evento == nil {
		return errors.New("mensagem e evento WhatsApp são obrigatórios")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar mensagem WhatsApp de saída: %w", err)
	}
	defer tx.Rollback()
	if err := inserirEventoOutbox(ctx, tx, evento); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO mensagens_whatsapp
		(id,conta_id,conversa_id,identificador_externo,direcao,tipo,conteudo_protegido,status,ocorrida_em,evento_outbox_id,criado_em,atualizado_em)
		VALUES ($1,$2,$3,NULL,'SAIDA',$4,$5,'PENDENTE',NOW(),$6,NOW(),NOW())`,
		solicitacao.IDMensagem, solicitacao.ContaID, solicitacao.ConversaID, solicitacao.Tipo, conteudoProtegido, evento.ID)
	if err != nil {
		return fmt.Errorf("persistir mensagem WhatsApp de saída: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar mensagem WhatsApp de saída: %w", err)
	}
	return nil
}

func (r *integracaoWhatsAppPostgres) MarcarMensagemEnviada(ctx context.Context, mensagemID, identificadorExterno string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar confirmação de envio WhatsApp: %w", err)
	}
	defer tx.Rollback()
	resultado, err := tx.ExecContext(ctx, `UPDATE mensagens_whatsapp SET identificador_externo=$2,status='ENVIADA',
		enviada_em=NOW(),atualizado_em=NOW() WHERE id=$1 AND direcao='SAIDA' AND status IN ('PENDENTE','ENVIADA')`, mensagemID, identificadorExterno)
	if err != nil {
		return fmt.Errorf("marcar mensagem WhatsApp enviada: %w", err)
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("mensagem WhatsApp de saída não encontrada")
	}
	if err := aplicarStatusWhatsAppPendente(ctx, tx, identificadorExterno); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar envio WhatsApp: %w", err)
	}
	return nil
}

func (r *integracaoWhatsAppPostgres) ObterIdentificadorMensagemSaida(ctx context.Context, mensagemID string) (*string, error) {
	var identificador sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT identificador_externo FROM mensagens_whatsapp WHERE id=$1 AND direcao='SAIDA'`, mensagemID).Scan(&identificador)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("obter mensagem WhatsApp de saída: %w", err)
	}
	if !identificador.Valid || identificador.String == "" {
		return nil, nil
	}
	return &identificador.String, nil
}

func (r *integracaoWhatsAppPostgres) AtualizarStatusMensagem(ctx context.Context, identificadorExterno, status, codigoErro, detalheErro string, ocorridoEm time.Time) error {
	detalheErro = strings.TrimSpace(detalheErro)
	if len([]rune(detalheErro)) > 500 {
		detalheErro = string([]rune(detalheErro)[:500])
	}
	ordem := map[string]int{"FALHOU": 0, "ENVIADA": 1, "ENTREGUE": 2, "LIDA": 3}[status]
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar status WhatsApp: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO status_mensagens_whatsapp
		(identificador_externo,status,ordem,ocorrido_em,erro_codigo,erro_detalhe,atualizado_em)
		VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),NOW())
		ON CONFLICT (identificador_externo) DO UPDATE SET status=EXCLUDED.status,ordem=EXCLUDED.ordem,
			ocorrido_em=EXCLUDED.ocorrido_em,erro_codigo=EXCLUDED.erro_codigo,erro_detalhe=EXCLUDED.erro_detalhe,atualizado_em=NOW()
		WHERE EXCLUDED.ordem > status_mensagens_whatsapp.ordem
		   OR (EXCLUDED.status='FALHOU' AND status_mensagens_whatsapp.status IN ('ENVIADA'))`,
		identificadorExterno, status, ordem, ocorridoEm, codigoErro, detalheErro)
	if err != nil {
		return fmt.Errorf("persistir status WhatsApp: %w", err)
	}
	if err := aplicarStatusWhatsAppPendente(ctx, tx, identificadorExterno); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar status WhatsApp: %w", err)
	}
	return nil
}

func aplicarStatusWhatsAppPendente(ctx context.Context, tx *sql.Tx, identificadorExterno string) error {
	_, err := tx.ExecContext(ctx, `UPDATE mensagens_whatsapp mensagem SET
		status=evento.status,
		enviada_em=CASE WHEN evento.status='ENVIADA' THEN COALESCE(mensagem.enviada_em,evento.ocorrido_em) ELSE mensagem.enviada_em END,
		entregue_em=CASE WHEN evento.status='ENTREGUE' THEN COALESCE(mensagem.entregue_em,evento.ocorrido_em) ELSE mensagem.entregue_em END,
		lida_em=CASE WHEN evento.status='LIDA' THEN COALESCE(mensagem.lida_em,evento.ocorrido_em) ELSE mensagem.lida_em END,
		falhou_em=CASE WHEN evento.status='FALHOU' THEN COALESCE(mensagem.falhou_em,evento.ocorrido_em) ELSE mensagem.falhou_em END,
		erro_codigo=evento.erro_codigo,erro_detalhe=evento.erro_detalhe,atualizado_em=NOW()
		FROM status_mensagens_whatsapp evento
		WHERE mensagem.identificador_externo=$1 AND evento.identificador_externo=$1 AND mensagem.direcao='SAIDA'
		  AND (evento.ordem > CASE mensagem.status WHEN 'PENDENTE' THEN 0 WHEN 'ENVIADA' THEN 1 WHEN 'ENTREGUE' THEN 2 WHEN 'LIDA' THEN 3 ELSE 0 END
		       OR (evento.status='FALHOU' AND mensagem.status IN ('PENDENTE','ENVIADA')))`, identificadorExterno)
	if err != nil {
		return fmt.Errorf("aplicar status WhatsApp: %w", err)
	}
	return nil
}

func (r *integracaoWhatsAppPostgres) RegistrarConsentimento(ctx context.Context, conversaID, contaID string, consentiu bool, origem, evidencia string) error {
	resultado, err := r.db.ExecContext(ctx, `UPDATE conversas_whatsapp SET consentimento_marketing=$3,
		consentimento_marketing_em=CASE WHEN $3 THEN NOW() ELSE NULL END,
		consentimento_marketing_origem=NULLIF($4,''),consentimento_marketing_evidencia=NULLIF($5,''),atualizado_em=NOW()
		WHERE id=$1 AND conta_id=$2`, conversaID, contaID, consentiu, origem, evidencia)
	if err != nil {
		return fmt.Errorf("registrar consentimento WhatsApp: %w", err)
	}
	linhas, err := resultado.RowsAffected()
	if err != nil || linhas != 1 {
		return errors.New("conversa WhatsApp não encontrada")
	}
	return nil
}
