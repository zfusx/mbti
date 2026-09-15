package web

import "embed"

// Static contains the built frontend assets served by the API.
//
//go:embed static/*
var Static embed.FS
