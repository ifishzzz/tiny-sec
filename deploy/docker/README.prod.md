# 生产部署（GitHub 构建 + 服务器拉镜像）

本目录提供一套面向当前项目（含 Falco 主机一期能力）的生产化 Docker 资产：

- `deploy/docker/server.prod.Dockerfile`
- `deploy/docker/web.prod.Dockerfile`
- `deploy/docker/nginx.prod.conf`
- `deploy/docker-compose/docker-compose.prod.yml`
- `deploy/docker-compose/.env.prod.example`
- `deploy/docker-compose/config/server-config.yaml.example`
- `.github/workflows/docker-prod.yml`

## 1. GitHub 自动构建镜像

工作流文件：`.github/workflows/docker-prod.yml`

触发方式：

- `push` 到 `main`
- 推送 `v*` tag
- 手动 `workflow_dispatch`

产物镜像（默认推到 GHCR）：

- `ghcr.io/<owner>/gin-vue-admin-server:<tag>`
- `ghcr.io/<owner>/gin-vue-admin-web:<tag>`

## 2. 服务器准备

```bash
sudo mkdir -p /opt/gva-prod
cd /opt/gva-prod
```

拷贝以下文件到服务器：

- `deploy/docker-compose/docker-compose.prod.yml`
- `deploy/docker-compose/.env.prod.example`（重命名为 `.env`）
- `deploy/docker-compose/config/server-config.yaml.example`

## 3. 修改配置

1) 编辑 `.env`：

- `SERVER_IMAGE`
- `WEB_IMAGE`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_PASSWORD`

2) 编辑 `SERVER_CONFIG_FILE` 指向的配置文件，默认就是：

- `config/server-config.yaml.example`

重点修改：

- `jwt.signing-key`
- `mysql.password`
- 其他业务参数（按需）

## 4. 启动

```bash
docker login ghcr.io -u <github-user>
docker compose -f docker-compose.prod.yml --env-file .env pull
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

## 5. 验证

```bash
docker compose -f docker-compose.prod.yml ps
docker logs gva-server --tail=100
docker logs gva-web --tail=100
```

## 6. Falco Agent 安装脚本可用性

`server.prod.Dockerfile` 已将 `deploy/falco-agent` 打入镜像。

因此以下接口在容器部署后仍可用：

- `/falco/agent/install/installer`
- `/falco/agent/install/runtime`
