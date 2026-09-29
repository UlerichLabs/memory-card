// Package service implementa a logica de negocio da aplicacao.
package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type mockListasRepo struct {
	criarFn                     func(ctx context.Context, params repository.CriarListaParams) (*repository.Lista, error)
	criarComItensFn             func(ctx context.Context, params repository.CriarListaComItensParams) (*repository.Lista, []*repository.ListaItem, error)
	buscarPorIDFn               func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error)
	listarPorUsuarioFn          func(ctx context.Context, usuarioID int32) ([]*repository.Lista, error)
	atualizarFn                 func(ctx context.Context, params repository.AtualizarListaParams) (*repository.Lista, error)
	excluirFn                   func(ctx context.Context, id int64, usuarioID int32) error
	criarItemFn                 func(ctx context.Context, listaID int64, usuarioID int32, params repository.CriarItemParams) (*repository.ListaItem, error)
	buscarItemPorIDFn           func(ctx context.Context, id int64, listaID int64, usuarioID int32) (*repository.ListaItem, error)
	listarItensPorListaFn       func(ctx context.Context, listaID int64, usuarioID int32) ([]*repository.ListaItem, error)
	listarTodosItensDoUsuarioFn func(ctx context.Context, usuarioID int32) ([]*repository.ListaItem, error)
	excluirItemERecompactarFn   func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error
	reordenarItensFn            func(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*repository.ListaItem, error)
	associarJogoZeradoFn        func(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*repository.ListaItem, error)
	desassociarJogoZeradoFn     func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*repository.ListaItem, error)
	sincronizarFranquiaFn       func(ctx context.Context, listaID int64, usuarioID int32, novosItens []repository.CriarItemParams) (int, error)
	listarJogosZeradosUsuarioFn func(ctx context.Context, usuarioID int32) ([]*repository.JogoZeradoResumo, error)
	buscarJogoZeradoUsuarioFn   func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZeradoResumo, error)
}

func (m *mockListasRepo) Criar(ctx context.Context, params repository.CriarListaParams) (*repository.Lista, error) {
	if m.criarFn != nil {
		return m.criarFn(ctx, params)
	}
	return &repository.Lista{ID: 1, UsuarioID: params.UsuarioID, Tipo: params.Tipo, Nome: params.Nome, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockListasRepo) CriarComItens(ctx context.Context, params repository.CriarListaComItensParams) (*repository.Lista, []*repository.ListaItem, error) {
	if m.criarComItensFn != nil {
		return m.criarComItensFn(ctx, params)
	}
	l := &repository.Lista{ID: 1, UsuarioID: params.Lista.UsuarioID, Tipo: params.Lista.Tipo, Nome: params.Lista.Nome, RegraTipo: params.Lista.RegraTipo, RegraValor: params.Lista.RegraValor, RegraIgdbID: params.Lista.RegraIgdbID, Meta: params.Lista.Meta, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return l, nil, nil
}

func (m *mockListasRepo) BuscarPorID(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
	if m.buscarPorIDFn != nil {
		return m.buscarPorIDFn(ctx, id, usuarioID)
	}
	return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "fila", Nome: "Fila", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockListasRepo) ListarPorUsuario(ctx context.Context, usuarioID int32) ([]*repository.Lista, error) {
	if m.listarPorUsuarioFn != nil {
		return m.listarPorUsuarioFn(ctx, usuarioID)
	}
	return []*repository.Lista{}, nil
}

func (m *mockListasRepo) Atualizar(ctx context.Context, params repository.AtualizarListaParams) (*repository.Lista, error) {
	if m.atualizarFn != nil {
		return m.atualizarFn(ctx, params)
	}
	return &repository.Lista{ID: params.ID, UsuarioID: params.UsuarioID, Nome: params.Nome, Descricao: params.Descricao, Meta: params.Meta, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockListasRepo) Excluir(ctx context.Context, id int64, usuarioID int32) error {
	if m.excluirFn != nil {
		return m.excluirFn(ctx, id, usuarioID)
	}
	return nil
}

func (m *mockListasRepo) CriarItem(ctx context.Context, listaID int64, usuarioID int32, params repository.CriarItemParams) (*repository.ListaItem, error) {
	if m.criarItemFn != nil {
		return m.criarItemFn(ctx, listaID, usuarioID, params)
	}
	return &repository.ListaItem{ID: 10, ListaID: listaID, IgdbID: params.IgdbID, Nome: params.Nome, Posicao: 1, CreatedAt: time.Now()}, nil
}

func (m *mockListasRepo) BuscarItemPorID(ctx context.Context, id int64, listaID int64, usuarioID int32) (*repository.ListaItem, error) {
	if m.buscarItemPorIDFn != nil {
		return m.buscarItemPorIDFn(ctx, id, listaID, usuarioID)
	}
	return &repository.ListaItem{ID: id, ListaID: listaID, Nome: "Item", Posicao: 1, CreatedAt: time.Now()}, nil
}

func (m *mockListasRepo) ListarItensPorLista(ctx context.Context, listaID int64, usuarioID int32) ([]*repository.ListaItem, error) {
	if m.listarItensPorListaFn != nil {
		return m.listarItensPorListaFn(ctx, listaID, usuarioID)
	}
	return []*repository.ListaItem{}, nil
}

func (m *mockListasRepo) ListarTodosItensDoUsuario(ctx context.Context, usuarioID int32) ([]*repository.ListaItem, error) {
	if m.listarTodosItensDoUsuarioFn != nil {
		return m.listarTodosItensDoUsuarioFn(ctx, usuarioID)
	}
	return []*repository.ListaItem{}, nil
}

func (m *mockListasRepo) ExcluirItemERecompactar(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error {
	if m.excluirItemERecompactarFn != nil {
		return m.excluirItemERecompactarFn(ctx, itemID, listaID, usuarioID)
	}
	return nil
}

func (m *mockListasRepo) ReordenarItens(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*repository.ListaItem, error) {
	if m.reordenarItensFn != nil {
		return m.reordenarItensFn(ctx, listaID, usuarioID, itemIDs)
	}
	res := make([]*repository.ListaItem, 0, len(itemIDs))
	for idx, id := range itemIDs {
		res = append(res, &repository.ListaItem{ID: id, ListaID: listaID, Nome: "Item", Posicao: idx + 1, CreatedAt: time.Now()})
	}
	return res, nil
}

func (m *mockListasRepo) AssociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*repository.ListaItem, error) {
	if m.associarJogoZeradoFn != nil {
		return m.associarJogoZeradoFn(ctx, itemID, listaID, usuarioID, jogoZeradoID)
	}
	return &repository.ListaItem{ID: itemID, ListaID: listaID, Nome: "Item", Posicao: 1, JogoZeradoID: &jogoZeradoID, CreatedAt: time.Now()}, nil
}

func (m *mockListasRepo) DesassociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*repository.ListaItem, error) {
	if m.desassociarJogoZeradoFn != nil {
		return m.desassociarJogoZeradoFn(ctx, itemID, listaID, usuarioID)
	}
	return &repository.ListaItem{ID: itemID, ListaID: listaID, Nome: "Item", Posicao: 1, JogoZeradoID: nil, CreatedAt: time.Now()}, nil
}

func (m *mockListasRepo) SincronizarFranquia(ctx context.Context, listaID int64, usuarioID int32, novosItens []repository.CriarItemParams) (int, error) {
	if m.sincronizarFranquiaFn != nil {
		return m.sincronizarFranquiaFn(ctx, listaID, usuarioID, novosItens)
	}
	return len(novosItens), nil
}

func (m *mockListasRepo) ListarJogosZeradosUsuario(ctx context.Context, usuarioID int32) ([]*repository.JogoZeradoResumo, error) {
	if m.listarJogosZeradosUsuarioFn != nil {
		return m.listarJogosZeradosUsuarioFn(ctx, usuarioID)
	}
	return []*repository.JogoZeradoResumo{}, nil
}

func (m *mockListasRepo) BuscarJogoZeradoUsuario(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZeradoResumo, error) {
	if m.buscarJogoZeradoUsuarioFn != nil {
		return m.buscarJogoZeradoUsuarioFn(ctx, id, usuarioID)
	}
	return &repository.JogoZeradoResumo{ID: id, UsuarioID: usuarioID, Nome: "Jogo", Console: "SNES", Genero: "RPG", FinalizadoEm: time.Now(), Nota: 10}, nil
}

type mockListasIGDB struct {
	obterFranquiaFn            func(ctx context.Context, id int64) (*igdbclient.Franchise, error)
	jogosDaFranquiaFn          func(ctx context.Context, id int64) ([]igdbclient.Game, error)
	atualizarJogosDaFranquiaFn func(ctx context.Context, id int64) ([]igdbclient.Game, error)
}

func (m *mockListasIGDB) ObterFranquia(ctx context.Context, id int64) (*igdbclient.Franchise, error) {
	if m.obterFranquiaFn != nil {
		return m.obterFranquiaFn(ctx, id)
	}
	return &igdbclient.Franchise{ID: id, Name: "The Legend of Zelda"}, nil
}

func (m *mockListasIGDB) JogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	if m.jogosDaFranquiaFn != nil {
		return m.jogosDaFranquiaFn(ctx, id)
	}
	return []igdbclient.Game{}, nil
}

func (m *mockListasIGDB) AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	if m.atualizarJogosDaFranquiaFn != nil {
		return m.atualizarJogosDaFranquiaFn(ctx, id)
	}
	return []igdbclient.Game{}, nil
}

func TestListasService_CriarLista_Validacoes(t *testing.T) {
	tests := []struct {
		name    string
		input   CriarListaInput
		wantErr error
	}{
		{
			name:    "nome vazio",
			input:   CriarListaInput{UsuarioID: 1, Tipo: "fila", Nome: "   "},
			wantErr: ErrListaNomeObrigatorio,
		},
		{
			name:    "nome longo",
			input:   CriarListaInput{UsuarioID: 1, Tipo: "fila", Nome: strings.Repeat("a", 101)},
			wantErr: ErrListaNomeInvalido,
		},
		{
			name: "descricao longa",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "fila",
				Nome:      "Fila",
				Descricao: func() *string { s := strings.Repeat("d", 201); return &s }(),
			},
			wantErr: ErrListaDescricaoMuitoLonga,
		},
		{
			name:    "tipo invalido",
			input:   CriarListaInput{UsuarioID: 1, Tipo: "invalido", Nome: "Minha Lista"},
			wantErr: ErrListaTipoInvalido,
		},
		{
			name: "fila com regra",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "fila",
				Nome:      "Fila",
				Regra:     &CriarListaRegraInput{Tipo: "manual"},
			},
			wantErr: ErrListaRegraInvalida,
		},
		{
			name: "fila com meta",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "fila",
				Nome:      "Fila",
				Meta:      func() *int { m := 10; return &m }(),
			},
			wantErr: ErrListaRegraInvalida,
		},
		{
			name:    "desafio sem regra",
			input:   CriarListaInput{UsuarioID: 1, Tipo: "desafio", Nome: "Desafio"},
			wantErr: ErrListaRegraInvalida,
		},
		{
			name: "desafio regra tipo invalida",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "Desafio",
				Regra:     &CriarListaRegraInput{Tipo: "desconhecido"},
			},
			wantErr: ErrListaRegraInvalida,
		},
		{
			name: "franquia sem igdb_id",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "Zelda",
				Regra:     &CriarListaRegraInput{Tipo: "franquia"},
			},
			wantErr: ErrListaRegraInvalida,
		},
		{
			name: "franquia meta invalida zero",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "Zelda",
				Regra:     &CriarListaRegraInput{Tipo: "franquia", IgdbID: func() *int32 { id := int32(106); return &id }()},
				Meta:      func() *int { m := 0; return &m }(),
			},
			wantErr: ErrListaMetaInvalida,
		},
		{
			name: "franquia meta invalida maior que 10000",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "Zelda",
				Regra:     &CriarListaRegraInput{Tipo: "franquia", IgdbID: func() *int32 { id := int32(106); return &id }()},
				Meta:      func() *int { m := 10001; return &m }(),
			},
			wantErr: ErrListaMetaInvalida,
		},
		{
			name: "plataforma sem valor",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "SNES",
				Regra:     &CriarListaRegraInput{Tipo: "plataforma", Valor: ""},
				Meta:      func() *int { m := 5; return &m }(),
			},
			wantErr: ErrListaRegraInvalida,
		},
		{
			name: "plataforma sem meta",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "SNES",
				Regra:     &CriarListaRegraInput{Tipo: "plataforma", Valor: "SNES"},
			},
			wantErr: ErrListaMetaInvalida,
		},
		{
			name: "genero sem valor",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "RPG",
				Regra:     &CriarListaRegraInput{Tipo: "genero", Valor: "   "},
				Meta:      func() *int { m := 5; return &m }(),
			},
			wantErr: ErrListaRegraInvalida,
		},
		{
			name: "genero meta invalida",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "RPG",
				Regra:     &CriarListaRegraInput{Tipo: "genero", Valor: "RPG"},
				Meta:      func() *int { m := -1; return &m }(),
			},
			wantErr: ErrListaMetaInvalida,
		},
		{
			name: "manual meta negativa",
			input: CriarListaInput{
				UsuarioID: 1,
				Tipo:      "desafio",
				Nome:      "Manual",
				Regra:     &CriarListaRegraInput{Tipo: "manual"},
				Meta:      func() *int { m := 0; return &m }(),
			},
			wantErr: ErrListaMetaInvalida,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewListasService(&mockListasRepo{}, &mockListasIGDB{})
			_, err := svc.CriarLista(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("esperava erro %v, obteve: %v", tt.wantErr, err)
			}
		})
	}
}

func TestListasService_CriarLista_TiposSucesso(t *testing.T) {
	t.Run("criar fila", func(t *testing.T) {
		repo := &mockListasRepo{
			criarFn: func(ctx context.Context, params repository.CriarListaParams) (*repository.Lista, error) {
				return &repository.Lista{ID: 10, UsuarioID: params.UsuarioID, Tipo: "fila", Nome: params.Nome, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
			buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
				return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "fila", Nome: "Fila 1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
		}
		svc := NewListasService(repo, &mockListasIGDB{})
		res, err := svc.CriarLista(context.Background(), CriarListaInput{
			UsuarioID: 1,
			Tipo:      "fila",
			Nome:      "Fila 1",
		})
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if res.Tipo != "fila" || res.Progresso != nil {
			t.Fatalf("esperava fila sem progresso, obteve %+v", res)
		}
	})

	t.Run("criar desafio franquia", func(t *testing.T) {
		created := false
		rel1 := int64(1000)
		rel2 := int64(2000)
		repo := &mockListasRepo{
			criarComItensFn: func(ctx context.Context, params repository.CriarListaComItensParams) (*repository.Lista, []*repository.ListaItem, error) {
				created = true
				if len(params.Itens) != 2 {
					t.Fatalf("esperava 2 itens, obteve %d", len(params.Itens))
				}
				if params.Itens[0].Nome != "Zelda 1" || params.Itens[1].Nome != "Zelda 2" {
					t.Fatalf("ordem incorreta de lancamento: %+v", params.Itens)
				}
				regra := "franquia"
				return &repository.Lista{ID: 20, UsuarioID: params.Lista.UsuarioID, Tipo: "desafio", Nome: params.Lista.Nome, RegraTipo: &regra, RegraValor: params.Lista.RegraValor, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil, nil
			},
			buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
				regra := "franquia"
				nomeFranquia := "The Legend of Zelda"
				return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "desafio", Nome: "Zelda", RegraTipo: &regra, RegraValor: &nomeFranquia, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
			listarItensPorListaFn: func(ctx context.Context, listaID int64, usuarioID int32) ([]*repository.ListaItem, error) {
				id1 := int32(101)
				id2 := int32(102)
				return []*repository.ListaItem{
					{ID: 1, ListaID: listaID, IgdbID: &id1, Nome: "Zelda 1", Posicao: 1, CreatedAt: time.Now()},
					{ID: 2, ListaID: listaID, IgdbID: &id2, Nome: "Zelda 2", Posicao: 2, CreatedAt: time.Now()},
				}, nil
			},
		}
		igdb := &mockListasIGDB{
			obterFranquiaFn: func(ctx context.Context, id int64) (*igdbclient.Franchise, error) {
				return &igdbclient.Franchise{ID: id, Name: "The Legend of Zelda"}, nil
			},
			jogosDaFranquiaFn: func(ctx context.Context, id int64) ([]igdbclient.Game, error) {
				return []igdbclient.Game{
					{ID: 102, Name: "Zelda 2", FirstReleaseDate: &rel2},
					{ID: 101, Name: "Zelda 1", FirstReleaseDate: &rel1},
				}, nil
			},
		}
		svc := NewListasService(repo, igdb)
		igdbID := int32(106)
		res, err := svc.CriarLista(context.Background(), CriarListaInput{
			UsuarioID: 1,
			Tipo:      "desafio",
			Nome:      "Zelda",
			Regra:     &CriarListaRegraInput{Tipo: "franquia", IgdbID: &igdbID},
		})
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if !created || res.TotalItens != 2 || res.Progresso == nil {
			t.Fatalf("resultado inesperado: %+v", res)
		}
	})

	t.Run("criar desafio plataforma", func(t *testing.T) {
		repo := &mockListasRepo{
			criarFn: func(ctx context.Context, params repository.CriarListaParams) (*repository.Lista, error) {
				return &repository.Lista{ID: 30, UsuarioID: params.UsuarioID, Tipo: "desafio", Nome: params.Nome, RegraTipo: params.RegraTipo, RegraValor: params.RegraValor, Meta: params.Meta, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
			buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
				regra := "plataforma"
				val := "Super Nintendo"
				meta := 3
				return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "desafio", Nome: "SNES 3", RegraTipo: &regra, RegraValor: &val, Meta: &meta, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
		}
		svc := NewListasService(repo, &mockListasIGDB{})
		meta := 3
		res, err := svc.CriarLista(context.Background(), CriarListaInput{
			UsuarioID: 1,
			Tipo:      "desafio",
			Nome:      "SNES 3",
			Regra:     &CriarListaRegraInput{Tipo: "plataforma", Valor: "Super Nintendo"},
			Meta:      &meta,
		})
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if res.Progresso == nil || res.Progresso.Meta != 3 {
			t.Fatalf("progresso incorreto: %+v", res.Progresso)
		}
	})
}

func TestListasService_CriarLista_FranquiaNaoEncontrada(t *testing.T) {
	igdb := &mockListasIGDB{
		obterFranquiaFn: func(ctx context.Context, id int64) (*igdbclient.Franchise, error) {
			return nil, ErrJogoIGDBNaoEncontrado
		},
	}
	svc := NewListasService(&mockListasRepo{}, igdb)
	igdbID := int32(99999)
	_, err := svc.CriarLista(context.Background(), CriarListaInput{
		UsuarioID: 1,
		Tipo:      "desafio",
		Nome:      "Inexistente",
		Regra:     &CriarListaRegraInput{Tipo: "franquia", IgdbID: &igdbID},
	})
	if !errors.Is(err, ErrListaFranquiaNaoEncontrada) {
		t.Fatalf("esperava ErrListaFranquiaNaoEncontrada, obteve %v", err)
	}
}

func TestListasService_AtualizarLista_CampoImutavel(t *testing.T) {
	svc := NewListasService(&mockListasRepo{}, &mockListasIGDB{})
	tipo := "desafio"
	_, err := svc.AtualizarLista(context.Background(), AtualizarListaInput{
		ID:        1,
		UsuarioID: 1,
		Tipo:      &tipo,
	})
	if !errors.Is(err, ErrListaCampoImutavel) {
		t.Fatalf("esperava ErrListaCampoImutavel, obteve %v", err)
	}

	_, err = svc.AtualizarLista(context.Background(), AtualizarListaInput{
		ID:        1,
		UsuarioID: 1,
		Regra:     &CriarListaRegraInput{Tipo: "manual"},
	})
	if !errors.Is(err, ErrListaCampoImutavel) {
		t.Fatalf("esperava ErrListaCampoImutavel para regra, obteve %v", err)
	}
}

func TestListasService_AdicionarItem_ContagemEResto(t *testing.T) {
	t.Run("item em desafio de contagem", func(t *testing.T) {
		repo := &mockListasRepo{
			buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
				regra := "plataforma"
				return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "desafio", RegraTipo: &regra, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
		}
		svc := NewListasService(repo, &mockListasIGDB{})
		_, err := svc.AdicionarItem(context.Background(), AdicionarItemInput{
			ListaID:   1,
			UsuarioID: 1,
			Nome:      "Chrono Trigger",
		})
		if !errors.Is(err, ErrListaItensNaoPermitidos) {
			t.Fatalf("esperava ErrListaItensNaoPermitidos, obteve %v", err)
		}
	})

	t.Run("item duplicado 23505", func(t *testing.T) {
		repo := &mockListasRepo{
			buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
				return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "fila", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
			criarItemFn: func(ctx context.Context, listaID int64, usuarioID int32, params repository.CriarItemParams) (*repository.ListaItem, error) {
				return nil, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
			},
		}
		svc := NewListasService(repo, &mockListasIGDB{})
		_, err := svc.AdicionarItem(context.Background(), AdicionarItemInput{
			ListaID:   1,
			UsuarioID: 1,
			Nome:      "Super Mario",
		})
		if !errors.Is(err, ErrListaItemDuplicado) {
			t.Fatalf("esperava ErrListaItemDuplicado, obteve %v", err)
		}
	})
}

func TestListasService_ReordenarItens_Validacao(t *testing.T) {
	repo := &mockListasRepo{
		buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
			return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "fila", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
		listarItensPorListaFn: func(ctx context.Context, listaID int64, usuarioID int32) ([]*repository.ListaItem, error) {
			return []*repository.ListaItem{
				{ID: 1, ListaID: listaID, Posicao: 1},
				{ID: 2, ListaID: listaID, Posicao: 2},
				{ID: 3, ListaID: listaID, Posicao: 3},
			}, nil
		},
	}
	svc := NewListasService(repo, &mockListasIGDB{})

	t.Run("tamanho diferente", func(t *testing.T) {
		_, err := svc.ReordenarItens(context.Background(), 1, 1, []int64{1, 2})
		if !errors.Is(err, ErrListaOrdemInvalida) {
			t.Fatalf("esperava ErrListaOrdemInvalida, obteve %v", err)
		}
	})

	t.Run("item repetido", func(t *testing.T) {
		_, err := svc.ReordenarItens(context.Background(), 1, 1, []int64{1, 2, 2})
		if !errors.Is(err, ErrListaOrdemInvalida) {
			t.Fatalf("esperava ErrListaOrdemInvalida, obteve %v", err)
		}
	})

	t.Run("item que nao pertence a lista", func(t *testing.T) {
		_, err := svc.ReordenarItens(context.Background(), 1, 1, []int64{1, 2, 999})
		if !errors.Is(err, ErrListaOrdemInvalida) {
			t.Fatalf("esperava ErrListaOrdemInvalida, obteve %v", err)
		}
	})
}

func TestListasService_SincronizarFranquia(t *testing.T) {
	t.Run("sincronizar em fila ou manual", func(t *testing.T) {
		repo := &mockListasRepo{
			buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
				return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "fila", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
		}
		svc := NewListasService(repo, &mockListasIGDB{})
		_, err := svc.SincronizarFranquia(context.Background(), 1, 1)
		if !errors.Is(err, ErrListaSincronizacaoNaoPermitida) {
			t.Fatalf("esperava ErrListaSincronizacaoNaoPermitida, obteve %v", err)
		}
	})

	t.Run("sincronizar adiciona apenas novos", func(t *testing.T) {
		regra := "franquia"
		val := "The Legend of Zelda"
		igdbID := int32(106)
		repo := &mockListasRepo{
			buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
				return &repository.Lista{ID: id, UsuarioID: usuarioID, Tipo: "desafio", RegraTipo: &regra, RegraValor: &val, RegraIgdbID: &igdbID, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
			},
			listarItensPorListaFn: func(ctx context.Context, listaID int64, usuarioID int32) ([]*repository.ListaItem, error) {
				ex1 := int32(101)
				return []*repository.ListaItem{
					{ID: 1, ListaID: listaID, IgdbID: &ex1, Nome: "Zelda 1", Posicao: 1},
				}, nil
			},
			sincronizarFranquiaFn: func(ctx context.Context, listaID int64, usuarioID int32, novosItens []repository.CriarItemParams) (int, error) {
				if len(novosItens) != 1 || *novosItens[0].IgdbID != 102 {
					t.Fatalf("esperava apenas Zelda 2 como novo, obteve %+v", novosItens)
				}
				return 1, nil
			},
		}
		igdb := &mockListasIGDB{
			atualizarJogosDaFranquiaFn: func(ctx context.Context, id int64) ([]igdbclient.Game, error) {
				return []igdbclient.Game{
					{ID: 101, Name: "Zelda 1"},
					{ID: 102, Name: "Zelda 2"},
				}, nil
			},
		}
		svc := NewListasService(repo, igdb)
		res, err := svc.SincronizarFranquia(context.Background(), 1, 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if res.Adicionados != 1 {
			t.Fatalf("esperava 1 adicionado, obteve %d", res.Adicionados)
		}
	})
}

func TestListasService_Progresso_Calculo(t *testing.T) {
	t.Run("meta nil usa total de itens e concluido_em na posicao meta", func(t *testing.T) {
		t1 := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		t2 := time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC)
		t3 := time.Date(2025, 3, 30, 0, 0, 0, 0, time.UTC)

		regra := "manual"
		meta := 2
		lista := &repository.Lista{
			ID:        1,
			UsuarioID: 1,
			Tipo:      "desafio",
			RegraTipo: &regra,
			Meta:      &meta,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		igdb1 := int32(11)
		igdb2 := int32(22)
		igdb3 := int32(33)
		itens := []*repository.ListaItem{
			{ID: 1, ListaID: 1, IgdbID: &igdb1, Nome: "Jogo 1", Posicao: 1},
			{ID: 2, ListaID: 1, IgdbID: &igdb2, Nome: "Jogo 2", Posicao: 2},
			{ID: 3, ListaID: 1, IgdbID: &igdb3, Nome: "Jogo 3", Posicao: 3},
		}

		jogos := []*repository.JogoZeradoResumo{
			{ID: 101, IgdbID: &igdb1, Nome: "Jogo 1", FinalizadoEm: t2, Nota: 10},
			{ID: 102, IgdbID: &igdb2, Nome: "Jogo 2", FinalizadoEm: t1, Nota: 9},
			{ID: 103, IgdbID: &igdb3, Nome: "Jogo 3", FinalizadoEm: t3, Nota: 8},
		}

		svc := NewListasService(&mockListasRepo{}, &mockListasIGDB{})
		resumo, _ := svc.calcularLista(lista, jogos, itens)

		if resumo.Progresso == nil {
			t.Fatal("esperava progresso preenchido")
		}
		if resumo.Progresso.Feitos != 3 {
			t.Fatalf("esperava feitos=3, obteve %d", resumo.Progresso.Feitos)
		}
		if resumo.Progresso.Meta != 2 {
			t.Fatalf("esperava meta=2, obteve %d", resumo.Progresso.Meta)
		}
		if resumo.Progresso.Percentual != 100 {
			t.Fatalf("esperava percentual limitado a 100, obteve %d", resumo.Progresso.Percentual)
		}
		if !resumo.Progresso.Concluido {
			t.Fatal("esperava concluido=true")
		}
		if resumo.Progresso.ConcluidoEm == nil || !resumo.Progresso.ConcluidoEm.Equal(t2) {
			t.Fatalf("esperava concluido_em na posicao 2 (t2: %v), obteve %v", t2, resumo.Progresso.ConcluidoEm)
		}
	})

	t.Run("nao concluido tem concluido_em nil", func(t *testing.T) {
		regra := "plataforma"
		val := "Genesis"
		meta := 3
		lista := &repository.Lista{
			ID:        2,
			UsuarioID: 1,
			Tipo:      "desafio",
			RegraTipo: &regra,
			RegraValor: &val,
			Meta:      &meta,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		jogos := []*repository.JogoZeradoResumo{
			{ID: 1, Console: "Genesis", Genero: "Acao", FinalizadoEm: time.Now()},
		}

		svc := NewListasService(&mockListasRepo{}, &mockListasIGDB{})
		resumo, _ := svc.calcularLista(lista, jogos, nil)

		if resumo.Progresso == nil {
			t.Fatal("esperava progresso")
		}
		if resumo.Progresso.Feitos != 1 || resumo.Progresso.Percentual != 33 || resumo.Progresso.Concluido || resumo.Progresso.ConcluidoEm != nil {
			t.Fatalf("progresso incorreto: %+v", resumo.Progresso)
		}
	})
}

func TestListasService_PgxErrNoRows(t *testing.T) {
	repo := &mockListasRepo{
		buscarPorIDFn: func(ctx context.Context, id int64, usuarioID int32) (*repository.Lista, error) {
			return nil, pgx.ErrNoRows
		},
		excluirFn: func(ctx context.Context, id int64, usuarioID int32) error {
			return pgx.ErrNoRows
		},
		excluirItemERecompactarFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error {
			return pgx.ErrNoRows
		},
	}
	svc := NewListasService(repo, &mockListasIGDB{})

	_, err := svc.ObterLista(context.Background(), 99, 1)
	if !errors.Is(err, ErrListaNaoEncontrada) {
		t.Fatalf("esperava ErrListaNaoEncontrada, obteve %v", err)
	}

	err = svc.ExcluirLista(context.Background(), 99, 1)
	if !errors.Is(err, ErrListaNaoEncontrada) {
		t.Fatalf("esperava ErrListaNaoEncontrada, obteve %v", err)
	}

	err = svc.ExcluirItem(context.Background(), 99, 1, 1)
	if !errors.Is(err, ErrListaItemNaoEncontrado) {
		t.Fatalf("esperava ErrListaItemNaoEncontrado, obteve %v", err)
	}
}
