package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrEmailJaCadastrado = errors.New("email ja cadastrado")
	ErrSenhaFraca        = errors.New("senha fraca")
	ErrEmailInvalido     = errors.New("email invalido")
)

type UsuarioRepository interface {
	ExistePorEmail(ctx context.Context, email string) (bool, error)
	Criar(ctx context.Context, nome, email, senhaHash string) (*repository.Usuario, error)
}

type CadastroService struct {
	repo UsuarioRepository
}

func NewCadastroService(repo UsuarioRepository) *CadastroService {
	return &CadastroService{repo: repo}
}

func (s *CadastroService) Cadastrar(ctx context.Context, nome, email, senha string) (*repository.Usuario, error) {
	nome = strings.TrimSpace(nome)
	email = strings.TrimSpace(email)

	if !validarEmail(email) {
		return nil, ErrEmailInvalido
	}

	if !validarSenha(senha) {
		return nil, ErrSenhaFraca
	}

	existe, err := s.repo.ExistePorEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("verificar email: %w", err)
	}
	if existe {
		return nil, ErrEmailJaCadastrado
	}

	senhaHash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("gerar hash da senha: %w", err)
	}

	usuario, err := s.repo.Criar(ctx, nome, email, string(senhaHash))
	if err != nil {
		return nil, fmt.Errorf("criar usuario no repositorio: %w", err)
	}

	return usuario, nil
}

func validarEmail(email string) bool {
	if email == "" {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	domain := parts[1]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	return true
}

func validarSenha(senha string) bool {
	if len([]rune(senha)) < 8 {
		return false
	}
	var hasUpper, hasDigit, hasSpecial bool
	for _, r := range senha {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r) || strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:',.<>?/`~", r):
			hasSpecial = true
		}
	}
	return hasUpper && hasDigit && hasSpecial
}
