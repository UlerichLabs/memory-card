// Package migrations expõe os arquivos SQL de migração embutidos no binário.
package migrations

import "embed"

// FS contém todos os arquivos de migração SQL da aplicação.
//
//go:embed *.sql
var FS embed.FS
