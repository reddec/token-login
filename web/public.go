package web

import (
	"embed"
	"io/fs"
	"log/slog"
)

//go:embed admin-ui/dist
var dist embed.FS

func Assets() fs.FS {
	sub, err := fs.Sub(dist, "admin-ui/dist")
	if err != nil {
		slog.Error("failed load UI - did you compile it?", "error", err)
		return dist
	}
	return sub
}
