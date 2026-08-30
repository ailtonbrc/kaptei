package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/msdev/kaptei/internal/core/domain"
	"github.com/msdev/kaptei/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type authService struct {
	userRepo        ports.UsuarioRepository
	contaRepo       ports.ContaRepository
	cadastroRepo    ports.CadastroRepository
	recuperacaoRepo ports.RecuperacaoSenhaRepository
	sessaoRepo      ports.SessaoRepository
	preparadorEmail ports.PreparadorEmailOutbox
	configRepo      ports.ConfiguracaoRepository
	planoRepo       ports.PlanoRepository
	jwtSecret       string
	urlPublica      string
}

var hashComparacaoInvalida = func() []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha-invalida-para-comparacao"), bcrypt.DefaultCost)
	return hash
}()

func NewAuthService(
	u ports.UsuarioRepository,
	c ports.ContaRepository,
	cadastro ports.CadastroRepository,
	r ports.RecuperacaoSenhaRepository,
	sessao ports.SessaoRepository,
	preparadorEmail ports.PreparadorEmailOutbox,
	cfg ports.ConfiguracaoRepository,
	p ports.PlanoRepository,
	jwtSecret, urlPublica string,
) ports.AuthService {
	if len(jwtSecret) < 32 {
		panic("JWT_SECRET deve estar configurada com pelo menos 32 caracteres")
	}
	return &authService{
		userRepo:        u,
		contaRepo:       c,
		cadastroRepo:    cadastro,
		recuperacaoRepo: r,
		sessaoRepo:      sessao,
		preparadorEmail: preparadorEmail,
		configRepo:      cfg,
		planoRepo:       p,
		jwtSecret:       jwtSecret,
		urlPublica:      strings.TrimRight(urlPublica, "/"),
	}
}

// Login valida email e senha e retorna um JWT
func (s *authService) Login(ctx context.Context, email, senha string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	usuario, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	hash := hashComparacaoInvalida
	if usuario != nil && usuario.SenhaHash != nil {
		hash = []byte(*usuario.SenhaHash)
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(senha))
	if err != nil || usuario == nil || usuario.SenhaHash == nil || !strings.EqualFold(usuario.Status, "ATIVO") {
		return "", errors.New("credenciais inválidas")
	}

	return s.generateJWT(ctx, usuario)
}

// Register cria uma nova conta SaaS e um novo Usuário
func (s *authService) Register(ctx context.Context, nome, email, senha, tipoConta, plano string) (string, error) {
	nome = strings.TrimSpace(nome)
	email = strings.ToLower(strings.TrimSpace(email))
	tipoConta = strings.ToUpper(strings.TrimSpace(tipoConta))
	plano = strings.ToUpper(strings.TrimSpace(plano))
	endereco, erroEmail := mail.ParseAddress(email)
	if nome == "" || len([]rune(nome)) > 120 || erroEmail != nil || !strings.EqualFold(endereco.Address, email) ||
		len(senha) < 6 || len([]byte(senha)) > 72 {
		return "", errors.New("nome, e-mail válido e senha de no mínimo 6 caracteres são obrigatórios")
	}
	// 1. Verifica se email já existe
	existente, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("não foi possível validar o e-mail")
	}
	if existente != nil {
		return "", errors.New("este e-mail já está em uso")
	}

	// 2. Hash da senha
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("erro ao criptografar senha")
	}
	hashStr := string(hash)

	// 3. Valida plano, papel e período de avaliação em uma única regra reutilizável.
	novaConta, papel, err := s.prepararNovaConta(ctx, tipoConta, plano)
	if err != nil {
		return "", err
	}
	// 5. Cria o Usuário Administrador da Conta
	usuario := &domain.Usuario{
		NomeCompleto: nome,
		Email:        email,
		SenhaHash:    &hashStr,
		Papel:        papel,
		Status:       "ativo",
	}

	err = s.cadastroRepo.CriarContaEUsuario(ctx, novaConta, usuario)
	if err != nil {
		return "", errors.New("não foi possível concluir o cadastro")
	}

	// 6. Retorna o JWT para login automático
	return s.generateJWT(ctx, usuario)
}

func (s *authService) prepararNovaConta(ctx context.Context, tipoConta, plano string) (*domain.ContaSaaS, domain.Role, error) {
	tipoConta = strings.ToUpper(strings.TrimSpace(tipoConta))
	plano = strings.ToUpper(strings.TrimSpace(plano))
	if tipoConta != "CORRETOR_SOLO" && tipoConta != "IMOBILIARIA" {
		return nil, "", errors.New("tipo de conta inválido")
	}
	planoObj, err := s.planoRepo.GetByCodigo(ctx, plano)
	if err != nil || planoObj == nil || !planoObj.Ativo {
		return nil, "", errors.New("plano inválido ou indisponível")
	}
	tipoPlanoEsperado := "CORRETOR"
	papel := domain.RoleCorretorSolo
	if tipoConta == "IMOBILIARIA" {
		tipoPlanoEsperado = "IMOBILIARIA"
		papel = domain.RoleGestor
	}
	if planoObj.Tipo != tipoPlanoEsperado {
		return nil, "", errors.New("plano incompatível com o tipo de conta")
	}

	conta := &domain.ContaSaaS{TipoConta: tipoConta, StatusPlano: "AGUARDANDO_PAGAMENTO", Plano: plano}
	if planoObj.Preco == 0 && strings.Contains(planoObj.Codigo, "TRIAL") {
		dias := 14
		if s.configRepo != nil {
			if cfg, _ := s.configRepo.Get(ctx, "TRIAL_DIAS_PADRAO"); cfg != nil {
				var configuracao struct {
					Dias int `json:"dias"`
				}
				if json.Unmarshal(cfg.Valor, &configuracao) == nil && configuracao.Dias > 0 {
					dias = configuracao.Dias
				}
			}
		}
		venceEm := time.Now().AddDate(0, 0, dias)
		conta.StatusPlano = "TRIAL"
		conta.TrialVenceEm = &venceEm
	}
	return conta, papel, nil
}

// GoogleLogin valida identidade, e-mail verificado e aplica as mesmas regras do cadastro tradicional.
func (s *authService) GoogleLogin(ctx context.Context, googleToken, tipoConta, plano string) (string, error) {
	clientID := ""
	if s.configRepo != nil {
		if cfg, _ := s.configRepo.Get(ctx, "GOOGLE_CLIENT_ID"); cfg != nil {
			if err := json.Unmarshal(cfg.Valor, &clientID); err != nil {
				clientID = strings.Trim(string(cfg.Valor), "\"")
			}
		}
	}
	if strings.TrimSpace(clientID) == "" {
		return "", errors.New("autenticação Google não configurada")
	}
	payload, err := idtoken.Validate(ctx, googleToken, clientID)
	if err != nil {
		return "", errors.New("token Google inválido")
	}
	email, emailValido := payload.Claims["email"].(string)
	emailVerificado, verificado := payload.Claims["email_verified"].(bool)
	if !emailValido || strings.TrimSpace(email) == "" || !verificado || !emailVerificado || payload.Subject == "" {
		return "", errors.New("identidade Google sem e-mail verificado")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	googleID := payload.Subject
	nome, _ := payload.Claims["name"].(string)
	if strings.TrimSpace(nome) == "" {
		nome = email
	}
	var avatar *string
	if imagem, ok := payload.Claims["picture"].(string); ok && strings.TrimSpace(imagem) != "" {
		avatar = &imagem
	}

	usuario, err := s.userRepo.GetByGoogleID(ctx, googleID)
	if err != nil {
		return "", err
	}
	if usuario == nil {
		usuario, err = s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return "", err
		}
	}
	if usuario != nil {
		if !strings.EqualFold(usuario.Status, "ATIVO") {
			return "", errors.New("usuário inativo")
		}
		if usuario.GoogleID != nil && *usuario.GoogleID != googleID {
			return "", errors.New("e-mail já vinculado a outra identidade Google")
		}
		if usuario.GoogleID == nil {
			if err := s.userRepo.VincularGoogle(ctx, usuario.ID, googleID, avatar); err != nil {
				return "", errors.New("não foi possível vincular a identidade Google")
			}
			usuario.GoogleID = &googleID
			usuario.URLAvatar = avatar
		}
		return s.generateJWT(ctx, usuario)
	}

	if strings.TrimSpace(tipoConta) == "" || strings.TrimSpace(plano) == "" {
		return "", errors.New("cadastro Google requer tipo de conta e plano")
	}
	novaConta, papel, err := s.prepararNovaConta(ctx, tipoConta, plano)
	if err != nil {
		return "", err
	}
	usuario = &domain.Usuario{
		NomeCompleto: strings.TrimSpace(nome), Email: email, GoogleID: &googleID,
		Papel: papel, Status: "ATIVO", URLAvatar: avatar,
	}
	if err := s.cadastroRepo.CriarContaEUsuario(ctx, novaConta, usuario); err != nil {
		return "", errors.New("não foi possível concluir o cadastro Google")
	}
	return s.generateJWT(ctx, usuario)
}

// ValidateToken decodifica o JWT e retorna o Usuário (Usado no Middleware)
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*domain.Usuario, error) {
	claims, err := s.lerClaimsToken(tokenString)
	if err != nil {
		return nil, err
	}
	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return nil, errors.New("identificador do token inválido")
	}
	sessaoID, ok := claims["sid"].(string)
	if !ok || sessaoID == "" {
		return nil, errors.New("sessão ausente no token")
	}
	ativa, err := s.sessaoRepo.EstaAtiva(ctx, sessaoID, userID)
	if err != nil || !ativa {
		return nil, errors.New("sessão expirada ou revogada")
	}

	usuario, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || usuario == nil || !strings.EqualFold(usuario.Status, "ATIVO") {
		return nil, errors.New("usuário inválido ou inativo")
	}
	versaoToken, ok := claims["versao_sessao"].(float64)
	if !ok || int(versaoToken) != usuario.VersaoSessao {
		return nil, errors.New("sessão revogada")
	}
	return usuario, nil
}

func (s *authService) lerClaimsToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(_ *jwt.Token) (interface{}, error) { return []byte(s.jwtSecret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer("kaptei"),
		jwt.WithAudience("kaptei-web"),
	)

	if err != nil || !token.Valid {
		return nil, errors.New("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("falha ao ler claims do token")
	}
	return claims, nil
}

// Internal func
func (s *authService) generateJWT(ctx context.Context, u *domain.Usuario) (string, error) {
	// Buscar informações da conta para embutir no JWT (Plano e Status)
	conta, err := s.contaRepo.GetByID(ctx, u.ContaID)
	if err != nil || conta == nil {
		return "", errors.New("não foi possível carregar a conta da sessão")
	}

	var trialVenceEm string
	var statusPlano string = "TRIAL"
	var plano string = ""

	statusPlano = conta.StatusPlano
	plano = conta.Plano
	if conta.TrialVenceEm != nil {
		trialVenceEm = conta.TrialVenceEm.Format(time.RFC3339)
	}

	agora := time.Now().UTC()
	expiraEm := agora.Add(24 * time.Hour)
	sessao := &domain.SessaoUsuario{UsuarioID: u.ID, ContaID: u.ContaID, ExpiraEm: expiraEm}
	if err := s.sessaoRepo.Criar(ctx, sessao); err != nil {
		return "", errors.New("não foi possível registrar a sessão")
	}
	claims := jwt.MapClaims{
		"sub":            u.ID,
		"conta_id":       u.ContaID,
		"papel":          u.Papel,
		"email":          u.Email,
		"nome":           u.NomeCompleto,
		"avatar":         u.URLAvatar,
		"status_plano":   statusPlano,
		"plano":          plano,
		"trial_vence_em": trialVenceEm,
		"versao_sessao":  u.VersaoSessao,
		"sid":            sessao.ID,
		"iss":            "kaptei",
		"aud":            "kaptei-web",
		"exp":            expiraEm.Unix(),
		"iat":            agora.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	assinado, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		_ = s.sessaoRepo.Revogar(ctx, sessao.ID, u.ID)
		return "", err
	}
	return assinado, nil
}

func (s *authService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.lerClaimsToken(tokenString)
	if err != nil {
		return err
	}
	usuarioID, usuarioOK := claims["sub"].(string)
	sessaoID, sessaoOK := claims["sid"].(string)
	if !usuarioOK || !sessaoOK || usuarioID == "" || sessaoID == "" {
		return errors.New("sessão inválida")
	}
	return s.sessaoRepo.Revogar(ctx, sessaoID, usuarioID)
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashTokenRecuperacao(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *authService) SolicitarRecuperacaoSenha(ctx context.Context, email string) error {
	usuario, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if usuario == nil {
		// Não retornar erro se não achar, para não expor quais emails existem
		return nil
	}

	if s.urlPublica == "" {
		return errors.New("URL pública da aplicação não configurada")
	}
	tokenStr, err := generateSecureToken()
	if err != nil {
		return errors.New("não foi possível gerar o token de recuperação")
	}
	expiraEm := time.Now().Add(2 * time.Hour) // Token válido por 2 horas

	token := &domain.RecuperacaoSenhaToken{
		UsuarioID: usuario.ID,
		Token:     hashTokenRecuperacao(tokenStr),
		ExpiraEm:  expiraEm,
		Usado:     false,
		CriadoEm:  time.Now(),
	}

	resetLink := fmt.Sprintf("%s/redefinir-senha?token=%s", s.urlPublica, tokenStr)
	body := fmt.Sprintf(`
		<h2>Recuperação de Senha</h2>
		<p>Olá, %s!</p>
		<p>Você solicitou a recuperação de senha. Clique no link abaixo para criar uma nova senha:</p>
		<p><a href="%s">Redefinir minha senha</a></p>
		<p>Este link expira em 2 horas.</p>
		<p>Se não foi você, ignore este e-mail.</p>
	`, html.EscapeString(usuario.NomeCompleto), html.EscapeString(resetLink))

	evento, err := s.preparadorEmail.PrepararEmail(
		&usuario.ContaID,
		"recuperacao-senha:"+token.Token,
		usuario.Email,
		"Recuperação de Senha - Kaptei",
		body,
	)
	if err != nil {
		return err
	}
	return s.recuperacaoRepo.CreateToken(ctx, token, evento)
}

func (s *authService) RedefinirSenha(ctx context.Context, tokenStr, novaSenha string) error {
	tokenStr = strings.TrimSpace(tokenStr)
	if len(tokenStr) != 64 {
		return errors.New("token inválido")
	}
	if len(novaSenha) < 6 || len([]byte(novaSenha)) > 72 {
		return errors.New("a nova senha deve possuir entre 6 e 72 bytes")
	}
	token, err := s.recuperacaoRepo.GetToken(ctx, hashTokenRecuperacao(tokenStr))
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("token inválido ou inexistente")
	}
	if token.Usado {
		return errors.New("este token já foi utilizado")
	}
	if token.TokenExpirado() {
		return errors.New("este token já expirou")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("erro ao criptografar a nova senha")
	}
	if err := s.recuperacaoRepo.ConsumirEAtualizarSenha(ctx, token.ID, token.UsuarioID, string(hash)); err != nil {
		return errors.New("não foi possível consumir o token e atualizar a senha")
	}

	return nil
}
