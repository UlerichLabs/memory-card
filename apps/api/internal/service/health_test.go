package service

import (
	"context"
	"errors"
	"testing"
)

type mockDatabasePinger struct {
	pingFn func(ctx context.Context) error
}

func (m *mockDatabasePinger) Ping(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return nil
}

func TestHealthService_CheckDatabase(t *testing.T) {
	errBanco := errors.New("falha de conexao com postgres")

	tests := []struct {
		name        string
		pingErr     error
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "banco saudavel retorna sucesso",
			pingErr:     nil,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "banco indisponivel retorna ErrDatabaseUnavailable empacotado",
			pingErr:     errBanco,
			wantErr:     true,
			expectedErr: ErrDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pinger := &mockDatabasePinger{
				pingFn: func(ctx context.Context) error {
					return tc.pingErr
				},
			}
			svc := NewHealthService(pinger)

			err := svc.CheckDatabase(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("CheckDatabase() erro = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
				t.Fatalf("CheckDatabase() erro = %v, esperava erro empacotando %v", err, tc.expectedErr)
			}
		})
	}
}
