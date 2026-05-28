# VPS 容器更新脚本

当前线上 VPS 仍然实际使用 `deploy/docker-compose/docker-compose.yaml` 这条链路，而不是 `.env + docker-compose.prod.yml`。

为了避免每次手工 `docker build`、`docker run` 或上传 `dist`，仓库新增了 `deploy/vps-update.sh`，用于在服务器仓库内直接完成：

- 可选拉取代码
- 本地构建 `server` / `web`
- 重建容器
- 健康检查
- 输出最近日志

## 适用场景

- 服务器部署目录为 `/opt/tiny-sec`
- 线上实际运行容器为 `gva-server`、`gva-web`
- 更新方式优先采用服务器本地构建
- 服务器仓库可能存在本地定制，不希望默认 `git reset --hard`

## 一次性准备

确保脚本有执行权限：

```bash
cd /opt/tiny-sec
sudo chmod +x deploy/vps-update.sh
```

首次执行 `server` 或 `all` 更新时，脚本会确保存在运行时配置文件：

```text
deploy/docker-compose/config/server-config.yaml
```

- 若服务器当前 `server/config.docker.yaml` 已包含可用的 MySQL 连接信息，脚本会自动复制一份到上述路径
- 若没有现成配置，脚本会基于 `deploy/docker-compose/config/server-config.yaml.example` 自动生成一份可用模板
- 之后 `gva-server` 会始终挂载这份宿主机配置，重建容器不会再把数据库连接信息打回镜像默认值
- 日常应备份和维护这份文件，而不是继续修改仓库跟踪的 `server/config.docker.yaml`

## 常用用法

更新全部容器，并先快进拉取 `main`：

```bash
cd /opt/tiny-sec
sudo bash deploy/vps-update.sh all --pull
```

只更新后端：

```bash
cd /opt/tiny-sec
sudo bash deploy/vps-update.sh server --pull
```

只更新前端：

```bash
cd /opt/tiny-sec
sudo bash deploy/vps-update.sh web --pull
```

不拉代码，直接基于当前工作树重建：

```bash
cd /opt/tiny-sec
sudo bash deploy/vps-update.sh all
```

禁用构建缓存：

```bash
cd /opt/tiny-sec
sudo bash deploy/vps-update.sh server --no-cache
```

## 行为说明

- 默认只基于当前工作树构建，不改 Git 状态
- `--pull` 仅做 `fetch + merge --ff-only`，不会强制覆盖本地改动
- 若检测到已跟踪文件存在改动，脚本默认会中止，避免误覆盖本地定制
- `server` 更新会自动确保 `mysql`、`redis` 已启动并健康
- `server` / `all` 更新前会自动校验运行时配置文件是否存在且包含 MySQL 连接信息
- `server` 健康检查走 `GET /health`
- `web` 健康检查走 `GET /`

## Falco 相关

当前 `docker-compose.yaml` 已为 `server` 容器增加只读挂载：

```text
../../deploy/falco-agent -> /go/src/github.com/flipped-aurora/gin-vue-admin/deploy/falco-agent
```

因此只要服务器代码更新到了新的 `deploy/falco-agent` 脚本，容器内的以下接口就能直接读取到最新脚本：

- `/falco/agent/install/installer`
- `/falco/agent/install/runtime`

## 建议 SOP

日常平台更新建议固定使用：

```bash
cd /opt/tiny-sec
git status --short
sudo bash deploy/vps-update.sh all --pull
```

如果 `git status --short` 发现服务器仓库有定制改动，先确认这些改动是否要保留，再决定是否提交、备份或手工处理；不要直接强制覆盖。
