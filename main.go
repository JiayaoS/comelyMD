package main

import (
	"embed"
	"fmt"
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

type databaseConfig struct {
	driver         string
	dataSourceName string
	localPath      string
	logTarget      string
}

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

func resolveDatabaseConfig(getenv func(string) string) (databaseConfig, error) {
	driver := strings.ToLower(strings.TrimSpace(getenv("DB_DRIVER")))
	if driver == "" {
		driver = "sqlite"
	}

	switch driver {
	case "sqlite":
		dataSourceName := strings.TrimSpace(getenv("DB_PATH"))
		if dataSourceName == "" {
			dataSourceName = "./data/app.db"
		}
		return databaseConfig{
			driver:         driver,
			dataSourceName: dataSourceName,
			localPath:      sqliteDatabasePath(dataSourceName),
			logTarget:      dataSourceName,
		}, nil
	case "libsql":
		databaseURL := strings.TrimSpace(getenv("DB_URL"))
		if databaseURL == "" {
			return databaseConfig{}, fmt.Errorf("DB_URL is required when DB_DRIVER=libsql")
		}

		dataSourceName := databaseURL
		authToken := strings.TrimSpace(getenv("DB_AUTH_TOKEN"))
		if authToken != "" {
			separator := "?"
			if strings.Contains(databaseURL, "?") {
				separator = "&"
			}
			dataSourceName = databaseURL + separator + "authToken=" + url.QueryEscape(authToken)
		}

		return databaseConfig{
			driver:         driver,
			dataSourceName: dataSourceName,
			logTarget:      databaseURL,
		}, nil
	default:
		return databaseConfig{}, fmt.Errorf("unsupported DB_DRIVER %q", driver)
	}
}

func main() {
	staticFiles, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("无法装载内嵌静态资源: %v", err)
	}
	handler.SetTemplates(embeddedFiles)
	handler.SetStatic(staticFiles)

	databaseConfig, err := resolveDatabaseConfig(os.Getenv)
	if err != nil {
		log.Fatalf("数据库配置无效: %v", err)
	}

	if databaseConfig.localPath != "" {
		if dbDir := filepath.Dir(databaseConfig.localPath); dbDir != "." {
			if err := os.MkdirAll(dbDir, 0o700); err != nil {
				log.Fatalf("无法预先建设持久化所需数据源数据安全保存文件夹: %v", err)
			}
		}
	}

	storage.InitDB(databaseConfig.driver, databaseConfig.dataSourceName)
	log.Printf("成功接入数据库进行基础载入，驱动=%s，相关目标定位：%s", databaseConfig.driver, databaseConfig.logTarget)

	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}

	addr := ":" + port
	handler.Run(addr)
}
