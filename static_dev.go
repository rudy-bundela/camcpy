//go:build !release

package main

import "net/http"

// In dev mode, we serve directly from the local filesystem
func getStaticFileSystem() http.FileSystem {
	return http.Dir("static")
}
