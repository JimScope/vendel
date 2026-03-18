package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distDir embed.FS

// DistDirFS is the frontend build output, ready to serve as a static file system.
var DistDirFS, _ = fs.Sub(distDir, "dist")
