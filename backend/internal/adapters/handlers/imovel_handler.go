package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/msdev/kaptei/internal/adapters/middlewares"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
)

type ImovelHandler struct {
	imovelService ports.ImovelService
	maximoUpload  int64
	vagasUpload   chan struct{}
}

func NewImovelHandler(s ports.ImovelService, maximoUpload int64, maximoConcorrente int) *ImovelHandler {
	return &ImovelHandler{imovelService: s, maximoUpload: maximoUpload, vagasUpload: make(chan struct{}, maximoConcorrente)}
}

func (h *ImovelHandler) Create(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	var imovel domain.Imovel
	if err := decodificarJSONLimitado(w, r, &imovel, 128*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "RequisiÃ§Ã£o invÃ¡lida")
		return
	}

	// ForÃ§ar vÃ­nculos de seguranÃ§a
	imovel.ContaID = usuario.ContaID
	imovel.UsuarioID = usuario.ID

	if err := h.imovelService.Create(r.Context(), &imovel); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(imovel)
}

func (h *ImovelHandler) List(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	imoveis, err := h.imovelService.List(r.Context(), usuario.ContaID, filtroPaginacao(r))
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(imoveis)
}

func (h *ImovelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	id := r.PathValue("id")

	imovel, err := h.imovelService.GetByID(r.Context(), id, usuario.ContaID)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}
	if imovel == nil {
		responderErro(w, http.StatusNotFound, "ImÃ³vel nÃ£o encontrado")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(imovel)
}

func (h *ImovelHandler) Update(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	id := r.PathValue("id")

	var imovel domain.Imovel
	if err := decodificarJSONLimitado(w, r, &imovel, 128*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "RequisiÃ§Ã£o invÃ¡lida")
		return
	}

	imovel.ID = id
	imovel.ContaID = usuario.ContaID // ForÃ§a seguranÃ§a Multi-Tenant

	if err := h.imovelService.Update(r.Context(), &imovel); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "ImÃ³vel atualizado com sucesso"})
}

func (h *ImovelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	id := r.PathValue("id")

	if err := h.imovelService.Delete(r.Context(), id, usuario.ContaID); err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "ImÃ³vel excluÃ­do com sucesso"})
}

type addFotoRequest struct {
	URL    string `json:"url"`
	IsCapa bool   `json:"is_capa"`
}

func (h *ImovelHandler) AddFoto(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}

	imovelID := r.PathValue("id")

	// Validar se imÃ³vel pertence Ã  conta antes de adicionar foto
	imovel, err := h.imovelService.GetByID(r.Context(), imovelID, usuario.ContaID)
	if err != nil || imovel == nil {
		responderErro(w, http.StatusNotFound, "ImÃ³vel nÃ£o encontrado")
		return
	}

	var req addFotoRequest
	if err := decodificarJSONLimitado(w, r, &req, 16*1024); err != nil {
		responderErro(w, http.StatusBadRequest, "RequisiÃ§Ã£o invÃ¡lida")
		return
	}

	foto, err := h.imovelService.AddFoto(r.Context(), imovelID, req.URL, req.IsCapa)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(foto)
}

func (h *ImovelHandler) UploadFoto(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "Não autorizado")
		return
	}
	select {
	case h.vagasUpload <- struct{}{}:
		defer func() { <-h.vagasUpload }()
	default:
		responderErro(w, http.StatusTooManyRequests, "Processamento de imagens ocupado; tente novamente em instantes")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maximoUpload+64*1024)
	if err := r.ParseMultipartForm(h.maximoUpload); err != nil {
		responderErro(w, http.StatusBadRequest, "Imagem ausente ou acima do limite permitido")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	arquivo, _, err := r.FormFile("arquivo")
	if err != nil {
		responderErro(w, http.StatusBadRequest, "Selecione uma imagem")
		return
	}
	defer arquivo.Close()
	conteudo, err := io.ReadAll(io.LimitReader(arquivo, h.maximoUpload+1))
	if err != nil || int64(len(conteudo)) > h.maximoUpload {
		responderErro(w, http.StatusBadRequest, "Não foi possível ler a imagem ou o limite foi excedido")
		return
	}
	isCapa, _ := strconv.ParseBool(r.FormValue("is_capa"))
	foto, err := h.imovelService.UploadFoto(r.Context(), r.PathValue("id"), usuario.ContaID, conteudo, isCapa)
	if err != nil {
		responderErro(w, http.StatusBadRequest, err.Error())
		return
	}
	responderJSON(w, http.StatusCreated, foto)
}

func (h *ImovelHandler) DeleteFoto(w http.ResponseWriter, r *http.Request) {
	usuario, ok := r.Context().Value(middlewares.UserContextKey).(*domain.Usuario)
	if !ok {
		responderErro(w, http.StatusUnauthorized, "NÃ£o autorizado")
		return
	}
	imovelID := r.PathValue("id")
	if imovel, err := h.imovelService.GetByID(r.Context(), imovelID, usuario.ContaID); err != nil || imovel == nil {
		responderErro(w, http.StatusNotFound, "ImÃ³vel nÃ£o encontrado")
		return
	}
	if err := h.imovelService.DeleteFoto(r.Context(), r.PathValue("foto_id"), imovelID, usuario.ContaID); err != nil {
		responderErro(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
