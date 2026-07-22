package web

import "embed"

// Templates holds HTML templates under templates/.
//
//go:embed all:templates
var Templates embed.FS

// Static holds static assets under static/.
//
//go:embed all:static
var Static embed.FS
