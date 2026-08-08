package internal

import "net/http"

const loginPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>云效 - Woodpecker 登录</title>
<style>
*, *::before, *::after {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  background: #f0f2f5;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  color: #333;
}
.card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
  padding: 40px 36px;
  width: 100%;
  max-width: 420px;
}
.card h1 {
  font-size: 22px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 4px;
}
.card .subtitle {
  font-size: 14px;
  color: #888;
  margin-bottom: 28px;
}
.form-group {
  margin-bottom: 20px;
}
.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 6px;
  color: #444;
}
.form-group input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
  outline: none;
}
.form-group input:focus {
  border-color: #1890ff;
  box-shadow: 0 0 0 3px rgba(24,144,255,0.12);
}
.form-group .hint {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
.form-group .hint a {
  color: #1890ff;
  text-decoration: none;
}
.form-group .hint a:hover {
  text-decoration: underline;
}
.btn {
  width: 100%;
  padding: 10px 0;
  background: #1890ff;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}
.btn:hover {
  background: #40a9ff;
}
.btn:active {
  background: #096dd9;
}
.error {
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
  padding: 10px 14px;
  color: #ff4d4f;
  font-size: 13px;
  margin-bottom: 20px;
  display: none;
}
.error.show {
  display: block;
}
</style>
</head>
<body>
<div class="card">
  <h1>云效 Woodpecker 登录</h1>
  <p class="subtitle">请输入您的云效个人访问令牌（PAT）</p>
  <div id="error" class="error"></div>
  <form id="loginForm">
    <div class="form-group">
      <label for="token">个人访问令牌</label>
      <input type="password" id="token" name="token" placeholder="请输入云效 PAT" autocomplete="off" autofocus>
      <p class="hint">
        <a href="https://help.aliyun.com/document_detail/cloudkey/pat.html" target="_blank" rel="noopener">
          如何获取个人访问令牌？
        </a>
      </p>
    </div>
    <button type="submit" class="btn">登录</button>
  </form>
</div>
<script>
(function() {
  var params = new URLSearchParams(window.location.search);
  var redirect = params.get('woodpecker_host');
  var errorEl = document.getElementById('error');
  document.getElementById('loginForm').addEventListener('submit', function(e) {
    e.preventDefault();
    var token = document.getElementById('token').value.trim();
    if (!token) {
      errorEl.textContent = '请输入个人访问令牌';
      errorEl.classList.add('show');
      return;
    }
    if (!redirect) {
      errorEl.textContent = '缺少 Woodpecker 服务地址参数，请从 Woodpecker 登录页面重新发起登录';
      errorEl.classList.add('show');
      return;
    }
    errorEl.classList.remove('show');
    window.location.href = redirect + '/authorize?code=' + encodeURIComponent(token);
  });
})();
</script>
</body>
</html>`

func (f *Forge) LoginHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(loginPageHTML))
}
