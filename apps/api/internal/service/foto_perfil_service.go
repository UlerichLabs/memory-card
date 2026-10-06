package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

const (
	MaxFotoTamanho    = 2 * 1024 * 1024
	maxDimensoesLado  = 8000
	maxTotalMegapixel = 40_000_000
)

var (
	ErrFotoArquivoObrigatorio   = errors.New("foto.arquivo_obrigatorio")
	ErrFotoTipoNaoSuportado     = errors.New("foto.tipo_nao_suportado")
	ErrFotoArquivoMuitoGrande   = errors.New("foto.arquivo_muito_grande")
	ErrFotoImagemInvalida       = errors.New("foto.imagem_invalida")
	ErrFotoDimensoesExcessivas  = errors.New("foto.dimensoes_excessivas")
	ErrFotoJogoNaoEncontrado    = errors.New("foto.jogo_nao_encontrado")
	ErrFotoJogoSemCapa          = errors.New("foto.jogo_sem_capa")
	ErrFotoArquivoNaoEncontrado = errors.New("foto.arquivo_nao_encontrado")
	ErrFotoJogoIDInvalido       = errors.New("foto.jogo_id_invalido")

	regexArquivoValido = regexp.MustCompile(`^[a-f0-9]{32}\.jpg$`)
)

type ArmazenamentoAvatar interface {
	Salvar(ctx context.Context, nome string, dados []byte) error
	Remover(ctx context.Context, nome string) error
	Abrir(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error)
}

type FotoPerfilRepository interface {
	BuscarPerfil(ctx context.Context, id int32) (*repository.PerfilUsuario, error)
	BuscarAvatar(ctx context.Context, id int32) (*string, error)
	AtualizarAvatarUpload(ctx context.Context, id int32, nomeArquivo string) (*repository.PerfilUsuario, error)
	AtualizarAvatarCapa(ctx context.Context, id int32, jogoID int32) (*repository.PerfilUsuario, error)
	RemoverAvatar(ctx context.Context, id int32) (*repository.PerfilUsuario, error)
}

type FotoPerfilService struct {
	repo    FotoPerfilRepository
	storage ArmazenamentoAvatar
}

func NewFotoPerfilService(repo FotoPerfilRepository, storage ArmazenamentoAvatar) *FotoPerfilService {
	return &FotoPerfilService{
		repo:    repo,
		storage: storage,
	}
}

func (svc *FotoPerfilService) Upload(
	ctx context.Context,
	subject string,
	r io.Reader,
) (*repository.PerfilUsuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("upload avatar cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}
	if r == nil {
		return nil, ErrFotoArquivoObrigatorio
	}

	dados, err := io.ReadAll(io.LimitReader(r, MaxFotoTamanho+1))
	if err != nil {
		return nil, ErrFotoImagemInvalida
	}
	if len(dados) == 0 {
		return nil, ErrFotoArquivoObrigatorio
	}
	if len(dados) > MaxFotoTamanho {
		return nil, ErrFotoArquivoMuitoGrande
	}

	sniffLen := len(dados)
	if sniffLen > 512 {
		sniffLen = 512
	}
	contentType := http.DetectContentType(dados[:sniffLen])
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return nil, ErrFotoTipoNaoSuportado
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(dados))
	if err != nil {
		return nil, ErrFotoImagemInvalida
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, ErrFotoImagemInvalida
	}
	if cfg.Width > maxDimensoesLado || cfg.Height > maxDimensoesLado || (int64(cfg.Width)*int64(cfg.Height) > maxTotalMegapixel) {
		return nil, ErrFotoDimensoesExcessivas
	}

	srcImg, _, err := image.Decode(bytes.NewReader(dados))
	if err != nil {
		return nil, ErrFotoImagemInvalida
	}

	bounds := srcImg.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	side := w
	if h < side {
		side = h
	}
	x0 := bounds.Min.X + (w-side)/2
	y0 := bounds.Min.Y + (h-side)/2
	cropRect := image.Rect(x0, y0, x0+side, y0+side)

	dst := image.NewRGBA(image.Rect(0, 0, 512, 512))
	fundo := color.RGBA{R: 0x1A, G: 0x1B, B: 0x20, A: 0xFF}
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: fundo}, image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), srcImg, cropRect, xdraw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("codificar jpeg: %w", err)
	}
	conteudoProcessado := buf.Bytes()

	var aleatorio [16]byte
	if _, err := rand.Read(aleatorio[:]); err != nil {
		return nil, fmt.Errorf("gerar bytes aleatorios: %w", err)
	}
	nomeArquivo := hex.EncodeToString(aleatorio[:]) + ".jpg"

	if err := svc.storage.Salvar(ctx, nomeArquivo, conteudoProcessado); err != nil {
		return nil, fmt.Errorf("salvar avatar no storage: %w", err)
	}

	avatarAntigo, err := svc.repo.BuscarAvatar(ctx, int32(id))
	if err != nil {
		_ = svc.storage.Remover(ctx, nomeArquivo)
		return nil, fmt.Errorf("buscar avatar atual: %w", err)
	}

	perfil, err := svc.repo.AtualizarAvatarUpload(ctx, int32(id), nomeArquivo)
	if err != nil {
		_ = svc.storage.Remover(ctx, nomeArquivo)
		return nil, fmt.Errorf("atualizar avatar upload: %w", err)
	}

	if avatarAntigo != nil && *avatarAntigo != "" {
		if errRemover := svc.storage.Remover(ctx, *avatarAntigo); errRemover != nil {
			slog.WarnContext(ctx, "falha ao remover avatar antigo do disco", "arquivo", *avatarAntigo, "error", errRemover)
		}
	}

	return perfil, nil
}

func (svc *FotoPerfilService) DefinirCapa(
	ctx context.Context,
	subject string,
	jogoID int32,
) (*repository.PerfilUsuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("definir capa cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}
	if jogoID <= 0 {
		return nil, ErrFotoJogoIDInvalido
	}

	avatarAntigo, err := svc.repo.BuscarAvatar(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("buscar avatar atual: %w", err)
	}

	perfil, err := svc.repo.AtualizarAvatarCapa(ctx, int32(id), jogoID)
	if err != nil {
		if errors.Is(err, repository.ErrJogoParaCapaNaoEncontrado) {
			return nil, ErrFotoJogoNaoEncontrado
		}
		if errors.Is(err, repository.ErrJogoSemCapa) {
			return nil, ErrFotoJogoSemCapa
		}
		return nil, fmt.Errorf("atualizar avatar capa: %w", err)
	}

	if avatarAntigo != nil && *avatarAntigo != "" {
		if errRemover := svc.storage.Remover(ctx, *avatarAntigo); errRemover != nil {
			slog.WarnContext(ctx, "falha ao remover avatar antigo do disco", "arquivo", *avatarAntigo, "error", errRemover)
		}
	}

	return perfil, nil
}

func (svc *FotoPerfilService) RemoverFoto(
	ctx context.Context,
	subject string,
) (*repository.PerfilUsuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("remover foto cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}

	avatarAntigo, err := svc.repo.BuscarAvatar(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("buscar avatar atual: %w", err)
	}

	perfil, err := svc.repo.RemoverAvatar(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("remover avatar: %w", err)
	}

	if avatarAntigo != nil && *avatarAntigo != "" {
		if errRemover := svc.storage.Remover(ctx, *avatarAntigo); errRemover != nil {
			slog.WarnContext(ctx, "falha ao remover avatar do disco", "arquivo", *avatarAntigo, "error", errRemover)
		}
	}

	return perfil, nil
}

func (svc *FotoPerfilService) BuscarArquivo(
	ctx context.Context,
	nome string,
) (io.ReadSeekCloser, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return nil, time.Time{}, fmt.Errorf("buscar arquivo cancelado: %w", err)
	}
	if !regexArquivoValido.MatchString(nome) {
		return nil, time.Time{}, ErrFotoArquivoNaoEncontrado
	}

	f, modTime, err := svc.storage.Abrir(ctx, nome)
	if err != nil {
		return nil, time.Time{}, ErrFotoArquivoNaoEncontrado
	}

	return f, modTime, nil
}
