# 源码发行与社区分支维护

本分支由菜且爱玩（ledslad）维护，基于 LaoYutang 的项目，保留管理器 GNU AGPL v3 许可及原作者鸣谢。源码和本分支 Releases 位于 [ledslad/l4d2-server-next](https://github.com/ledslad/l4d2-server-next)。

## 源码包包含什么

公开仓库提供管理器、前端、文档、构建脚本和已审核的插件文本源码/配置，不保证附带完整第三方运行包。`.sp` 与 `.inc` 是源码，不能替代服务器所需的 `.smx`；SourceMod、Metamod 和扩展还需要适合操作系统及游戏架构的 `.so` 或 `.dll`。管理器构建成功、容器启动成功都不表示这些游戏插件已经安装成功。

本轮上游基线固定为 [`f3cd77360a4ef2064bc7b1f7fdeded0c0915641a`](https://github.com/LaoYutang/l4d2-server-next/tree/f3cd77360a4ef2064bc7b1f7fdeded0c0915641a)。原版目录和版本可在该提交的 [plugins](https://github.com/LaoYutang/l4d2-server-next/tree/f3cd77360a4ef2064bc7b1f7fdeded0c0915641a/plugins) 下核对。固定提交仅确定来源，不代表已经核验每个第三方组件的再分发许可。本分支不会自动下载或重新发布许可尚未确认的二进制、数据库、客户端 VPK 和本地编译插件。

## 部署本分支及补齐插件

1. 按[构建与开发](development/build.md)构建前端和管理器，或构建 `manifest/docker/manager.Dockerfile`。容器只具备面板功能时仍可管理地图及配置；缺少插件平台时不能直接运行 SourceMod 插件或套用整套插件预设。Docker 构建会提示未发现完整运行包。
2. 从 [SourceMod](https://www.sourcemod.net/) 和 [Metamod:Source](https://www.sourcemm.net/) 官方项目获取适合 L4D2 的平台，核对版本、校验值、许可证和对应源码。若使用上面固定上游快照中的某个插件包，先检查该组件自己的版权、许可、所需扩展及对应源码，不要把整个 `plugins/` 目录统一改标为 AGPL。SourceMod 的 GPL v3 及 Valve 链接例外、插件各自的许可证均须保留；不清楚许可或找不到对应源码的组件不应重新打包发布。
3. 将完成核验的运行文件放入本地独立插件库，每个包保持 `包名/left4dead2/addons/...` 和 `包名/left4dead2/cfg/...` 的原结构。可参考固定快照的目录布局；也可按[插件 ZIP 规范](operations/plugin-package.md)通过面板导入。不要把个人服务器运行目录、授权数据库或管理员配置当作插件模板上传。
4. 原生部署通过 `L4D2_PLUGIN_STORE_PATH` 指向该插件库。Docker 可将本地插件库挂载到 `/plugins`，并设置 `L4D2_PLUGIN_STORE_PATH=/plugins`；插件库需要允许管理器写入自身插件状态。已有数据卷可能覆盖镜像中的 `/plugins`，升级时应检查实际挂载目录。
5. 在面板先启用匹配系统的插件平台，再按依赖启用 left4dhooks 等组件，最后逐个启用所需插件。重启游戏服前安排停机时间。用游戏控制台或 RCON 执行 `meta version`、`sm version` 和 `sm plugins list` 确认平台及插件确实加载，查看 `addons/sourcemod/logs/` 的错误日志。只有源码或预设文件存在不能作为安装完成依据。

仓库内一键安装脚本默认镜像仍是原作者的 `laoyutang/l4d2-pure` 和 `laoyutang/l4d2-manager-next`。它们属于上游发行渠道，不会包含本分支的新功能。要使用本分支，须自行构建管理器镜像并调整部署 Compose 的 manager 镜像；先保留数据卷和配置备份，勿将更新面板源码等同于重装游戏服。

## 合并上游与公开同步

合并前提交或备份本地改动，记录作者的精确提交，使用正常 merge 处理差异。保留 Workshop 元数据与 ID 命名、独立授权、下载排序、站点设置等本分支功能；更新依赖及功能文档，运行后端测试、前端类型检查及构建、文档构建。上游发行版与本分支发行版分开记录和发布。

个性备案号、管理密码、Token、运行数据库及本地运维说明放在运行配置或私有维护仓库，不写入公开默认值。公开发布前审查完整待提交清单和文件内容；第三方二进制只有在来源、精确文件 hash、再分发许可及对应源码均核验后才能加入明确白名单。普通公开仓库维护使用正常 Git 提交和 `git push origin main`；私有维护工作区的双仓库导出工具不随公开源码包发布。

更新检查默认读取本分支的 GitHub Releases；未发布 Release 时提示无法取得版本，不回退查询原作者版本。ICP备案默认留空，可在系统设置或通过环境变量配置，详见[站点设置](features/site-settings.md)。
