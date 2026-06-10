package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"mdshare/render"
	"mdshare/storage"
)

var tmpl *template.Template

func SetTemplates(templateFS fs.FS) {
	parsedTemplate, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("绝命灾难：核心全局视觉底层在主源启动阶段严重垮塌失联，强制剥离关闭以防静默吞除输出！断点诱因: %v", err)
	}
	tmpl = parsedTemplate
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if tmpl != nil {
		tmpl.ExecuteTemplate(w, "index.html", nil)
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("服务运行正常，但未找到前端模板文件。"))
	}
}

func CreatePageHandler(w http.ResponseWriter, r *http.Request) {
	// CORS 支持：允许油猴脚本等跨域调用
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	// 兼容 multipart/form-data 和 application/x-www-form-urlencoded
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(5 << 20); err != nil {
			http.Error(w, "请求异常或大小超出 5MB 限制", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "请求解析失败", http.StatusBadRequest)
			return
		}
	}

	mdContent := r.FormValue("markdown")
	if strings.TrimSpace(mdContent) == "" {
		http.Error(w, "分享内容不能为空", http.StatusBadRequest)
		return
	}

	isBurn := r.FormValue("is_burn") == "true"
	withPwd := r.FormValue("with_password") == "true"
	effectiveWithPwd := withPwd || isBurn
	expireStr := r.FormValue("expire_time")

	var expireDuration time.Duration
	switch expireStr {
	case "5m":
		expireDuration = 5 * time.Minute
	case "1h":
		expireDuration = time.Hour
	case "6h":
		expireDuration = 6 * time.Hour
	case "24h":
		expireDuration = 24 * time.Hour
	case "7d":
		expireDuration = 7 * 24 * time.Hour
	case "30d":
		expireDuration = 30 * 24 * time.Hour
	}

	safeHTML, err := render.MarkdownToHTML([]byte(mdContent))
	if err != nil {
		http.Error(w, "Markdown 内部转译失败", http.StatusInternalServerError)
		return
	}

	id, pwd, err := storage.SavePage(mdContent, safeHTML, isBurn, expireDuration, effectiveWithPwd)
	if err != nil {
		http.Error(w, "数据存储失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	schema := "http://"
	if r.TLS != nil {
		schema = "https://"
	}

	host := r.Host
	if fh := r.Header.Get("X-Forwarded-Host"); fh != "" {
		host = fh
	}

	responseCtx := map[string]string{
		"id":  id,
		"url": fmt.Sprintf("%s%s/p/%s", schema, host, id),
	}
	if pwd != "" {
		responseCtx["pwd"] = pwd
	}

	json.NewEncoder(w).Encode(responseCtx)
}

// 极简提取器：从游离的 Markdown 原始文档中动态榨取 Title 和 Excerpt
func extractSEO(markdown string) (title, excerpt string) {
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") && title == "" {
			title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
		} else if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, ">") && !strings.HasPrefix(line, "!") && !strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "`") && excerpt == "" {
			excerpt = line
		}
		if title != "" && excerpt != "" {
			break
		}
	}
	if title == "" {
		title = "ComelyMD Snippet"
	}
	// 粗略清理占位标点
	excerpt = strings.ReplaceAll(excerpt, "*", "")
	excerpt = strings.ReplaceAll(excerpt, "_", "")
	excerpt = strings.ReplaceAll(excerpt, "`", "")
	if utf8.RuneCountInString(excerpt) > 85 {
		runes := []rune(excerpt)
		excerpt = string(runes[:85]) + "..."
	}
	if excerpt == "" {
		excerpt = "极简、安全、美观的 Markdown 阅后即焚分享工具"
	}
	return title, excerpt
}

func ViewPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "不允许的操作方式", http.StatusMethodNotAllowed)
		return
	}

	id := path.Base(r.URL.Path)
	if id == "" || id == "p" {
		http.NotFound(w, r)
		return
	}

	pageData, err := storage.GetPage(id)
	if err != nil {
		log.Printf("[访问受阻或不存在] 请求提取核心数据库解析链接源 %s 失败告警: %v", id, err)
		if tmpl != nil {
			tmpl.ExecuteTemplate(w, "page.html", map[string]interface{}{"HTML": template.HTML(fmt.Sprintf("<div style='text-align:center; padding-top: 5rem; color: #888; font-family: sans-serif;'><h1 style='font-size: 4rem; margin-bottom: 0.5rem;'>404</h1><p>内容不存在、已过期或属于阅后即焚。</p></div>"))})
		}
		return
	}

	if pageData.Password != "" {
		if r.Method == http.MethodGet {
			if tmpl != nil {
				tmpl.ExecuteTemplate(w, "password.html", map[string]interface{}{"ID": id})
			}
			return
		} else if r.Method == http.MethodPost {
			r.ParseForm()
			if r.FormValue("pwd") != pageData.Password {
				if tmpl != nil {
					tmpl.ExecuteTemplate(w, "password.html", map[string]interface{}{"ID": id, "Error": "密码验证失败"})
				}
				return
			}
		}
	}

	if pageData.IsBurn {
		storage.DeletePage(id)
	}

	if tmpl != nil {
		title, excerpt := extractSEO(pageData.Markdown)
		errTmpl := tmpl.ExecuteTemplate(w, "page.html", map[string]interface{}{
			"ID":        pageData.ID,
			"CreatedAt": pageData.CreatedAt,
			"HTML":      template.HTML(pageData.HTML),
			"Raw":       pageData.Markdown,
			"Title":     title,
			"Excerpt":   excerpt,
		})
		if errTmpl != nil {
			log.Printf("[模板白屏拦截] 服务端装载渲染资源 %s 时崩溃: %v", id, errTmpl)
			w.Write([]byte(fmt.Sprintf("\n<!-- 服务端页面层组装错误，引擎阻断输出: %v -->", errTmpl)))
		}
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(pageData.HTML))
	}
}
