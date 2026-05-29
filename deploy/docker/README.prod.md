# 统一 Docker/Compose 部署

当前仓库只保留一套新的 full-stack Docker/Compose 方案，用于本地和线上统一部署：

- `deploy/docker/server.full.Dockerfile`
- `deploy/docker/web.full.Dockerfile`
- `deploy/docker/Caddyfile`
- `deploy/docker-compose/docker-compose.full.yml`
- `deploy/docker-compose/.env.full.example`
- `deploy/docker-compose/config/server-config.full.yaml.example`

这套方案直接运行以下 4 个容器：

- `mysql`
- `redis`
- `server`
- `web`

不再使用以下旧入口：

- `deploy/vps-deploy.sh`
- `deploy/vps-update.sh`
- `deploy/docker-compose/docker-compose.yaml`
- `deploy/docker-compose/docker-compose.prod.yml`

## 1. 本地部署

在仓库根目录执行：

```bash
cp deploy/docker-compose/.env.full.example deploy/docker-compose/.env.full
cp deploy/docker-compose/config/server-config.full.yaml.example deploy/docker-compose/config/server-config.full.yaml

docker compose \
  --env-file deploy/docker-compose/.env.full \
  -f deploy/docker-compose/docker-compose.full.yml \
  up --build -d
```

如果只想看状态和日志：

```bash
docker compose \
  --env-file deploy/docker-compose/.env.full \
  -f deploy/docker-compose/docker-compose.full.yml \
  ps

docker compose \
  --env-file deploy/docker-compose/.env.full \
  -f deploy/docker-compose/docker-compose.full.yml \
  logs -f
```

停止并清理容器：

```bash
docker compose \
  --env-file deploy/docker-compose/.env.full \
  -f deploy/docker-compose/docker-compose.full.yml \
  down
```

## 2. 线上部署

在服务器部署目录执行：

```bash
cd /opt/tiny-sec

cp deploy/docker-compose/.env.full.example deploy/docker-compose/.env.full
cp deploy/docker-compose/config/server-config.full.yaml.example deploy/docker-compose/config/server-config.full.yaml
```

按实际情况修改：

- `deploy/docker-compose/.env.full`
- `deploy/docker-compose/config/server-config.full.yaml`

然后启动：

```bash
docker compose \
  --env-file deploy/docker-compose/.env.full \
  -f deploy/docker-compose/docker-compose.full.yml \
  up --build -d
```

## 3. 配置说明

### `.env.full`

主要控制：

- `WEB_PORT`
- `SERVER_PORT`
- `MYSQL_PORT`
- `REDIS_PORT`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `SERVER_CONFIG_FILE`

默认应指向：

- `./config/server-config.full.yaml`

### `server-config.full.yaml`

这是 `server` 的真实运行时配置，重点确认：

- `jwt.signing-key`
- `system.env`
- `captcha.open-captcha`
- `mysql.path`
- `mysql.db-name`
- `mysql.username`
- `mysql.password`

注意：

- 该文件用于运行时挂载，不应再依赖镜像内默认配置
- 如需长期运行，建议在线上使用 `server-config.full.yaml`，不要直接改 example 文件
- 如果 `.env.full` 里的 `SERVER_CONFIG_FILE` 没有指向 `server-config.full.yaml`，你的真实修改不会生效
- 对于首次安装，建议先保持 `mysql.path`、`mysql.port`、`mysql.db-name`、`mysql.username`、`mysql.password` 为空，让前端进入初始化页
- 初始化页中可填写 `host=mysql`、`port=3306`、`userName=root`、`password=<MYSQL_ROOT_PASSWORD>`、`dbName=qmPlus`，并自行设置管理员密码
- 不要在容器化部署的初始化页里填写 `127.0.0.1`，对 `server` 容器来说 MySQL 服务地址是 `mysql`
- 初始化成功后，程序会把数据库连接信息回写到挂载的 `server-config.full.yaml`，因此该挂载不能是只读
- 初始化成功后，后端会立即补跑一次全局自动迁移，Falco 等业务表会自动创建，不需要再手动重启 `server`

## 4. 首次初始化

如果 `POST /init/checkdb` 返回 `needInit: true`，说明数据库尚未初始化。

这时需要通过接口完成一次初始化，再进行登录验证。

## 5. 验证项

部署完成后至少验证以下内容：

```bash
curl -s http://127.0.0.1:8888/health
curl -s -X POST http://127.0.0.1:8888/init/checkdb
curl -s -X POST http://127.0.0.1:8080/api/base/captcha
```

浏览器侧验证：

- 登录页正常显示
- 初始化完成后可正常登录后台
- 登录后不白屏

## 6. Falco Agent 安装脚本可用性

新的 `server.full.Dockerfile` 仍会把 `deploy/falco-agent` 打进镜像，因此以下接口应继续可用：

- `/falco/agent/install/installer`
- `/falco/agent/install/runtime`
