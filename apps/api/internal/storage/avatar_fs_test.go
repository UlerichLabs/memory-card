package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAvatarFS(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "avatarfs-test-*")
	if err != nil {
		t.Fatalf("criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "nested", "avatares")
	fs, err := NewAvatarFS(subDir)
	if err != nil {
		t.Fatalf("NewAvatarFS falhou: %v", err)
	}

	nomeValido := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg"
	dados := []byte("fake_image_content")

	ctx := context.Background()

	if err := fs.Salvar(ctx, "invalido.png", dados); !errors.Is(err, ErrNomeArquivoInvalido) {
		t.Fatalf("esperava ErrNomeArquivoInvalido, obteve %v", err)
	}

	if err := fs.Salvar(ctx, nomeValido, dados); err != nil {
		t.Fatalf("Salvar falhou: %v", err)
	}

	rc, modTime, err := fs.Abrir(ctx, nomeValido)
	if err != nil {
		t.Fatalf("Abrir falhou: %v", err)
	}
	defer rc.Close()
	if modTime.IsZero() {
		t.Fatalf("modTime invalido")
	}
	lido, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ler conteudo: %v", err)
	}
	if string(lido) != string(dados) {
		t.Fatalf("conteudo divergente: %q", string(lido))
	}

	if _, _, err := fs.Abrir(ctx, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.jpg"); !errors.Is(err, ErrArquivoNaoEncontrado) {
		t.Fatalf("esperava ErrArquivoNaoEncontrado, obteve %v", err)
	}
	if _, _, err := fs.Abrir(ctx, "../tentativa"); !errors.Is(err, ErrArquivoNaoEncontrado) {
		t.Fatalf("esperava ErrArquivoNaoEncontrado para traversal, obteve %v", err)
	}

	if err := fs.Remover(ctx, nomeValido); err != nil {
		t.Fatalf("Remover falhou: %v", err)
	}
	if err := fs.Remover(ctx, nomeValido); err != nil {
		t.Fatalf("Remover idempotente falhou: %v", err)
	}
	if err := fs.Remover(ctx, "invalido"); err != nil {
		t.Fatalf("Remover com nome invalido falhou: %v", err)
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := fs.Salvar(cancelCtx, nomeValido, dados); err == nil {
		t.Fatalf("esperava erro de contexto cancelado")
	}
}
