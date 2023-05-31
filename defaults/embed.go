package defaults

import "embed"

//go:embed compiler/generated/compiler.js
var CompilerFS embed.FS

//go:embed all:core/*
var CoreFS embed.FS

//go:embed all:node_modules/*
var NodeModulesFS embed.FS

//go:embed all:starters/bare/*
var BareFS embed.FS

//go:embed all:starters/learner/*
var LearnerFS embed.FS
