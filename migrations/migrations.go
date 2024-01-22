package migrations

import "embed"

// Migrations -
// Issue - https://github.com/golang/go/issues/46056
//
//go:embed *.sql
var Migrations embed.FS
