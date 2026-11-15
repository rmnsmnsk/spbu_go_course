package migrations

import "embed"

// Files contains SQL migrations compiled into the service binary.
//
//go:embed *.sql
var Files embed.FS
