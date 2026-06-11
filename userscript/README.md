# 🔌 ComelyMD Share 油猴脚本

将 AI 聊天平台的回复内容一键分享到自建 ComelyMD 服务。

## ✨ 支持平台

| 平台 | 状态 |
|------|------|
| ChatGPT | ✅ |
| Gemini | ✅ |
| Grok | 🔲 计划中 |
| 豆包 | 🔲 计划中 |
| Kimi | 🔲 计划中 |

## 📦 安装

### 前置条件
安装 [Tampermonkey](https://www.tampermonkey.net/) 浏览器扩展（Chrome / Edge / Firefox 均支持）

### 安装脚本
1. 打开 Tampermonkey 管理面板
2. 点击「新建脚本」
3. 将 `comelymd-share.user.js` 的完整内容粘贴进去
4. 保存（Ctrl + S）

## ⚙️ 配置

首次点击分享按钮时会自动弹出配置面板，填写你的 ComelyMD 服务器地址即可。

后续修改配置：点击浏览器中 Tampermonkey 图标 → `ComelyMD 设置`

## 🎯 使用

1. 在 ChatGPT 或 Gemini 中正常对话
2. AI 回复完成后，在消息操作栏会出现一个 ↑ 分享按钮
3. 点击后选择分享选项（阅后即焚 / 密码 / 过期时间）
   - 启用阅后即焚时会自动开启密码保护，并在该状态下锁定密码选项
4. 点击「分享」，链接自动复制到剪贴板

## 🛠️ 技术说明

- **Markdown 还原**：使用 [Turndown.js](https://github.com/mixmark-io/turndown) 从渲染后的 HTML 逆向还原
- **跨域请求**：使用油猴原生 `GM_xmlhttpRequest`，无需服务端额外配置 CORS
- **配置存储**：使用 `GM_setValue / GM_getValue` 持久化到浏览器本地
