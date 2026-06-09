# 🎨 ComelyMD

![Version](https://img.shields.io/badge/version-v1.4.1-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

极简、安全、美观的 Markdown 内容分享平台。基于 Go + SQLite，专为低配服务器设计。

---

## ✨ 功能特性

- 🔗 **极短链接** — Base62 随机 ID，独立访问，无法遍历
- 🔥 **阅后即焚** — 单次阅读后自动销毁
- 🔒 **密码保护** — 自动生成 4 位随机密码
- ⏱️ **自动过期** — 支持 5 分钟 ~ 30 天灵活时限
- 🌗 **暗/亮主题** — 跟随系统偏好，支持手动切换
- 📋 **一键操作** — 复制链接、复制链接+密码、点击复制密码
- 📄 **阅读增强** — 侧边栏快捷：复制内容 / 下载 .md / 导出 PDF
- 🖥️ **工作流级双栏体验** — 搭载前端极速解析沙盒引擎，沉浸式实时分栏预览，打字效果无延迟跟随
- 🧮 **顶级学术/工程支持** — 核心拦截器原生重构，对数学公式 `KaTeX` / `Mermaid` 引擎双闭环解析，完全不破坏原生下划线等渲染闭环
- 🎯 **社交卡片分享提全** — 轻量正则去标签化洗稿提纯，自动自内容体摘出 `OG:Title` 等 SEO 级别高转链开屏图文卡片
- 📱 **响应式排版** — PC 端沉浸式等宽居中，移动端防溢出滑动保护
- 📑 **智能导览树** — 桌面级悬浮锚点目录 (TOC)，移动端支持顶层抽屉交互
- 🧑‍💻 **代码块增强** — 智能编程语言探测标题栏，附带一键入板复制组件
- 🎨 **现代 UI** — 自定义设计系统，品牌渐变色，无第三方 CSS 框架 
- 🪶 **极致轻量** — 纯 Go 编译，Docker 构建依托双缓存机制（镜像 < 20MB）

## 🖥️ 页面预览

| 首页              | 阅读页                   | 密码页               |
| ----------------- | ------------------------ | -------------------- |
| 编辑器 + 配置面板 | Markdown 渲染 + 侧边工具 | 品牌渐变锁图标验证卡 |

## 🚀 部署

### Docker Compose（推荐，持久化生产部署）

```bash
git clone https://github.com/Loxonl/comelyMD.git
cd comelyMD

# 拉取预构建镜像并启动
docker-compose pull
docker-compose up -d
```

> 💡 **更新版本**：`docker-compose pull && docker-compose up -d`

### Vercel 一键部署（演示/预览）

[![Deploy to Vercel](https://vercel.com/button)](https://vercel.com/import/project?template=https://github.com/Loxonl/comelyMD)

本仓库已包含 `vercel.json`，Vercel 会使用 Go 运行时的 Go Framework Preset 从根目录 `main.go` 构建并运行服务。默认配置：

```text
DB_PATH=/tmp/comelymd/comelymd.db
```

> ⚠️ Vercel Functions 只有只读部署文件系统和 `/tmp` 临时写入空间。当前 SQLite 文件在 Vercel 上不具备可靠持久化，数据可能在冷启动、重新部署、实例切换或平台清理后丢失。请保持 Vercel 的 `DB_PATH` 指向 `/tmp`，不要改成部署目录内路径。因此 Vercel 部署适合演示、预览和临时分享；需要长期保存分享内容时，请使用 Docker Compose / VPS 挂载卷部署，或后续接入外部持久化数据库。

<details>
<summary><b>Vercel 部署步骤</b></summary>

1. 点击上方 `Deploy to Vercel` 按钮。
2. 使用 GitHub 登录 Vercel。
3. 按提示导入或 fork 本仓库。
4. 保持默认项目配置；仓库内的 `vercel.json` 已声明 Go 项目配置和 `DB_PATH`。
5. 点击 Deploy，部署完成后访问 Vercel 分配的域名。

</details>

### 本地开发

```bash
# 本地构建并启动（含热更新支持）
docker-compose -f docker-compose.dev.yml up -d --build

# 访问 http://localhost:18080
```

## 📡 API

```
POST /api/pages
Content-Type: multipart/form-data
```

| 参数              | 必填 | 说明                                        |
| ----------------- | ---- | ------------------------------------------- |
| `markdown`      | ✅   | Markdown 内容                               |
| `is_burn`       | —   | `"true"` 开启阅后即焚                     |
| `with_password` | —   | `"true"` 生成随机密码                     |
| `expire_time`   | —   | `5m` `1h` `6h` `24h` `7d` `30d` |

**响应示例：**

```json
{
  "id": "aBcDeFgH",
  "url": "https://your-domain/p/aBcDeFgH",
  "pwd": "x9k2"
}
```

## 🛠️ 技术栈

| 组件     | 技术                                |
| -------- | ----------------------------------- |
| 后端     | Go 1.21 · net/http                 |
| 数据库   | SQLite（modernc.org/sqlite，纯 Go） |
| Markdown | Goldmark + Bluemonday               |
| 代码高亮 | Highlight.js                        |
| 字体     | Inter + JetBrains Mono              |
| 图标     | Font Awesome 6                      |
| 部署     | Docker · Vercel · GitHub Actions · GHCR |

## 📋 Roadmap

- [ ] Markdown 实时预览
- [ ] 多文件/标签页聚合分享
- [ ] 自定义短链接别名
- [ ] 访问统计（可选开启）


## 🔌 油猴脚本 — AI 对话一键分享

在 ChatGPT / Gemini 等 AI 聊天页面中，将 AI 回复内容一键分享到自建 ComelyMD 服务。

支持阅后即焚、密码保护、过期时间等完整选项。

👉 **安装与使用指南**：[userscript/README.md](./userscript/README.md)

## 🏞️ 预览

![1](image/README/1.png)

![2](image/README/2.png)

![3](image/README/3.png)

## 🙏 致谢 (Acknowledgements)

本项目在开发中引用/借鉴了以下优秀的开源项目，特此鸣谢（所有引用均依据其各自的开源协议合法使用）：

- [Goldmark](https://github.com/yuin/goldmark) (MIT): 极速且高度可扩展的 Markdown 解析器，用作本项目的核心渲染引擎。
- [Bluemonday](https://github.com/microcosm-cc/bluemonday) (BSD-3-Clause): 强大的 HTML 净化器，用于彻底防御 XSS 攻击。
- [modernc.org/sqlite](https://modernc.org/sqlite) (Zlib): CGO-Free 的 SQLite 数据库引擎。
- [Highlight.js](https://highlightjs.org/) (BSD-3-Clause): 极简优美的语法高亮显示工具。
- [KaTeX](https://katex.org/) (MIT): 极速的数学公式网页端排版支持库。
- [Mermaid](https://mermaid.js.org/) (MIT): 使用类 Markdown 文本生成图表的出众库。
- [Turndown](https://github.com/mixmark-io/turndown) (MIT): 用于油猴脚本，将 HTML 页面结构高效地逆向转换为纯净 Markdown。
- [Font Awesome](https://fontawesome.com/) (SIL OFL 1.1 / MIT): 项目界面中使用的精美矢量图标准。

## 📄 协议

[MIT License](./LICENSE)
