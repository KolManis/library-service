// Package migrations вшивает SQL-миграции в бинарник (go:embed),
// чтобы приложению не нужна была папка с файлами рядом с собой.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
