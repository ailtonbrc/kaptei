package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type configuracaoService struct {
	configRepo ports.ConfiguracaoRepository
	segredos   ports.ProtetorSegredos
}

func NewConfiguracaoService(repo ports.ConfiguracaoRepository, segredos ports.ProtetorSegredos) ports.ConfiguracaoService {
	return &configuracaoService{configRepo: repo, segredos: segredos}
}

func (s *configuracaoService) GetConfig(ctx context.Context, chave string) (*domain.ConfiguracaoSistema, error) {
	if chave == "" {
		return nil, errors.New("chave não informada")
	}
	return s.configRepo.Get(ctx, chave)
}

func (s *configuracaoService) UpdateConfig(ctx context.Context, chave string, valor interface{}, descricao string) error {
	chave = strings.ToUpper(strings.TrimSpace(chave))
	if chave == "" {
		return errors.New("chave não informada")
	}

	valorNormalizado, err := s.normalizarValor(ctx, chave, valor)
	if err != nil {
		return err
	}
	valorJSON, err := json.Marshal(valorNormalizado)
	if err != nil {
		return errors.New("erro ao converter valor para JSON")
	}

	var desc *string
	if descricao != "" {
		desc = &descricao
	}

	config := &domain.ConfiguracaoSistema{
		Chave:        chave,
		Valor:        valorJSON,
		Descricao:    desc,
		AtualizadoEm: time.Now(),
	}

	return s.configRepo.Set(ctx, config)
}

func (s *configuracaoService) normalizarValor(ctx context.Context, chave string, valor interface{}) (interface{}, error) {
	switch chave {
	case "GOOGLE_CLIENT_ID":
		clientID, ok := valor.(string)
		clientID = strings.TrimSpace(clientID)
		if !ok || clientID == "" || !strings.HasSuffix(clientID, ".apps.googleusercontent.com") {
			return nil, errors.New("Google Client ID inválido")
		}
		return clientID, nil
	case "TRIAL_DIAS_PADRAO":
		mapa, ok := valor.(map[string]interface{})
		dias, numero := mapa["dias"].(float64)
		if !ok || !numero || dias < 1 || dias > 90 || dias != float64(int(dias)) {
			return nil, errors.New("o período de avaliação deve possuir entre 1 e 90 dias")
		}
		return map[string]int{"dias": int(dias)}, nil
	case "SMTP_CONFIG":
		dados, ok := valor.(map[string]interface{})
		if !ok {
			return nil, errors.New("configuração SMTP inválida")
		}
		host, _ := dados["host"].(string)
		usuario, _ := dados["user"].(string)
		remetente, _ := dados["from_email"].(string)
		nomeRemetente, _ := dados["from_name"].(string)
		senha, _ := dados["password"].(string)
		senhaInformada := strings.TrimSpace(senha) != ""
		porta, portaValida := dados["port"].(float64)
		host, usuario, remetente, nomeRemetente = strings.TrimSpace(host), strings.TrimSpace(usuario), strings.TrimSpace(remetente), strings.TrimSpace(nomeRemetente)
		if host == "" || len(host) > 253 || usuario == "" || len(usuario) > 254 || nomeRemetente == "" || len([]rune(nomeRemetente)) > 120 || !emailExato(remetente) || !portaValida || porta < 1 || porta > 65535 || porta != float64(int(porta)) {
			return nil, errors.New("host, porta, usuário e e-mail remetente SMTP são obrigatórios")
		}
		if strings.TrimSpace(senha) == "" {
			atual, err := s.configRepo.Get(ctx, chave)
			if err != nil {
				return nil, fmt.Errorf("carregar senha SMTP existente: %w", err)
			}
			if atual != nil {
				var anterior map[string]interface{}
				if json.Unmarshal(atual.Valor, &anterior) == nil {
					senha, _ = anterior["password"].(string)
				}
			}
		}
		if senha == "" {
			return nil, errors.New("senha SMTP obrigatória na primeira configuração")
		}
		if len(senha) > 4096 {
			return nil, errors.New("senha SMTP excede o limite permitido")
		}
		if senhaInformada || !strings.HasPrefix(senha, "enc:v1:") {
			senhaProtegida, err := s.segredos.Proteger(senha)
			if err != nil {
				return nil, fmt.Errorf("proteger senha SMTP: %w", err)
			}
			senha = senhaProtegida
		}
		return map[string]interface{}{
			"host": host, "port": int(porta), "user": usuario,
			"password": senha, "from_email": remetente, "from_name": nomeRemetente,
		}, nil
	case "OBSERVABILIDADE_CONFIG":
		dados, ok := valor.(map[string]interface{})
		if !ok {
			return nil, errors.New("configuração de observabilidade inválida")
		}
		ativa, _ := dados["ativa"].(bool)
		token, _ := dados["token"].(string)
		token = strings.TrimSpace(token)
		tokenInformado := token != ""
		if !tokenInformado {
			atual, err := s.configRepo.Get(ctx, chave)
			if err != nil {
				return nil, fmt.Errorf("carregar token de métricas existente: %w", err)
			}
			if atual != nil {
				var anterior map[string]interface{}
				if json.Unmarshal(atual.Valor, &anterior) == nil {
					token, _ = anterior["token"].(string)
				}
			}
		}
		if ativa && token == "" {
			return nil, errors.New("token de métricas obrigatório quando a exportação está ativa")
		}
		if tokenInformado {
			if len(token) < 32 || len(token) > 512 {
				return nil, errors.New("token de métricas deve possuir entre 32 e 512 caracteres")
			}
			protegido, err := s.segredos.Proteger(token)
			if err != nil {
				return nil, fmt.Errorf("proteger token de métricas: %w", err)
			}
			token = protegido
		}
		return map[string]interface{}{"ativa": ativa, "token": token}, nil
	default:
		return nil, errors.New("chave de configuração não suportada")
	}
}

func emailExato(valor string) bool {
	endereco, err := mail.ParseAddress(valor)
	return err == nil && strings.EqualFold(endereco.Address, valor) && len(valor) <= 254
}
