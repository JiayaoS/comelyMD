package handler

import (
	"io/fs"
	"log"
	"net/http"
	"time"
)

var staticFS fs.FS

func SetStatic(staticAssets fs.FS) {
	staticFS = staticAssets
}

// Run 挂载路由端口配置服务并启动阻塞监听运行。
func Run(addr string) {
	mux := http.NewServeMux()

	// 开放接口与页面展现资源绑定
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/api/pages", CreatePageHandler)
	mux.HandleFunc("/p/", ViewPageHandler)

	if staticFS != nil {
		fs := http.FileServer(http.FS(staticFS))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Markdown 服务成功启动，正监听终端点: %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务因严重异常停止或关闭: %v", err)
	}
}
