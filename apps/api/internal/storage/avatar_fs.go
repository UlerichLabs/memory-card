// Package storage implementa adaptadores de persistencia em disco para arquivos.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	ErrNomeArquivoInvalido  = errors.New("storage.nome_arquivo_invalido")
	ErrArquivoNaoEncontrado = errors.New("storage.arquivo_nao_encontrado")
	regexNomeArquivoValido  = regexp.MustCompile(`^[a-f0-9]{32}\.jpg$`)
)

type AvatarFS struct {
	dir string
}

func NewAvatarFS(dir string) (*AvatarFS, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("criar diretorio de avatares: %w", err)
	}
	return &AvatarFS{dir: dir}, nil
}

func (s *AvatarFS) Salvar(ctx context.Context, nome string, dados []byte) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("salvar avatar cancelado: %w", err)
	}
	if !regexNomeArquivoValido.MatchString(nome) {
		return ErrNomeArquivoInvalido
	}
	caminho := filepath.Join(s.dir, nome)
	if err := os.WriteFile(caminho, dados, 0o644); err != nil {
		return fmt.Errorf("gravar avatar no disco: %w", err)
	}
	return nil
}

func (s *AvatarFS) Remover(ctx context.Context, nome string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("remover avatar cancelado: %w", err)
	}
	if !regexNomeArquivoValido.MatchString(nome) {
		return nil
	}
	caminho := filepath.Join(s.dir, nome)
	if err := os.Remove(caminho); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remover avatar do disco: %w", err)
	}
	return nil
}

func (s *AvatarFS) Abrir(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return nil, time.Time{}, fmt.Errorf("abrir avatar cancelado: %w", err)
	}
	if !regexNomeArquivoValido.MatchString(nome) {
		return nil, time.Time{}, ErrArquivoNaoEncontrado
	}
	caminho := filepath.Join(s.dir, nome)
	f, err := os.Open(caminho)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, time.Time{}, ErrArquivoNaoEncontrado
		}
		return nil, time.Time{}, fmt.Errorf("abrir avatar: %w", err)
	}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, time.Time{}, fmt.Errorf("obter info do avatar: %w", err)
	}
	return f, stat.ModTime(), nil
}
