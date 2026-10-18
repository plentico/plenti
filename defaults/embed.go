package defaults

import "embed"

//go:embed all:core/*
var CoreFS embed.FS

//go:embed all:starters/default/*
var StarterFS embed.FS
