package mdblog

import "embed"

//go:embed web/templates/**/*.html
var TemplateFS embed.FS

//go:embed web/assets/*
var AssetFS embed.FS

//go:embed web/admin/dist/* web/admin/dist/assets/*
var AdminFS embed.FS
