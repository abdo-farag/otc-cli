package web

import "embed"

//go:embed templates/*
var Content embed.FS
