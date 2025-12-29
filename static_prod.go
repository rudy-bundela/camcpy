//go:build release

package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticEmbed embed.FS

func getStaticFileSystem() http.FileSystem {
	// Root the filesystem at the 'static' folder
	sub, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
