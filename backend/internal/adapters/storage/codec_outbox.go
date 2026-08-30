package storage

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type codecObjetoOutbox struct {
	segredos         ports.ProtetorSegredos
	maximoTentativas int
}

func NewCodecObjetoOutbox(segredos ports.ProtetorSegredos, maximoTentativas int) ports.PreparadorObjetoOutbox {
	return &codecObjetoOutbox{segredos: segredos, maximoTentativas: maximoTentativas}
}

func (c *codecObjetoOutbox) PrepararExclusaoObjeto(contaID, chaveIdempotencia, provedor, chave string) (*domain.EventoOutbox, error) {
	chaveIdempotencia = strings.TrimSpace(chaveIdempotencia)
	provedor = strings.TrimSpace(provedor)
	chave = strings.TrimSpace(chave)
	if contaID == "" || chaveIdempotencia == "" || provedor == "" || chave == "" || len(chaveIdempotencia) > 200 {
		return nil, errors.New("dados da exclusão de objeto inválidos")
	}
	payload, err := json.Marshal(domain.SolicitacaoExclusaoObjeto{Provedor: provedor, Chave: chave})
	if err != nil {
		return nil, errors.New("não foi possível preparar a exclusão do objeto")
	}
	protegido, err := c.segredos.Proteger(string(payload))
	if err != nil {
		return nil, errors.New("não foi possível proteger a exclusão do objeto")
	}
	agora := time.Now().UTC()
	return &domain.EventoOutbox{
		ContaID: &contaID, Tipo: domain.TipoEventoExcluirObjeto,
		PayloadProtegido: protegido, ChaveIdempotencia: chaveIdempotencia,
		Status: "PENDENTE", MaximoTentativas: c.maximoTentativas,
		DisponivelEm: agora, CriadoEm: agora,
	}, nil
}

func (c *codecObjetoOutbox) DecodificarExclusaoObjeto(evento *domain.EventoOutbox) (*domain.SolicitacaoExclusaoObjeto, error) {
	if evento == nil || evento.Tipo != domain.TipoEventoExcluirObjeto {
		return nil, errors.New("evento de exclusão de objeto inválido")
	}
	payload, err := c.segredos.Revelar(evento.PayloadProtegido)
	if err != nil {
		return nil, errors.New("não foi possível revelar a exclusão do objeto")
	}
	var solicitacao domain.SolicitacaoExclusaoObjeto
	if err := json.Unmarshal([]byte(payload), &solicitacao); err != nil || solicitacao.Provedor == "" || solicitacao.Chave == "" {
		return nil, errors.New("payload da exclusão de objeto inválido")
	}
	return &solicitacao, nil
}
