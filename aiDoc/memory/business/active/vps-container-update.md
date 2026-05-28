# VPS 容器更新脚本

## 背景

用户要求把当前线上 VPS 的容器更新流程沉淀成后续可复用的正式脚本，不再为单次发布继续使用临时手工命令或一次性上传 `dist` 的方式。

## 目标

- 提供一个可在服务器仓库内执行的更新脚本
- 支持 `server`、`web`、`all` 三种更新目标
- 默认采用服务器本地构建
- 把更新代码、重建容器、健康检查收敛为统一流程

## 已确认约束

- 当前线上实际仍使用 `deploy/docker-compose/docker-compose.yaml`
- 服务器部署目录为 `/opt/tiny-sec`
- 服务器源码目录可能存在本地定制，默认不能假设可安全 `reset --hard`
- 日常更新优先基于当前工作树或快进拉取，而不是强制覆盖
- 需要保证 `server` 容器可以稳定访问 `deploy/falco-agent`

## 落地约束

- 新脚本路径固定为 `deploy/vps-update.sh`
- 脚本默认不修改数据库初始化状态，只负责更新容器
- `server` 更新时自动确保 `mysql`、`redis` 已启动
- 健康检查优先使用后端 `/health` 和前端 `/`
- `gva-server` 运行时配置不再依赖镜像内默认 `server/config.docker.yaml`
- 服务器实际运行配置应落在 `deploy/docker-compose/config/server-config.yaml`，并通过 compose 挂载进容器，避免重建 `server` 后丢失数据库连接信息
