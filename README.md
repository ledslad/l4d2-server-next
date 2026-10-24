# l4d2-server-next

[![Release](https://img.shields.io/github/v/release/ledslad/l4d2-server-next?label=release)](https://github.com/ledslad/l4d2-server-next/releases)
[![Manager image size](https://img.shields.io/docker/image-size/laoyutang/l4d2-manager-next/latest?label=manager%20image)](https://hub.docker.com/r/laoyutang/l4d2-manager-next)
[![L4D2 image size](https://img.shields.io/docker/image-size/laoyutang/l4d2-pure/latest?label=L4D2%20image)](https://hub.docker.com/r/laoyutang/l4d2-pure)
[![License](https://img.shields.io/github/license/ledslad/l4d2-server-next?label=license)](LICENSE)

Left 4 Dead 2 服务器与 Web 管理后台，提供地图与插件管理、RCON、状态监控、玩家统计、备份恢复和授权访问，支持 Linux Docker 与 Windows 部署。

本项目完全由 AI 维护。

## 社区维护分支与鸣谢

本分支由 **菜且爱玩（ledslad）** 维护，基于 **LaoYutang** 的 [l4d2-server-next](https://github.com/LaoYutang/l4d2-server-next) 开发，感谢原作者及所有贡献者。项目采用 [GNU AGPL v3](LICENSE) 许可证。

源码：[ledslad/l4d2-server-next](https://github.com/ledslad/l4d2-server-next) · 发行版：[Releases](https://github.com/ledslad/l4d2-server-next/releases)

源码发行版不包含完整的第三方插件运行包，使用插件预设前请参考[插件安装指引](docs/community-maintenance.md)。

## 文档

- [上游使用手册](https://l4d2-manage-docs.laoyutang.cn/)
- [构建与插件安装](docs/community-maintenance.md)
- [站点设置与更新检查](docs/features/site-settings.md)

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
