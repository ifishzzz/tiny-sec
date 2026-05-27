#!/usr/bin/env bash
# ============================================================
# gin-vue-admin VPS 一键部署脚本
# 目标: Ubuntu 24+
# 仓库: https://github.com/ifishzzz/tiny-sec
#
# 前置条件:
#   - 本地执行，需要有 SSH 密钥访问 VPS 的权限
#   - 本地需要有 Node 20+ 和 npm（用于构建前端）
#
# 用法:
#   bash deploy/vps-deploy.sh <VPS_IP> [SSH_KEY_PATH]
#   例: bash deploy/vps-deploy.sh 161.35.5.15 ~/.ssh/id_ed25519_vps
# ============================================================
set -euo pipefail

# ---------- 参数 ----------
VPS_IP="${1:?用法: $0 <VPS_IP> [SSH_KEY_PATH]}"
SSH_KEY="${2:-$HOME/.ssh/id_ed25519_vps}"
SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=no -o ConnectTimeout=15"
REPO_URL="https://github.com/ifishzzz/tiny-sec.git"
BRANCH="main"
INSTALL_DIR="/opt/tiny-sec"
MYSQL_PASSWORD="Gva@2026Secure!"
ADMIN_PASSWORD="Admin@2026!"
# ----------------------------------------

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }

# ==========================================
# Step 1: 本地构建前端
# ==========================================
info "=== Step 1: 本地构建前端 ==="
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WEB_DIR="$PROJECT_DIR/web"

if [ ! -d "$WEB_DIR/dist" ]; then
    info "安装前端依赖..."
    cd "$WEB_DIR"
    npm install --quiet
    info "构建前端..."
    npm run build
    cd "$PROJECT_DIR"
else
    info "dist 目录已存在，跳过构建（如需重新构建请先删除 web/dist）"
fi

# ==========================================
# Step 2: VPS 准备
# ==========================================
info "=== Step 2: VPS 准备 ==="
ssh $SSH_OPTS root@$VPS_IP "bash -s" << 'REMOTE_SETUP'
set -e
if ! command -v docker &>/dev/null; then
    echo "[INFO] 安装 Docker..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
    echo "[INFO] Docker 安装完成"
else
    echo "[INFO] Docker 已安装: $(docker --version)"
fi

apt-get update -qq && apt-get install -y -qq git > /dev/null 2>&1
REMOTE_SETUP

# ==========================================
# Step 3: 拉取代码
# ==========================================
info "=== Step 3: 拉取代码 ==="
ssh $SSH_OPTS root@$VPS_IP "bash -s" << REMOTE_CLONE
set -e
if [ -d "$INSTALL_DIR" ]; then
    echo "[INFO] 代码已存在，拉取最新..."
    cd "$INSTALL_DIR"
    git fetch origin "$BRANCH"
    git reset --hard "origin/$BRANCH"
else
    echo "[INFO] 克隆仓库..."
    git clone -b "$BRANCH" "$REPO_URL" "$INSTALL_DIR"
fi
REMOTE_CLONE

# ==========================================
# Step 4: 修改配置
# ==========================================
info "=== Step 4: 修改配置 ==="
ssh $SSH_OPTS root@$VPS_IP "bash -s" << REMOTE_CONFIG
set -e
cd "$INSTALL_DIR"

# 4a. 修改 MySQL 密码（同时取消注释 MYSQL_ROOT_PASSWORD）
sed -i "s/Aa@6447985/${MYSQL_PASSWORD}/g" deploy/docker-compose/docker-compose.yaml
sed -i "s/#MYSQL_ROOT_PASSWORD.*/MYSQL_ROOT_PASSWORD: \"${MYSQL_PASSWORD}\"/" deploy/docker-compose/docker-compose.yaml

# 4b. 切换生产环境
sed -i 's/env: local/env: release/' server/config.docker.yaml

# 4c. 关闭验证码（设为 999999 次失败后才要求）
sed -i 's/open-captcha: 0/open-captcha: 999999/' server/config.docker.yaml
REMOTE_CONFIG

# ==========================================
# Step 5: 构建 Docker 镜像（server 在 VPS 上构建）
# ==========================================
info "=== Step 5: 构建后端 Docker 镜像 ==="
ssh $SSH_OPTS root@$VPS_IP "bash -s" << 'REMOTE_BUILD_SERVER'
set -e
cd /opt/tiny-sec/deploy/docker-compose
echo "[INFO] 构建后端镜像..."
docker compose build server --no-cache
REMOTE_BUILD_SERVER

# ==========================================
# Step 6: 上传本地前端 dist + 构建 web 镜像
# ==========================================
info "=== Step 6: 上传前端并构建 web 镜像 ==="
# 打包 dist
cd "$WEB_DIR"
tar czf /tmp/gva-dist.tar.gz dist/
cd "$PROJECT_DIR"

# 上传
scp $SSH_OPTS /tmp/gva-dist.tar.gz root@$VPS_IP:/tmp/

ssh $SSH_OPTS root@$VPS_IP "bash -s" << 'REMOTE_BUILD_WEB'
set -e
cd /tmp && tar xzf gva-dist.tar.gz

# 构建 nginx 镜像，把本地 dist 打包进去
cd /opt/tiny-sec
docker build -t gva-web-local -f- . << 'DOCKERFILE'
FROM nginx:alpine
COPY web/.docker-compose/nginx/conf.d/my.conf /etc/nginx/conf.d/my.conf
COPY /tmp/dist/ /usr/share/nginx/html/
RUN ls -al /usr/share/nginx/html
DOCKERFILE
REMOTE_BUILD_WEB

# ==========================================
# Step 7: 启动 MySQL + Redis + Server
# ==========================================
info "=== Step 7: 启动服务 ==="
ssh $SSH_OPTS root@$VPS_IP "bash -s" << 'REMOTE_START'
set -e
cd /opt/tiny-sec/deploy/docker-compose

# 启动 MySQL + Redis
docker compose up -d mysql redis

# 等待 MySQL 就绪
echo "[INFO] 等待 MySQL 就绪..."
for i in $(seq 1 30); do
    if docker exec gva-mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
        echo "[INFO] MySQL 已就绪"
        break
    fi
    sleep 2
done

# 启动 Server
docker compose up -d server

# 等待 Server 就绪
sleep 5

# 用本地构建的 web 镜像替换 compose 中的 web
# 先停掉原来的 web（如果有）
docker compose stop web 2>/dev/null || true
docker rm -f gva-web 2>/dev/null || true

# 启动自定义 web 容器
docker run -d \
    --name gva-web \
    --network docker-compose_network \
    --ip 177.7.0.11 \
    -p 8080:8080 \
    --restart always \
    --volumes-from gva-server:ro \
    gva-web-local

echo "[INFO] 所有服务已启动"
REMOTE_START

# ==========================================
# Step 8: 初始化数据库
# ==========================================
info "=== Step 8: 初始化数据库 ==="
ssh $SSH_OPTS root@$VPS_IP "bash -s" << REMOTE_INIT
set -e
# 检查是否已初始化
RESULT=\$(curl -s -X POST http://localhost:8888/init/checkdb)
echo "[INFO] checkdb 结果: \$RESULT"

if echo "\$RESULT" | grep -q '"needInit":true'; then
    echo "[INFO] 开始初始化数据库..."
    curl -s -X POST http://localhost:8888/init/initdb \
        -H 'Content-Type: application/json' \
        -d '{
            "dbType": "mysql",
            "host": "177.7.0.13",
            "port": "3306",
            "dbName": "qmPlus",
            "userName": "gva",
            "password": "'"${MYSQL_PASSWORD}"'",
            "adminPassword": "'"${ADMIN_PASSWORD}"'"
        }'
    echo ""
    echo "[INFO] 数据库初始化完成"
else
    echo "[INFO] 数据库已初始化，跳过"
fi
REMOTE_INIT

# ==========================================
# Step 9: 验证
# ==========================================
info "=== Step 9: 验证部署 ==="
ssh $SSH_OPTS root@$VPS_IP "bash -s" << 'REMOTE_VERIFY'
set -e
echo ""
echo "============================================"
echo "  容器状态:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "  服务验证:"
HTTP_CODE=\$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080)
echo "  前端页面: HTTP \$HTTP_CODE"
API_CODE=\$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8888/base/captcha)
echo "  后端 API: HTTP \$API_CODE"
echo "============================================"
REMOTE_VERIFY

echo ""
echo -e "${GREEN}============================================"
echo "  部署完成！"
echo "  访问: http://$VPS_IP:8080"
echo "  用户名: admin"
echo "  密码: $ADMIN_PASSWORD"
echo -e "============================================${NC}"
echo ""
echo "常用命令 (SSH 到 VPS 后执行):"
echo "  查看日志: docker logs -f gva-server"
echo "  重启所有: cd $INSTALL_DIR/deploy/docker-compose && docker compose restart"
echo "  停止所有: cd $INSTALL_DIR/deploy/docker-compose && docker compose down"
