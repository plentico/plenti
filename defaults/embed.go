package defaults

import "embed"

//go:embed all:ejected/*
var EjectedFS embed.FS

//go:embed all:node_modules/*
var NodeModulesFS embed.FS

//go:embed all:starters/bare/*
var BareFS embed.FS

//go:embed all:starters/learner/*
var LearnerFS embed.FS
