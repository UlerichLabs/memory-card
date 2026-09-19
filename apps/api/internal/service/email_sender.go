package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type EmailSender interface {
	EnviarRecuperacaoSenha(ctx context.Context, destinatario, link string) error
}

type SMTPEmailSender struct {
	host      string
	port      int
	usuario   string
	senha     string
	remetente string
}

func NewSMTPEmailSender(host string, port int, usuario, senha, remetente string) *SMTPEmailSender {
	return &SMTPEmailSender{host: host, port: port, usuario: usuario, senha: senha, remetente: remetente}
}

func (sender *SMTPEmailSender) EnviarRecuperacaoSenha(ctx context.Context, destinatario, link string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("enviar email cancelado: %w", err)
	}
	client, err := smtp.DialStartTLS(fmt.Sprintf("%s:%d", sender.host, sender.port), &tls.Config{ServerName: sender.host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("conectar ao SMTP: %w", err)
	}
	defer client.Close()
	if err := client.Auth(sasl.NewPlainClient("", sender.usuario, sender.senha)); err != nil {
		return fmt.Errorf("autenticar no SMTP: %w", err)
	}
	message := []byte("To: " + destinatario + "\r\nFrom: " + sender.remetente + "\r\nSubject: Recuperacao de senha\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nUse este link para redefinir sua senha: " + link + "\r\n")
	if err := client.SendMail(sender.remetente, []string{destinatario}, strings.NewReader(string(message))); err != nil {
		return fmt.Errorf("enviar email SMTP: %w", err)
	}
	return nil
}

type LogEmailSender struct{}

func NewLogEmailSender() *LogEmailSender {
	return &LogEmailSender{}
}

func (sender *LogEmailSender) EnviarRecuperacaoSenha(ctx context.Context, destinatario, link string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("registrar email cancelado: %w", err)
	}
	slog.InfoContext(ctx, "email de recuperacao de senha", "destinatario", destinatario, "link", link)
	return nil
}
