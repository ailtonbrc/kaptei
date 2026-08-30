package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type ConfiguracaoHandler struct {
	configService ports.ConfiguracaoService
}

func NewConfiguracaoHandler(cs ports.ConfiguracaoService) *ConfiguracaoHandler {
	return &ConfiguracaoHandler{configService: cs}
}

func exigirSuperAdmin(w http.ResponseWriter, r *http.Request) bool {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return false
	}
	if usuario.Papel != domain.RoleSuperAdmin {
		responderErro(w, http.StatusForbidden, "Acesso restrito ao administrador da plataforma")
		return false
	}
	return true
}

func ocultarSegredos(config *domain.ConfiguracaoSistema) {
	if config == nil || (config.Chave != "SMTP_CONFIG" && config.Chave != "OBSERVABILIDADE_CONFIG") {
		return
	}

	var valor map[string]interface{}
	if err := json.Unmarshal(config.Valor, &valor); err != nil {
		return
	}
	if config.Chave == "SMTP_CONFIG" {
		senha, possuiSenha := valor["password"].(string)
		valor["password"] = ""
		valor["password_configured"] = possuiSenha && senha != ""
	} else {
		token, possuiToken := valor["token"].(string)
		valor["token"] = ""
		valor["token_configurado"] = possuiToken && token != ""
	}
	if valorSeguro, err := json.Marshal(valor); err == nil {
		config.Valor = valorSeguro
	}
}

func (h *ConfiguracaoHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if !exigirSuperAdmin(w, r) {
		return
	}

	chave := r.PathValue("chave")
	if chave == "" {
		responderErro(w, http.StatusBadRequest, "Chave nÃ£o fornecida")
		return
	}

	config, err := h.configService.GetConfig(r.Context(), chave)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	if config == nil {
		responderErro(w, http.StatusNotFound, "ConfiguraÃ§Ã£o nÃ£o encontrada")
		return
	}
	ocultarSegredos(config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

type updateConfigRequest struct {
	Valor     interface{} `json:"valor"`
	Descricao string      `json:"descricao"`
}

func (h *ConfiguracaoHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if !exigirSuperAdmin(w, r) {
		return
	}

	chave := r.PathValue("chave")
	if chave == "" {
		responderErro(w, http.StatusBadRequest, "Chave nÃ£o fornecida")
		return
	}

	var req updateConfigRequest
	if err := decodificarJSONLimitado(w, r, &req, 32*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "Corpo da requisiÃ§Ã£o invÃ¡lido")
		return
	}

	err := h.configService.UpdateConfig(r.Context(), chave, req.Valor, req.Descricao)
	if err != nil {
		responderErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "ConfiguraÃ§Ã£o salva com sucesso"})
}

func (h *ConfiguracaoHandler) GetPublicConfig(w http.ResponseWriter, r *http.Request) {
	chave := r.PathValue("chave")
	if chave == "" {
		responderErro(w, http.StatusBadRequest, "Chave nÃ£o fornecida")
		return
	}

	// Whitelist de configuraÃ§Ãµes pÃºblicas seguras
	publicKeys := map[string]bool{
		"GOOGLE_CLIENT_ID": true,
	}

	if !publicKeys[chave] {
		responderErro(w, http.StatusForbidden, "Acesso negado para esta configuraÃ§Ã£o")
		return
	}

	config, err := h.configService.GetConfig(r.Context(), chave)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	if config == nil {
		responderErro(w, http.StatusNotFound, "ConfiguraÃ§Ã£o nÃ£o encontrada")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}
