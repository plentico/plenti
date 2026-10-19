package defaults

import "embed"

//go:embed all:core/*
var CoreFS embed.FS

//go:embed all:starter/*
var StarterFS embed.FS
