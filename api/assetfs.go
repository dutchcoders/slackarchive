package api

import (
	"net/http"
	"os"

	assets "github.com/dutchcoders/slackarchive-app"
	assetfs "github.com/elazarl/go-bindata-assetfs"
)

func AssetFS() *assetFS {
	return &assetFS{
		FileSystem: &assetfs.AssetFS{
			AssetInfo: func(path string) (os.FileInfo, error) {
				return os.Stat(path)
			},
			Asset:    assets.Asset,
			AssetDir: assets.AssetDir,
			Prefix:   assets.Prefix,
		},
		DefaultDoc: "index.html",
	}
}

type assetFS struct {
	http.FileSystem

	DefaultDoc string
}

func (fs assetFS) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err == nil {
		return f, err
	} else if !os.IsNotExist(err) {
		return f, err
	}

	return fs.FileSystem.Open(fs.DefaultDoc)
}
