package usersdocfirst

import "embed"

// SpecFS stores the doc-first OpenAPI spec and referenced component files.
//
//go:embed swagger.yaml components/*.yaml
var SpecFS embed.FS
