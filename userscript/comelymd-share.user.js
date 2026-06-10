// ==UserScript==
// @name         ComelyMD Share — AI 对话分享
// @namespace    https://github.com/Loxonl/comelyMD
// @version      1.4.8
// @description  将 ChatGPT / Gemini 的 AI 回复一键分享到自建 ComelyMD 服务
// @author       Loxonl
// @match        https://chatgpt.com/*
// @match        https://chat.openai.com/*
// @match        https://gemini.google.com/*
// @grant        GM_xmlhttpRequest
// @grant        GM_setValue
// @grant        GM_getValue
// @grant        GM_registerMenuCommand
// @connect      *
// @require      https://unpkg.com/turndown@7.1.3/dist/turndown.js
// @require      https://unpkg.com/turndown-plugin-gfm@1.0.2/dist/turndown-plugin-gfm.js
// @run-at       document-idle
// ==/UserScript==

(function () {
  'use strict';

  // ─── 常量 ───
  const SCRIPT_ID = 'comelymd-share';
  const BRAND_GRADIENT = 'linear-gradient(135deg, #3b82f6, #6366f1, #a855f7)';
  const PROCESSED_ATTR = `data-${SCRIPT_ID}`;

  // ─── DOM 工厂函数（纯 DOM API，完全绕过 Trusted Types CSP）───
  function h(tag, attrs, ...children) {
    const el = document.createElement(tag);
    if (attrs) {
      Object.entries(attrs).forEach(([k, v]) => {
        if (v == null) return;
        if (k === 'style' && typeof v === 'object') Object.assign(el.style, v);
        else if (k === 'className') el.className = v;
        else if (k.startsWith('on') && typeof v === 'function') el.addEventListener(k.slice(2), v);
        else el.setAttribute(k, v);
      });
    }
    children.flat().forEach(c => {
      if (c == null) return;
      el.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
    });
    return el;
  }

  function opt(value, label, selected) {
    const o = document.createElement('option');
    o.value = value; o.textContent = label;
    if (selected) o.selected = true;
    return o;
  }

  // ─── Turndown 实例 ───
  const td = new TurndownService({ headingStyle: 'atx', codeBlockStyle: 'fenced' });
  if (typeof turndownPluginGfm !== 'undefined') {
    td.use(turndownPluginGfm.gfm);
  }

  // ─── 配置管理 ───
  function getConfig() {
    return {
      server: GM_getValue('server', ''),
      defaultBurn: GM_getValue('defaultBurn', false),
      defaultPwd: GM_getValue('defaultPwd', false),
      defaultExpire: GM_getValue('defaultExpire', '0'),
    };
  }
  function setConfig(cfg) {
    GM_setValue('server', cfg.server);
    GM_setValue('defaultBurn', cfg.defaultBurn);
    GM_setValue('defaultPwd', cfg.defaultPwd);
    GM_setValue('defaultExpire', cfg.defaultExpire);
  }

  // ─── 全局样式注入 ───
  function injectStyles() {
    if (document.getElementById(`${SCRIPT_ID}-styles`)) return;
    const style = document.createElement('style');
    style.id = `${SCRIPT_ID}-styles`;
    style.textContent = `
      .cmd-share-btn {
        display: inline-flex; align-items: center; justify-content: center;
        width: 28px; height: 28px; border-radius: 6px; cursor: pointer;
        background: transparent; border: none; color: inherit; padding: 0;
        transition: transform 0.2s; position: relative;
      }
      .cmd-share-btn:hover { transform: scale(1.05); }
      .cmd-share-btn svg { width: 18px; height: 18px; }
      .cmd-overlay {
        position: fixed; inset: 0; background: rgba(0,0,0,0.5);
        z-index: 99998; display: flex; align-items: center; justify-content: center;
        animation: cmd-fade-in 0.2s ease;
      }
      @keyframes cmd-fade-in { from { opacity: 0; } to { opacity: 1; } }
      .cmd-panel {
        background: #1a1c2e; color: #e4e5ea; border-radius: 16px;
        width: 380px; max-width: 90vw; box-shadow: 0 24px 64px rgba(0,0,0,0.4);
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        font-size: 14px; overflow: hidden;
        animation: cmd-slide-up 0.25s ease;
      }
      @keyframes cmd-slide-up { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: translateY(0); } }
      .cmd-panel-header {
        padding: 1.25rem 1.5rem; font-weight: 700; font-size: 15px;
        background: ${BRAND_GRADIENT}; color: #fff;
      }
      .cmd-panel-body { padding: 1.25rem 1.5rem; }
      .cmd-field { margin-bottom: 1rem; }
      .cmd-field label { display: block; font-size: 12px; color: #8b8fa3; margin-bottom: 0.4rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; }
      .cmd-field input[type="text"] {
        width: 100%; padding: 0.6rem 0.8rem; background: #0f1117; color: #e4e5ea;
        border: 1px solid #2a2d3a; border-radius: 8px; font-size: 13px;
        outline: none; transition: border-color 0.2s; box-sizing: border-box;
      }
      .cmd-field input[type="text"]:focus { border-color: #6366f1; }
      .cmd-opts { display: flex; flex-wrap: wrap; gap: 0.75rem; margin-bottom: 1rem; }
      .cmd-opt {
        display: flex; align-items: center; gap: 0.4rem; font-size: 13px;
        cursor: pointer; user-select: none; font-weight: 500;
      }
      .cmd-opt input[type="checkbox"] {
        appearance: none; -webkit-appearance: none; width: 16px; height: 16px;
        border: 2px solid #2a2d3a; border-radius: 4px; cursor: pointer;
        transition: all 0.15s; position: relative; background: #0f1117; flex-shrink: 0;
      }
      .cmd-opt input[type="checkbox"]:checked { background: #6366f1; border-color: #6366f1; }
      .cmd-opt input[type="checkbox"]:checked::after {
        content: '\\2713'; position: absolute; top: -1px; left: 2px; font-size: 12px; color: #fff; font-weight: 700;
      }
      .cmd-select { display: flex; align-items: center; gap: 0.4rem; font-size: 13px; font-weight: 500; }
      .cmd-select select {
        background: #0f1117; color: #e4e5ea; border: 1px solid #2a2d3a;
        border-radius: 6px; padding: 0.3rem 0.5rem; font-size: 12px; outline: none; cursor: pointer;
      }
      .cmd-panel-footer { display: flex; gap: 0.75rem; padding: 0 1.5rem 1.25rem; }
      .cmd-btn {
        flex: 1; padding: 0.65rem; border: none; border-radius: 8px;
        font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s;
        font-family: inherit;
      }
      .cmd-btn-primary { background: ${BRAND_GRADIENT}; color: #fff; box-shadow: 0 2px 8px rgba(99,102,241,0.3); }
      .cmd-btn-primary:hover { opacity: 0.9; transform: translateY(-1px); }
      .cmd-btn-primary:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }
      .cmd-btn-cancel { background: #2a2d3a; color: #8b8fa3; }
      .cmd-btn-cancel:hover { color: #e4e5ea; }
      .cmd-result { text-align: center; padding: 0.5rem 0; }
      .cmd-result .icon { font-size: 2rem; margin-bottom: 0.75rem; }
      .cmd-result-url {
        width: 100%; padding: 0.55rem 0.7rem; background: #0f1117; color: #e4e5ea;
        border: 1px solid #2a2d3a; border-radius: 6px; font-size: 12px;
        font-family: monospace; margin: 0.75rem 0; text-align: center;
        outline: none; box-sizing: border-box;
      }
      .cmd-result-pwd {
        display: inline-block; font-family: monospace; font-size: 1.1rem; letter-spacing: 3px;
        background: rgba(99,102,241,0.15); padding: 0.3rem 0.8rem; border-radius: 6px;
        color: #818cf8; margin: 0.5rem 0; font-weight: 700;
      }
      .cmd-toast {
        position: fixed; left: 50%; bottom: 40px; transform: translateX(-50%);
        background: #1a1c2e; color: #e4e5ea; padding: 0.6rem 1.2rem;
        border-radius: 8px; font-size: 13px; font-weight: 500; z-index: 99999;
        box-shadow: 0 8px 24px rgba(0,0,0,0.4); animation: cmd-fade-in 0.2s ease;
      }
    `;
    document.head.appendChild(style);
  }

  // ─── 工具函数 ───
  function toast(msg, dur = 2500) {
    const el = h('div', {className: 'cmd-toast'}, msg);
    document.body.appendChild(el);
    setTimeout(() => el.remove(), dur);
  }

  async function clipText(text) {
    try { await navigator.clipboard.writeText(text); } catch {
      const ta = document.createElement('textarea');
      ta.value = text; document.body.appendChild(ta); ta.select();
      document.execCommand('copy'); document.body.removeChild(ta);
    }
  }

  // ─── 分享按钮（纯 DOM API 构建 SVG）───
  function createShareIcon() {
    const btn = document.createElement('button');
    btn.className = 'cmd-share-btn';
    btn.title = '分享到 ComelyMD';
    const ns = 'http://www.w3.org/2000/svg';
    const svg = document.createElementNS(ns, 'svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '2');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('stroke-linejoin', 'round');
    const path = document.createElementNS(ns, 'path');
    path.setAttribute('d', 'M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8');
    const polyline = document.createElementNS(ns, 'polyline');
    polyline.setAttribute('points', '16 6 12 2 8 6');
    const line = document.createElementNS(ns, 'line');
    line.setAttribute('x1', '12'); line.setAttribute('y1', '2');
    line.setAttribute('x2', '12'); line.setAttribute('y2', '15');
    svg.append(path, polyline, line);
    
    // 用户需求：添加 MD 角标以区分官方分享按钮，黑底白字高亮
    const badge = document.createElement('span');
    badge.textContent = 'MD';
    Object.assign(badge.style, {
      position: 'absolute', right: '-6px', bottom: '-4px', fontSize: '10px', 
      fontWeight: '800', background: '#000', 
      color: 'white', borderRadius: '4px', padding: '1px 3px', transform: 'scale(0.8)',
      boxShadow: '0 1px 3px rgba(0,0,0,0.3)', pointerEvents: 'none'
    });
    
    btn.append(svg, badge);
    return btn;
  }

  // ─── 面板 UI（全部纯 DOM API，零 innerHTML）───
  function showPanel(markdownContent) {
    const cfg = getConfig();
    if (!cfg.server) { showSettingsPanel(() => showPanel(markdownContent)); return; }

    const serverInput = h('input', {type:'text', value:cfg.server, placeholder:'https://md.example.com'});
    const burnCheck = h('input', {type:'checkbox'});
    if (cfg.defaultBurn) burnCheck.checked = true;
    const pwdCheck = h('input', {type:'checkbox'});
    if (cfg.defaultPwd) pwdCheck.checked = true;
    const expireSelect = h('select', {},
      opt('0', '永不', cfg.defaultExpire === '0'),
      opt('5m', '5 分钟', cfg.defaultExpire === '5m'),
      opt('1h', '1 小时', cfg.defaultExpire === '1h'),
      opt('6h', '6 小时', cfg.defaultExpire === '6h'),
      opt('24h', '24 小时', cfg.defaultExpire === '24h'),
      opt('7d', '7 天', cfg.defaultExpire === '7d'),
      opt('30d', '30 天', cfg.defaultExpire === '30d'),
    );
    const cancelBtn = h('button', {className:'cmd-btn cmd-btn-cancel'}, '取消');
    const submitBtn = h('button', {className:'cmd-btn cmd-btn-primary'}, '分享');
    const passwordHint = h('div', {style:{fontSize:'12px', color:'#8b8fa3', display:'none'}}, '阅后即焚启用时会自动开启密码保护。');

    function syncProtectionOptions() {
      const isBurnEnabled = burnCheck.checked;
      if (isBurnEnabled) pwdCheck.checked = true;
      pwdCheck.disabled = isBurnEnabled;
      passwordHint.style.display = isBurnEnabled ? 'block' : 'none';
    }

    const bodyDiv = h('div', {className:'cmd-panel-body'},
      h('div', {className:'cmd-field'}, h('label', {}, '服务器'), serverInput),
      h('div', {className:'cmd-opts'},
        h('label', {className:'cmd-opt'}, burnCheck, ' 阅后即焚'),
        h('label', {className:'cmd-opt'}, pwdCheck, ' 密码保护'),
      ),
      passwordHint,
      h('div', {className:'cmd-select'},
        h('span', {style:{color:'#8b8fa3'}}, '过期:'),
        expireSelect
      ),
    );
    const footerDiv = h('div', {className:'cmd-panel-footer'}, cancelBtn, submitBtn);

    const overlay = h('div', {className:'cmd-overlay'},
      h('div', {className:'cmd-panel'},
        h('div', {className:'cmd-panel-header'}, '📤 分享到 ComelyMD'),
        bodyDiv,
        footerDiv
      )
    );
    document.body.appendChild(overlay);
    syncProtectionOptions();
    burnCheck.addEventListener('change', syncProtectionOptions);

    overlay.addEventListener('click', (e) => { if (e.target === overlay) overlay.remove(); });
    cancelBtn.addEventListener('click', () => overlay.remove());

    submitBtn.addEventListener('click', () => {
      const server = serverInput.value.trim().replace(/\/+$/, '');
      if (!server) { toast('⚠️ 请填写服务器地址'); return; }

      setConfig({ server, defaultBurn: burnCheck.checked, defaultPwd: pwdCheck.checked, defaultExpire: expireSelect.value });

      submitBtn.disabled = true;
      submitBtn.textContent = '分享中…';

      const formBody = new URLSearchParams();
      formBody.append('markdown', markdownContent);
      if (burnCheck.checked) formBody.append('is_burn', 'true');
      if (pwdCheck.checked) formBody.append('with_password', 'true');
      if (expireSelect.value !== '0') formBody.append('expire_time', expireSelect.value);

      GM_xmlhttpRequest({
        method: 'POST',
        url: `${server}/api/pages`,
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        data: formBody.toString(),
        onload(res) {
          try {
            const data = JSON.parse(res.responseText);
            showResult(overlay, bodyDiv, footerDiv, `${server}/p/${data.id}`, data.pwd || '');
          } catch {
            toast('❌ 服务器响应异常');
            submitBtn.disabled = false; submitBtn.textContent = '分享';
          }
        },
        onerror() {
          toast('❌ 网络请求失败，请检查服务器地址');
          submitBtn.disabled = false; submitBtn.textContent = '分享';
        }
      });
    });
  }

  function showResult(overlay, bodyDiv, footerDiv, url, pwd) {
    bodyDiv.textContent = '';
    const urlInput = h('input', {type:'text', className:'cmd-result-url', value:url, readonly:'readonly', onclick: function(){ this.select(); }});
    const resultItems = [
      h('div', {className:'icon'}, '✅'),
      h('div', {style:{fontWeight:'600', marginBottom:'0.5rem'}}, '分享成功'),
      urlInput,
    ];
    if (pwd) {
      resultItems.push(
        h('div', {style:{fontSize:'12px', color:'#8b8fa3', marginTop:'0.5rem'}}, '密码'),
        h('div', {className:'cmd-result-pwd'}, pwd)
      );
    }
    bodyDiv.appendChild(h('div', {className:'cmd-result'}, ...resultItems));

    footerDiv.textContent = '';
    const copyBtn = h('button', {className:'cmd-btn cmd-btn-cancel'}, pwd ? '仅复制 (含密码)' : '仅复制');
    const viewBtn = h('button', {className:'cmd-btn cmd-btn-primary'}, '复制并查看');
    footerDiv.append(copyBtn, viewBtn);

    copyBtn.addEventListener('click', () => {
      clipText(pwd ? `链接: ${url}\n密码: ${pwd}` : url);
      toast('✅ 已复制到剪贴板');
      overlay.remove();
    });
    viewBtn.addEventListener('click', () => {
      clipText(pwd ? `链接: ${url}\n密码: ${pwd}` : url);
      toast('✅ 已复制并在新标签页打开');
      window.open(url, '_blank');
      overlay.remove();
    });
    clipText(url);
  }

  // ─── 设置面板 ───
  function showSettingsPanel(onSave) {
    const cfg = getConfig();
    const serverInput = h('input', {type:'text', value:cfg.server, placeholder:'https://md.example.com'});
    const cancelBtn = h('button', {className:'cmd-btn cmd-btn-cancel'}, '取消');
    const saveBtn = h('button', {className:'cmd-btn cmd-btn-primary'}, '保存');

    const overlay = h('div', {className:'cmd-overlay'},
      h('div', {className:'cmd-panel'},
        h('div', {className:'cmd-panel-header'}, '⚙️ ComelyMD 配置'),
        h('div', {className:'cmd-panel-body'},
          h('div', {className:'cmd-field'}, h('label', {}, '服务器地址'), serverInput),
          h('p', {style:{fontSize:'12px', color:'#8b8fa3', margin:'0'}}, '填写你自建的 ComelyMD 服务地址，首次使用必填。'),
        ),
        h('div', {className:'cmd-panel-footer'}, cancelBtn, saveBtn),
      )
    );
    document.body.appendChild(overlay);

    cancelBtn.addEventListener('click', () => overlay.remove());
    saveBtn.addEventListener('click', () => {
      const server = serverInput.value.trim().replace(/\/+$/, '');
      if (!server) { toast('⚠️ 服务器地址不能为空'); return; }
      GM_setValue('server', server);
      toast('✅ 配置已保存');
      overlay.remove();
      if (onSave) onSave();
    });
  }

  // 注册油猴菜单
  GM_registerMenuCommand('⚙️ ComelyMD 设置', () => showSettingsPanel());

  // ─── 平台适配器 ───
  const adapters = {

    // ── ChatGPT ──
    chatgpt: {
      name: 'ChatGPT',
      match: () => /chat(gpt)?\.openai\.com|chatgpt\.com/.test(location.hostname),
      getMessageBlocks: () => {
        const blocks = [];
        document.querySelectorAll('[data-message-author-role="assistant"]').forEach(el => {
          const article = el.closest('article') || el.closest('[data-testid^="conversation-turn"]') || el;
          if (article && !blocks.includes(article)) blocks.push(article);
        });
        if (blocks.length === 0) {
          document.querySelectorAll('article').forEach(el => {
            if (!el.querySelector('[data-message-author-role="user"]')) blocks.push(el);
          });
        }
        return blocks;
      },
      getActionBar: (block) => {
        const candidates = block.querySelectorAll('.flex.items-center');
        for (let i = candidates.length - 1; i >= 0; i--) {
          const c = candidates[i];
          if (c.querySelector('button') && c.closest('article') === block) return c;
        }
        const allBtns = block.querySelectorAll('button');
        if (allBtns.length > 0) return allBtns[allBtns.length - 1].parentElement;
        return null;
      },
      getContentHTML: (block) => {
        const contents = block.querySelectorAll('.prose, .markdown, .markdown-body');
        if (contents.length > 0) {
          const wrapper = document.createElement('div');
          contents.forEach(c => wrapper.appendChild(c.cloneNode(true)));
          return wrapper;
        }
        return block;
      },
    },

    // ── Gemini ──
    gemini: {
      name: 'Gemini',
      match: () => location.hostname.includes('gemini.google.com'),
      getMessageBlocks: () => {
        return [...document.querySelectorAll('model-response')];
      },
      getActionBar: (block) => {
        // Gemini 操作按钮是 Angular 自定义组件：copy-button, thumb-up-button 等
        // 它们的共同父容器 = buttons-container-v2
        const copyBtn = block.querySelector('copy-button');
        if (copyBtn && copyBtn.parentElement) return copyBtn.parentElement;
        const moreMenu = block.querySelector('.more-menu-button-container');
        if (moreMenu && moreMenu.parentElement) return moreMenu.parentElement;
        const thumbUp = block.querySelector('thumb-up-button');
        if (thumbUp && thumbUp.parentElement) return thumbUp.parentElement;
        return null;
      },
      getContentHTML: (block) => {
        const content = block.querySelector('.model-response-text, message-content, .response-container .markdown-main-panel');
        if (content) return content;
        const rc = block.querySelector('response-container');
        if (rc) return rc;
        return block;
      },
    },

    // ── 预留：Grok ──
    // grok: { name: 'Grok', match: () => location.hostname.includes('grok.com'), ... },
    // ── 预留：豆包 ──
    // doubao: { name: '豆包', match: () => location.hostname.includes('doubao.com'), ... },
    // ── 预留：Kimi ──
    // kimi: { name: 'Kimi', match: () => location.hostname.includes('kimi.moonshot.cn'), ... },
  };

  // ─── 核心注入逻辑 ───
  function getActiveAdapter() {
    return Object.values(adapters).find(a => a.match());
  }

  function injectShareButtons() {
    const adapter = getActiveAdapter();
    if (!adapter) return;

    const blocks = adapter.getMessageBlocks();
    blocks.forEach(block => {
      const targetActionBar = adapter.getActionBar(block);
      if (!targetActionBar) return;

      const existingBtn = block.querySelector('.cmd-share-btn');
      if (existingBtn) {
        if (existingBtn.parentElement === targetActionBar) {
          return; // 已经正确安放在当前时刻最新的有效底栏中
        } else {
          existingBtn.remove(); // 掉队在弃用隐藏的底栏里，销毁并准备重新注入
        }
      }

      block.setAttribute(PROCESSED_ATTR, '1');

      const btn = createShareIcon();
      
      // 用户需求：将 ChatGPT 的分享按钮排在最右下角
      if (adapter.name === 'ChatGPT') {
        targetActionBar.style.width = '100%';
        btn.style.marginLeft = 'auto';
      }

      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        e.preventDefault();
        // 传入 DOM 节点的克隆给 Turndown，避免 Trusted Types 限制
        const contentNode = adapter.getContentHTML(block);
        const markdown = td.turndown(contentNode.cloneNode(true));
        showPanel(markdown);
      });

      targetActionBar.appendChild(btn);
    });
  }

  // ─── 启动 ───
  function init() {
    if (!getActiveAdapter()) return;
    injectStyles();
    injectShareButtons();
    
    let debounceTimer = null;
    const observer = new MutationObserver(() => {
      if (debounceTimer) clearTimeout(debounceTimer);
      debounceTimer = setTimeout(injectShareButtons, 500);
    });
    observer.observe(document.body, { childList: true, subtree: true });

    // 增加一个定时轮询，作为单页面应用（SPA）长内容流加载的一个强力兜底。
    // 因为历史重加载时，MutationObserver 可能在异步返回之前捕捉不到结构化改变。
    setInterval(injectShareButtons, 2000);
  }

  // 监听 SPA 的路由变化，辅助触发
  const _pushState = history.pushState;
  history.pushState = function() {
    _pushState.apply(history, arguments);
    setTimeout(() => { if (typeof init !== 'undefined') injectShareButtons(); }, 1000);
  };
  const _replaceState = history.replaceState;
  history.replaceState = function() {
    _replaceState.apply(history, arguments);
    setTimeout(() => { if (typeof init !== 'undefined') injectShareButtons(); }, 1000);
  };
  window.addEventListener('popstate', () => {
    setTimeout(() => { if (typeof init !== 'undefined') injectShareButtons(); }, 1000);
  });

  if (document.readyState === 'complete') {
    setTimeout(init, 1500);
  } else {
    window.addEventListener('load', () => setTimeout(init, 1500));
  }

})();
