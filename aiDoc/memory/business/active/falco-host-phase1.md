# Falco 主机一期管理模块

## 背景

用户需要在当前 `gin-vue-admin` 仓库内落地 Falco 管理能力，用于主机场景的纳管、健康检测、规则更新、资源监控以及安装更新。

## 当前范围

- 一期只围绕主机场景建设
- 暂时不做 Kubernetes 和容器编排侧管理
- 部署环境以 AWS 为主
- Linux 首要验证基线为 Amazon Linux 2023，但安装链路需要兼容常见 Linux 发行版
- CPU 架构优先支持 `x86_64`，并补充 `aarch64`
- 事件中心一期使用 MySQL
- 接入方式采用 Agent 模式
- Agent 注册采用 `enrollKey + accessToken`
- 通信方式以 HTTPS 接口为主，预留 WebSocket 控制通道
- 后端使用独立 `falco` 业务目录，不放入 `system`，也不走插件机制

## 安装兼容性要求

- `install-falco-agent.sh` 不能再只限制 Amazon Linux 2023
- Agent 运行时下发的 `falco.install`、`falco.upgrade`、`falco.rollback` 需要自动兼容 `apt`、`dnf`、`yum`、`zypper`
- Falco 仓库初始化和非交互式安装参数参考官方 host package 文档
- 平台自身维护规则发布时，默认关闭 `falcoctl` 自动规则跟随，避免和平台规则中心冲突

## 一期模块

- 主机管理
- Agent 管理
- 安装/升级/重载任务中心
- 规则包与规则发布
- 事件中心
- 仪表盘与系统设置

## 约束

- 继续沿用现有 `Router -> API -> Service -> Model` 分层
- 菜单、API 权限和表结构需要按项目现有初始化方式接入
- 前端页面放在 `web/src/view/falco/`
