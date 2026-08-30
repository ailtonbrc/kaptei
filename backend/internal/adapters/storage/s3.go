package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ConfiguracaoS3 struct {
	Regiao        string
	Bucket        string
	Endpoint      string
	AccessKey     string
	SecretKey     string
	BaseURL       string
	UsarPathStyle bool
}

type armazenamentoS3 struct {
	cliente *s3.Client
	bucket  string
	baseURL string
}

func NewArmazenamentoS3(ctx context.Context, cfg ConfiguracaoS3) (*armazenamentoS3, error) {
	if cfg.Regiao == "" || cfg.Bucket == "" || cfg.BaseURL == "" {
		return nil, errors.New("região, bucket e URL pública S3 são obrigatórios")
	}
	base, err := url.Parse(strings.TrimRight(cfg.BaseURL, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, errors.New("URL pública S3 inválida")
	}
	opcoes := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.Regiao)}
	if cfg.AccessKey != "" || cfg.SecretKey != "" {
		if cfg.AccessKey == "" || cfg.SecretKey == "" {
			return nil, errors.New("credenciais S3 incompletas")
		}
		opcoes = append(opcoes, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")))
	}
	configuracaoAWS, err := awsconfig.LoadDefaultConfig(ctx, opcoes...)
	if err != nil {
		return nil, fmt.Errorf("carregar configuração S3: %w", err)
	}
	cliente := s3.NewFromConfig(configuracaoAWS, func(opcao *s3.Options) {
		opcao.UsePathStyle = cfg.UsarPathStyle
		if strings.TrimSpace(cfg.Endpoint) != "" {
			opcao.BaseEndpoint = aws.String(strings.TrimRight(cfg.Endpoint, "/"))
		}
	})
	return &armazenamentoS3{cliente: cliente, bucket: cfg.Bucket, baseURL: base.String()}, nil
}

func (a *armazenamentoS3) Nome() string { return "s3" }

func (a *armazenamentoS3) Salvar(ctx context.Context, chave string, conteudo []byte, tipoConteudo string) (string, error) {
	if _, err := validarChaveObjeto(chave); err != nil {
		return "", err
	}
	soma := sha256.Sum256(conteudo)
	checksum := base64.StdEncoding.EncodeToString(soma[:])
	_, err := a.cliente.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket), Key: aws.String(chave), Body: bytes.NewReader(conteudo),
		ContentType: aws.String(tipoConteudo), CacheControl: aws.String("public, max-age=31536000, immutable"),
		ChecksumSHA256: aws.String(checksum),
	})
	if err != nil {
		return "", fmt.Errorf("enviar objeto ao S3: %w", err)
	}
	return construirURLPublica(a.baseURL, chave), nil
}

func (a *armazenamentoS3) Excluir(ctx context.Context, chave string) error {
	if _, err := validarChaveObjeto(chave); err != nil {
		return err
	}
	_, err := a.cliente.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(a.bucket), Key: aws.String(chave)})
	if err != nil {
		return fmt.Errorf("excluir objeto do S3: %w", err)
	}
	return nil
}

func validarChaveObjeto(chave string) (string, error) {
	chave = strings.ReplaceAll(strings.TrimSpace(chave), "\\", "/")
	limpa := path.Clean("/" + chave)
	if chave == "" || strings.Contains(chave, "..") || limpa == "/" {
		return "", errors.New("chave de objeto inválida")
	}
	return strings.TrimPrefix(limpa, "/"), nil
}

func construirURLPublica(baseURL, chave string) string {
	segmentos := strings.Split(strings.Trim(chave, "/"), "/")
	for indice := range segmentos {
		segmentos[indice] = url.PathEscape(segmentos[indice])
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.Join(segmentos, "/")
}
