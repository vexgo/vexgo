package theme

import "embed"

//go:embed themes/*
var DefaultThemeFS embed.FS
