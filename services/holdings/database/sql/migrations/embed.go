package migrations

import "embed"

// Embed all SQL migration files.
//
//go:embed *.sql
var Embed embed.FS



