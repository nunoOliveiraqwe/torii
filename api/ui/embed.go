package ui

import "embed"

//go:embed templates
var Assets embed.FS

//go:embed static
var Static embed.FS
