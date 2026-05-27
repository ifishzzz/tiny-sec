# VPS 部署踩坑记录

## 环境
- VPS: Ubuntu 24 / 2vCPU 2GB (DigitalOcean)
- IP: 161.35.5.15
- 仓库: https://github.com/ifishzzz/tiny-sec

## 踩坑列表

### 1. pnpm 11 要求 Node 22+（web/Dockerfile）
**现象**: `pnpm install` 失败，`This version of pnpm requires at least Node.js v22.13`
**原因**: Dockerfile 用 `node:20-slim` + `corepack enable`，corepack 拉到 pnpm 11 但 Node 20 不兼容
**修复**: 改用 `node:22-slim` 或直接用 `npm` 替代 pnpm（没有 lock 文件时 pnpm 无优势）
**最终方案**: Dockerfile 改为 npm，去掉 pnpm/corepack

### 2. pnpm approve-builds 安全限制
**现象**: `ERR_PNPM_IGNORED_BUILDS` 错误
**原因**: pnpm 10+ 默认阻止构建脚本执行，需要 `pnpm approve-builds` 交互式审批
**修复**: 直接用 npm 替代 pnpm（CI/CD 环境不适合交互式审批）

### 3. MYSQL_ROOT_PASSWORD 被注释导致容器启动失败
**现象**: MySQL 容器反复重启 `Database is uninitialized and password option is not specified`
**原因**: docker-compose.yaml 中 `MYSQL_ROOT_PASSWORD` 被注释掉了（`#MYSQL_ROOT_PASSWORD`）
**修复**: 取消注释并设置密码

### 4. Nginx 未开启 gzip 导致 JS 传输截断（核心问题）
**现象**: 页面一直显示"系统正在加载中"，Vue 不挂载
**原因**:
  - Nginx 默认 gzip 关闭
  - 主 JS bundle 1.3MB 无压缩传输
  - VPS 低带宽（~12KB/s 下降速度），传输 ~234KB 后连接被关闭
  - 浏览器收到截断的 JS，模块执行失败但无 JS 报错
  - `Failed to fetch dynamically imported module`
**修复**: 在 `web/.docker-compose/nginx/conf.d/my.conf` 中加入 gzip 配置
**排查方法**:
  - `curl -v http://IP/assets/xxx.js -o /dev/null` 看到 `transfer closed with xxx bytes remaining`
  - 浏览器 Performance API: `decodedBodySize` 远小于实际文件大小

### 5. 生产环境隐藏了"前往初始化"按钮
**现象**: 页面直接跳到登录页，没有初始化入口
**原因**: `web/src/view/login/index.vue` 中"前往初始化"按钮受 `v-if="isDev"` 控制，`isDev = import.meta.env.DEV` 在生产构建中为 false
**修复**: 通过 curl 直接调用 `POST /api/init/initdb` API 完成初始化

### 6. 验证码阻止登录测试
**现象**: 登录需要验证码，无法通过 API 直接登录
**原因**: `config.docker.yaml` 中 `open-captcha: 0` 表示始终开启验证码
**修复**: 改为 `open-captcha: 999999`（失败 999999 次后才要求验证码）

### 7. VPS 上 npm build 产物有 bug（核心问题）
**现象**: 登录后白屏，Dashboard 不渲染，`init_reactivity_esm_bundler is not defined`
**原因**: VPS 上 `npm run build` 使用 Vite 8 + Rolldown 构建时，Vue 响应式模块未正确打包到动态 chunk 的共享作用域中
**修复**: 在本地 Mac 上 `npm run build`，将 dist 上传到 VPS 替换
**根本原因**: 可能是 VPS 环境（Node 版本、系统架构）与 Rolldown 的兼容性问题

### 8. 浏览器缓存损坏的 JS
**现象**: 修复后仍无法加载
**原因**: 之前截断的 JS 被浏览器缓存
**修复**: `chrome://settings/clearBrowserData` 清除缓存

## 正确部署流程总结

```
1. 本地构建前端 (npm run build)
2. VPS 安装 Docker
3. git clone 代码到 VPS
4. 修改 docker-compose 密码 + 生产配置
5. VPS 上构建后端 Docker 镜像
6. 上传本地 dist，构建自定义 nginx 镜像（含 gzip）
7. docker compose up mysql redis server
8. 启动自定义 web 容器
9. curl 调 API 初始化数据库
10. 浏览器访问登录
```

## 关键配置

### nginx gzip（必须）
```nginx
gzip on;
gzip_types text/plain text/css application/javascript application/json image/svg+xml;
gzip_min_length 1024;
gzip_comp_level 6;
```

### 验证码关闭
```yaml
captcha:
    open-captcha: 999999
```

### 数据库初始化 API
```bash
curl -X POST http://localhost:8888/init/initdb \
  -H 'Content-Type: application/json' \
  -d '{"dbType":"mysql","host":"177.7.0.13","port":"3306","dbName":"qmPlus","userName":"gva","password":"密码","adminPassword":"admin密码"}'
```
