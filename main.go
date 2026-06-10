package main

import (
	"embed"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"mdshare/handler"
	"mdshare/storage"
)

//go:embed templates/*.html static/*
var embeddedFiles embed.FS

func sqliteDatabasePath(dataSourceName string) string {
	if dataSourceName == ":memory:" {
		return ""
	}

	if !strings.HasPrefix(dataSourceName, "file:") {
		return dataSourceName
	}

	parsed, err := url.Parse(dataSourceName)
	if err != nil {
		return ""
	}

	databasePath := parsed.Path
	if parsed.Opaque != "" {
		databasePath = parsed.Opaque
	}
	if strings.HasPrefix(databasePath, ":memory:") {
		return ""
	}
	return databasePath
}

func main() {
	staticFiles, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("无法装载内嵌静态资源: %v", err)
	}
	handler.SetTemplates(embeddedFiles)
	handler.SetStatic(staticFiles)

	// 加载持久化保护，以环境变量声明位置优先加载否则向容下写入映射
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}

	databasePath := sqliteDatabasePath(dbPath)
	if databasePath != "" {
		if dbDir := filepath.Dir(databasePath); dbDir != "." {
			if err := os.MkdirAll(dbDir, 0o700); err != nil {
				log.Fatalf("无法预先建设持久化所需数据源数据安全保存文件夹: %v", err)
			}
		}
	}

	storage.InitDB(dbPath)
	log.Printf("成功接入数据库进行基础载入，相关目标文件定位： %s", dbPath)

	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}

	addr := ":" + port
	handler.Run(addr)
}
