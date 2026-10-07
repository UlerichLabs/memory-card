package service

import (
	"bytes"
	"context"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type fotoPerfilRepoMock struct {
	buscarPerfil           func(context.Context, int32) (*repository.PerfilUsuario, error)
	buscarAvatar           func(context.Context, int32) (*string, error)
	atualizarAvatarUpload  func(context.Context, int32, string) (*repository.PerfilUsuario, error)
	atualizarAvatarCapa    func(context.Context, int32, int32) (*repository.PerfilUsuario, error)
	removerAvatar          func(context.Context, int32) (*repository.PerfilUsuario, error)
}

func (m fotoPerfilRepoMock) BuscarPerfil(ctx context.Context, id int32) (*repository.PerfilUsuario, error) {
	if m.buscarPerfil != nil {
		return m.buscarPerfil(ctx, id)
	}
	return nil, nil
}

func (m fotoPerfilRepoMock) BuscarAvatar(ctx context.Context, id int32) (*string, error) {
	if m.buscarAvatar != nil {
		return m.buscarAvatar(ctx, id)
	}
	return nil, nil
}

func (m fotoPerfilRepoMock) AtualizarAvatarUpload(ctx context.Context, id int32, nomeArquivo string) (*repository.PerfilUsuario, error) {
	if m.atualizarAvatarUpload != nil {
		return m.atualizarAvatarUpload(ctx, id, nomeArquivo)
	}
	return nil, nil
}

func (m fotoPerfilRepoMock) AtualizarAvatarCapa(ctx context.Context, id int32, jogoID int32) (*repository.PerfilUsuario, error) {
	if m.atualizarAvatarCapa != nil {
		return m.atualizarAvatarCapa(ctx, id, jogoID)
	}
	return nil, nil
}

func (m fotoPerfilRepoMock) RemoverAvatar(ctx context.Context, id int32) (*repository.PerfilUsuario, error) {
	if m.removerAvatar != nil {
		return m.removerAvatar(ctx, id)
	}
	return nil, nil
}

type storageMock struct {
	arquivos map[string][]byte
	salvar   func(ctx context.Context, nome string, dados []byte) error
	remover  func(ctx context.Context, nome string) error
	abrir    func(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error)
}

func newStorageMock() *storageMock {
	return &storageMock{
		arquivos: make(map[string][]byte),
	}
}

func (s *storageMock) Salvar(ctx context.Context, nome string, dados []byte) error {
	if s.salvar != nil {
		return s.salvar(ctx, nome, dados)
	}
	s.arquivos[nome] = dados
	return nil
}

func (s *storageMock) Remover(ctx context.Context, nome string) error {
	if s.remover != nil {
		return s.remover(ctx, nome)
	}
	delete(s.arquivos, nome)
	return nil
}

func (s *storageMock) Abrir(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error) {
	if s.abrir != nil {
		return s.abrir(ctx, nome)
	}
	dados, ok := s.arquivos[nome]
	if !ok {
		return nil, time.Time{}, ErrFotoArquivoNaoEncontrado
	}
	return &readSeekNopCloser{Reader: bytes.NewReader(dados)}, time.Now(), nil
}

type readSeekNopCloser struct {
	*bytes.Reader
}

func (r *readSeekNopCloser) Close() error {
	return nil
}

func gerarJPEGValido(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 50, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

func gerarPNGValido(w, h int, transparente bool) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if transparente {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
			} else {
				img.Set(x, y, color.RGBA{R: 50, G: 200, B: 50, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func gerarWebPValido() []byte {
	return []byte{
		'R', 'I', 'F', 'F',
		0x1a, 0x00, 0x00, 0x00,
		'W', 'E', 'B', 'P',
		'V', 'P', '8', 'L',
		0x0d, 0x00, 0x00, 0x00,
		0x2f, 0x00, 0x00, 0x00, 0x00, 0x07, 0x10, 0x11, 0x11, 0x88, 0x88, 0xfe, 0x07, 0x00,
	}
}

func gerarPNGComDimensoes(w, h uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})

	var ihdr bytes.Buffer
	ihdr.WriteString("IHDR")
	ihdr.Write([]byte{
		byte(w >> 24), byte(w >> 16), byte(w >> 8), byte(w),
		byte(h >> 24), byte(h >> 16), byte(h >> 8), byte(h),
		8, 2, 0, 0, 0,
	})

	crc := crc32.ChecksumIEEE(ihdr.Bytes())

	length := uint32(13)
	buf.Write([]byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)})
	buf.Write(ihdr.Bytes())
	buf.Write([]byte{byte(crc >> 24), byte(crc >> 16), byte(crc >> 8), byte(crc)})

	return buf.Bytes()
}

func TestFotoPerfilService_Upload_FormatosValidos(t *testing.T) {
	for _, tc := range []struct {
		name  string
		dados []byte
	}{
		{"JPEG valido", gerarJPEGValido(100, 80)},
		{"PNG valido", gerarPNGValido(80, 100, false)},
		{"WebP valido", gerarWebPValido()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage := newStorageMock()
			var nomeGravado string

			repo := fotoPerfilRepoMock{
				buscarAvatar: func(ctx context.Context, id int32) (*string, error) {
					return nil, nil
				},
				atualizarAvatarUpload: func(ctx context.Context, id int32, nome string) (*repository.PerfilUsuario, error) {
					nomeGravado = nome
					tipo := "upload"
					url := "/api/v1/avatares/" + nome
					return &repository.PerfilUsuario{
						Nome:       "Lucas",
						AvatarURL:  &url,
						AvatarTipo: &tipo,
					}, nil
				},
			}

			svc := NewFotoPerfilService(repo, storage)
			perfil, err := svc.Upload(context.Background(), "42", bytes.NewReader(tc.dados))
			if err != nil {
				t.Fatalf("Upload falhou: %v", err)
			}
			if perfil == nil || perfil.AvatarURL == nil || perfil.AvatarTipo == nil {
				t.Fatalf("perfil retornado invalido: %+v", perfil)
			}
			if *perfil.AvatarTipo != "upload" {
				t.Fatalf("avatar_tipo esperado upload, obteve %s", *perfil.AvatarTipo)
			}
			if !strings.HasPrefix(*perfil.AvatarURL, "/api/v1/avatares/") {
				t.Fatalf("avatar_url esperado caminho relativo, obteve %s", *perfil.AvatarURL)
			}

			conteudo, ok := storage.arquivos[nomeGravado]
			if !ok {
				t.Fatalf("arquivo %s nao foi salvo no storage", nomeGravado)
			}

			cfg, format, err := image.DecodeConfig(bytes.NewReader(conteudo))
			if err != nil {
				t.Fatalf("falha ao ler config da imagem gravada: %v", err)
			}
			if format != "jpeg" {
				t.Fatalf("formato esperado jpeg, obteve %s", format)
			}
			if cfg.Width != 512 || cfg.Height != 512 {
				t.Fatalf("dimensoes esperadas 512x512, obteve %dx%d", cfg.Width, cfg.Height)
			}
		})
	}
}

func TestFotoPerfilService_Upload_ExtensaoPNG_ConteudoTexto(t *testing.T) {
	storage := newStorageMock()
	svc := NewFotoPerfilService(fotoPerfilRepoMock{}, storage)

	_, err := svc.Upload(context.Background(), "42", strings.NewReader("isto eh um texto puro"))
	if !errors.Is(err, ErrFotoTipoNaoSuportado) {
		t.Fatalf("esperado ErrFotoTipoNaoSuportado, obteve %v", err)
	}
}

func TestFotoPerfilService_Upload_ImagemCorrompida(t *testing.T) {
	storage := newStorageMock()
	svc := NewFotoPerfilService(fotoPerfilRepoMock{}, storage)

	corrompido := append([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F'}, []byte("corrupted_garbage_bytes")...)
	_, err := svc.Upload(context.Background(), "42", bytes.NewReader(corrompido))
	if !errors.Is(err, ErrFotoImagemInvalida) {
		t.Fatalf("esperado ErrFotoImagemInvalida, obteve %v", err)
	}
}

func TestFotoPerfilService_Upload_DimensoesAcimaDoLimite(t *testing.T) {
	storage := newStorageMock()
	svc := NewFotoPerfilService(fotoPerfilRepoMock{}, storage)

	pngGrande := gerarPNGComDimensoes(8001, 100)
	_, err := svc.Upload(context.Background(), "42", bytes.NewReader(pngGrande))
	if !errors.Is(err, ErrFotoDimensoesExcessivas) {
		t.Fatalf("esperado ErrFotoDimensoesExcessivas para 8001px de lado, obteve %v", err)
	}

	pngMegapixel := gerarPNGComDimensoes(7000, 6000)
	_, err = svc.Upload(context.Background(), "42", bytes.NewReader(pngMegapixel))
	if !errors.Is(err, ErrFotoDimensoesExcessivas) {
		t.Fatalf("esperado ErrFotoDimensoesExcessivas para 42 megapixels, obteve %v", err)
	}
}

func TestFotoPerfilService_Upload_SubstituicaoEApagarAntigo(t *testing.T) {
	storage := newStorageMock()
	arquivoAntigo := "11111111111111111111111111111111.jpg"
	storage.arquivos[arquivoAntigo] = []byte("antigo")

	var nomeNovo string
	repo := fotoPerfilRepoMock{
		buscarAvatar: func(ctx context.Context, id int32) (*string, error) {
			return &arquivoAntigo, nil
		},
		atualizarAvatarUpload: func(ctx context.Context, id int32, nome string) (*repository.PerfilUsuario, error) {
			nomeNovo = nome
			tipo := "upload"
			url := "/api/v1/avatares/" + nome
			return &repository.PerfilUsuario{
				Nome:       "Lucas",
				AvatarURL:  &url,
				AvatarTipo: &tipo,
			}, nil
		},
	}

	svc := NewFotoPerfilService(repo, storage)
	_, err := svc.Upload(context.Background(), "42", bytes.NewReader(gerarJPEGValido(50, 50)))
	if err != nil {
		t.Fatalf("Upload falhou: %v", err)
	}

	if _, ok := storage.arquivos[arquivoAntigo]; ok {
		t.Fatalf("arquivo antigo deveria ter sido removido do storage")
	}
	if _, ok := storage.arquivos[nomeNovo]; !ok {
		t.Fatalf("arquivo novo deveria estar no storage")
	}
}

func TestFotoPerfilService_Upload_FalhaBanco_ApagaArquivoNovo(t *testing.T) {
	storage := newStorageMock()
	var nomeTentado string

	storage.salvar = func(ctx context.Context, nome string, dados []byte) error {
		nomeTentado = nome
		storage.arquivos[nome] = dados
		return nil
	}

	repo := fotoPerfilRepoMock{
		buscarAvatar: func(ctx context.Context, id int32) (*string, error) {
			return nil, nil
		},
		atualizarAvatarUpload: func(ctx context.Context, id int32, nome string) (*repository.PerfilUsuario, error) {
			return nil, errors.New("db failure")
		},
	}

	svc := NewFotoPerfilService(repo, storage)
	_, err := svc.Upload(context.Background(), "42", bytes.NewReader(gerarJPEGValido(50, 50)))
	if err == nil {
		t.Fatalf("esperava erro ao falhar banco")
	}

	if _, ok := storage.arquivos[nomeTentado]; ok {
		t.Fatalf("arquivo novo deveria ter sido apagado apos falha no banco")
	}
}

func TestFotoPerfilService_DefinirCapa_Cenarios(t *testing.T) {
	t.Run("sucesso remove arquivo antigo", func(t *testing.T) {
		storage := newStorageMock()
		arquivoAntigo := "22222222222222222222222222222222.jpg"
		storage.arquivos[arquivoAntigo] = []byte("antigo")

		repo := fotoPerfilRepoMock{
			buscarAvatar: func(ctx context.Context, id int32) (*string, error) {
				return &arquivoAntigo, nil
			},
			atualizarAvatarCapa: func(ctx context.Context, id int32, jogoID int32) (*repository.PerfilUsuario, error) {
				tipo := "jogo"
				capa := "https://images.igdb.com/cover.jpg"
				return &repository.PerfilUsuario{
					Nome:       "Lucas",
					AvatarURL:  &capa,
					AvatarTipo: &tipo,
				}, nil
			},
		}

		svc := NewFotoPerfilService(repo, storage)
		perfil, err := svc.DefinirCapa(context.Background(), "42", 10)
		if err != nil {
			t.Fatalf("DefinirCapa falhou: %v", err)
		}
		if perfil == nil || *perfil.AvatarTipo != "jogo" {
			t.Fatalf("perfil inesperado: %+v", perfil)
		}
		if _, ok := storage.arquivos[arquivoAntigo]; ok {
			t.Fatalf("arquivo antigo deveria ter sido removido do storage")
		}
	})

	t.Run("jogo nao encontrado", func(t *testing.T) {
		repo := fotoPerfilRepoMock{
			buscarAvatar: func(ctx context.Context, id int32) (*string, error) { return nil, nil },
			atualizarAvatarCapa: func(ctx context.Context, id int32, jogoID int32) (*repository.PerfilUsuario, error) {
				return nil, repository.ErrJogoParaCapaNaoEncontrado
			},
		}
		svc := NewFotoPerfilService(repo, newStorageMock())
		_, err := svc.DefinirCapa(context.Background(), "42", 999)
		if !errors.Is(err, ErrFotoJogoNaoEncontrado) {
			t.Fatalf("esperava ErrFotoJogoNaoEncontrado, obteve %v", err)
		}
	})

	t.Run("jogo sem capa", func(t *testing.T) {
		repo := fotoPerfilRepoMock{
			buscarAvatar: func(ctx context.Context, id int32) (*string, error) { return nil, nil },
			atualizarAvatarCapa: func(ctx context.Context, id int32, jogoID int32) (*repository.PerfilUsuario, error) {
				return nil, repository.ErrJogoSemCapa
			},
		}
		svc := NewFotoPerfilService(repo, newStorageMock())
		_, err := svc.DefinirCapa(context.Background(), "42", 10)
		if !errors.Is(err, ErrFotoJogoSemCapa) {
			t.Fatalf("esperava ErrFotoJogoSemCapa, obteve %v", err)
		}
	})
}

func TestFotoPerfilService_RemoverFoto_Idempotente(t *testing.T) {
	storage := newStorageMock()
	arquivoAntigo := "33333333333333333333333333333333.jpg"
	storage.arquivos[arquivoAntigo] = []byte("antigo")

	repo := fotoPerfilRepoMock{
		buscarAvatar: func(ctx context.Context, id int32) (*string, error) {
			return &arquivoAntigo, nil
		},
		removerAvatar: func(ctx context.Context, id int32) (*repository.PerfilUsuario, error) {
			return &repository.PerfilUsuario{
				Nome:       "Lucas",
				AvatarURL:  nil,
				AvatarTipo: nil,
			}, nil
		},
	}

	svc := NewFotoPerfilService(repo, storage)
	perfil, err := svc.RemoverFoto(context.Background(), "42")
	if err != nil {
		t.Fatalf("RemoverFoto falhou: %v", err)
	}
	if perfil.AvatarURL != nil || perfil.AvatarTipo != nil {
		t.Fatalf("foto deveria estar nula no perfil: %+v", perfil)
	}
	if _, ok := storage.arquivos[arquivoAntigo]; ok {
		t.Fatalf("arquivo deveria ter sido removido do storage")
	}

	repo.buscarAvatar = func(ctx context.Context, id int32) (*string, error) {
		return nil, nil
	}
	perfil2, err := svc.RemoverFoto(context.Background(), "42")
	if err != nil {
		t.Fatalf("segunda chamada RemoverFoto falhou: %v", err)
	}
	if perfil2.AvatarURL != nil || perfil2.AvatarTipo != nil {
		t.Fatalf("foto deveria continuar nula")
	}
}

func TestFotoPerfilService_TransparenciaFundo(t *testing.T) {
	storage := newStorageMock()
	var nomeGravado string

	repo := fotoPerfilRepoMock{
		buscarAvatar: func(ctx context.Context, id int32) (*string, error) { return nil, nil },
		atualizarAvatarUpload: func(ctx context.Context, id int32, nome string) (*repository.PerfilUsuario, error) {
			nomeGravado = nome
			tipo := "upload"
			url := "/api/v1/avatares/" + nome
			return &repository.PerfilUsuario{
				Nome:       "Lucas",
				AvatarURL:  &url,
				AvatarTipo: &tipo,
			}, nil
		},
	}

	svc := NewFotoPerfilService(repo, storage)
	pngTransparente := gerarPNGValido(64, 64, true)
	_, err := svc.Upload(context.Background(), "42", bytes.NewReader(pngTransparente))
	if err != nil {
		t.Fatalf("Upload falhou: %v", err)
	}

	conteudo := storage.arquivos[nomeGravado]
	img, err := jpeg.Decode(bytes.NewReader(conteudo))
	if err != nil {
		t.Fatalf("falha ao decodificar jpeg: %v", err)
	}

	r, g, b, _ := img.At(256, 256).RGBA()
	r8 := int(r >> 8)
	g8 := int(g >> 8)
	b8 := int(b >> 8)

	if math.Abs(float64(r8-0x1A)) > 4 || math.Abs(float64(g8-0x1B)) > 4 || math.Abs(float64(b8-0x20)) > 4 {
		t.Fatalf("pixel esperado perto de #1A1B20 (26, 27, 32), obteve (%d, %d, %d)", r8, g8, b8)
	}
}
