package handler

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
)

//go:embed static/*
var distFiles embed.FS

// GetStaticHandler 返回处理静态文件的服务
func GetStaticHandler() http.Handler {
	fsys, err := fs.Sub(distFiles, "static")
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		} else if path[0] == '/' {
			path = path[1:]
		}

		// 检查文件是否存在
		f, err := fsys.Open(path)
		if err != nil {
			path = "index.html"
			f, err = fsys.Open(path)
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		}
		f.Close()

		// 设置正确的 Content-Type
		ext := filepath.Ext(path)
		switch ext {
		case ".html":
			w.Header().Set("Content-Type", "text/html")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".js", ".mjs":
			w.Header().Set("Content-Type", "application/javascript")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		default:
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		// 使用 http.FileServer 提供文件
		http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
	})
}
