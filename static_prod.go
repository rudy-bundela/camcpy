//go:build release

package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static
var staticEmbed embed.FS

func getStaticFileSystem() http.FileSystem {
	sub, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		log.Fatalf("Failed to create sub-filesystem: %v", err)
	}
	return http.FS(sub)
}
