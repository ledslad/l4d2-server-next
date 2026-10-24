# l4d2-server-next

[![Release](https://img.shields.io/github/v/release/ledslad/l4d2-server-next?label=release)](https://github.com/ledslad/l4d2-server-next/releases)
[![Manager image size](https://img.shields.io/docker/image-size/laoyutang/l4d2-manager-next/latest?label=manager%20image)](https://hub.docker.com/r/laoyutang/l4d2-manager-next)
[![L4D2 image size](https://img.shields.io/docker/image-size/laoyutang/l4d2-pure/latest?label=L4D2%20image)](https://hub.docker.com/r/laoyutang/l4d2-pure)
[![License](https://img.shields.io/github/license/ledslad/l4d2-server-next?label=license)](LICENSE)

新一代 Left 4 Dead 2 服务器与 Web 管理后台。

## 社区维护分支与鸣谢

本分支由 **菜且爱玩（ledslad）** 维护，基于 **LaoYutang** 的 [l4d2-server-next](https://github.com/LaoYutang/l4d2-server-next) 二次开发，感谢原作者及所有上游贡献者。保留原作者版权、署名与 **GNU AGPL v3** 许可证；此项目不是 GPL v3 重新许可版本。

- 本分支源码与发布：[ledslad/l4d2-server-next](https://github.com/ledslad/l4d2-server-next)。管理器的“检查本分支更新”默认查询该仓库的 Releases，不再把原作者的 Releases 当作本分支更新。
- 登录页和管理页备案号由独立站点配置决定，默认留空；可在系统管理设置或通过 `L4D2_MANAGER_ICP_NUMBER` 设置。不要将部署备案号、密码、Token 或私有运行数据提交到公开仓库。
- 2026-10-24：合并作者 `f3cd77360a4ef2064bc7b1f7fdeded0c0915641a`（含 Go 1.26.10 安全升级、地图导入风险策略及 RCON 多包修复），保留本地 Workshop 元数据与 ID 命名定制，新增可配置站点信息和独立更新来源。

项目提供 Linux Docker 与 Windows 管理器构建流程和完整 Web 面板，覆盖地图管理、插件管理、服务器状态、玩家统计、性能监控、RCON 控制台、日志、备份恢复和授权码访问等日常运维场景。**本分支公开仓库是源码发行版，不保证包含完整的 SourceMod/Metamod 和第三方插件运行二进制。** 从源码构建管理器不会自动生成这些运行文件；启用插件预设前请按[源码发行与插件安装指引](docs/community-maintenance.md)补齐并核验对应组件。

## 文档

[https://l4d2-manage-docs.laoyutang.cn/](https://l4d2-manage-docs.laoyutang.cn/)

> [!IMPORTANT]
> 本分支包含 Workshop 元数据等定制。拉取或合并作者更新前，先阅读
> [社区分支维护说明](docs/community-maintenance.md)，不要用
> `git reset --hard origin/master` 覆盖本地提交。

## 视频教程

- [一键安装脚本教程视频](https://www.bilibili.com/video/BV1p5tw6UEoi)
- [面板功能介绍视频](https://www.bilibili.com/video/BV1BN846UEQS)

## 快速部署

下面的一键脚本和默认镜像来自原作者上游，部署的是上游发行版，并非本分支构建产物。部署本分支源码请阅读[源码发行与插件安装指引](docs/community-maintenance.md)。

Linux 一键脚本：

请使用 root 用户运行（普通用户可先执行 `sudo -i`），再运行以下命令：

```sh
bash <(curl -sL l4d2-manage.laoyutang.cn)
```

官方源：

```sh
bash <(curl -sL https://raw.githubusercontent.com/LaoYutang/l4d2-server-next/master/manifest/install/manage.sh)
```

安装完成后访问：

```text
http://服务器IP:27020
```

## 主要能力

- 多平台部署：Linux Docker、Docker Compose、Windows 原生管理器、接管已有服务器。
- 地图管理：上传、下载、工坊/合集解析、任务队列、详情解析、切图、重命名、删除、资源精简、BSP 字典与全局脚本风险检查。
- 插件管理：上传 ZIP、启用/禁用、立即加载/重载/卸载、批量操作、插件商店、预设、CVAR 配置、导出。
- 服务器运维：状态首页、在线玩家操作、RCON 终端、服务器信息、`server.cfg` 常用配置。
- 统计与监控：CPU、内存、网络、磁盘实时监控，历史趋势，玩家在线统计和 Steam 游戏时长查询。
- 安全与迁移：管理员密码、临时授权码、自助授权、面板黑白名单、可信代理、备份导入导出、插件迁移。

## 预览截图

<div align="center">
  <img src="https://images.laoyutang.cn/2026/06/99465f674d128e3a406d294753147479.png" alt="首页" width="45%" />
  <img src="https://images.laoyutang.cn/2026/06/2868050dbad60124dcb313a23d422685.png" alt="性能监控" width="45%" />
  <br/>
  <img src="https://images.laoyutang.cn/2026/06/3dbe0462012ca01b6971e6744655da95.png" alt="插件管理" width="45%" />
  <img src="https://images.laoyutang.cn/2026/06/5a9cc014b0552f975dd39b2b1844d221.png" alt="地图管理" width="45%" />
</div>
