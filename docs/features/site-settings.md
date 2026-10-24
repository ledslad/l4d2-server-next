# 站点信息与版本检查

系统管理的管理员设置提供独立的“站点信息与更新来源”。

- **ICP备案号**：默认空，空值隐藏登录页和管理页备案页脚；保存后写入 `data/site_config.json`，不要把运行文件提交源码仓库。
- **GitHub 更新仓库**：格式为 `用户名/仓库名`，默认 `ledslad/l4d2-server-next`。“检查本分支更新”只查询这个仓库的最新正式 Release，失败时提示错误，不回退到原作者的版本。
- 环境变量 `L4D2_MANAGER_ICP_NUMBER`、`L4D2_MANAGER_UPDATE_REPOSITORY` 可覆盖上述值。设置页面保存后显示实际生效值；需要修改环境覆盖时，应更新部署环境后重启管理器。
- 备案链接固定指向工信部网站，备案文字以纯文本显示；不接受任意 HTML。更新仓库只接受 GitHub 的 `owner/repo` 标识，不接受任意 URL。

`POST /site/config` 提供两项公开展示配置；`POST /site/config/update` 仅管理员可写；`POST /version/latest` 提供缓存 5 分钟的 GitHub 最新版本信息。更新检查不自动替换二进制，管理员仍需阅读发布说明并按平台部署。

本社区分支由菜且爱玩（ledslad）维护，原作者 LaoYutang，继续遵循 GNU AGPL v3。公开分发与部署时保留许可证及修改说明，并提供本分支对应源代码链接。
