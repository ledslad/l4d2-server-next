# 地图工坊元数据

通过“解析链接”加入 Steam Workshop 地图下载任务时，Manager 会在下载并处理 VPK 成功后，于地图目录写入同名 `.meta` 文件。

解析接口会在 Manager 进程内按下载 URL 缓存 30 分钟元数据。即使旧版前端没有把 `workshop_meta` 随 `/download/add` 回传，紧接解析操作添加的下载任务也能自动关联 Workshop ID。

新下载的 Workshop VPK 按每个项目自身的 Workshop ID 命名，例如 `123734884.vpk` 对应 `123734884.meta`；依赖项也使用各自的 ID，避免作者使用相同原文件名时发生冲突。带工坊元数据的下载任务会强制使用该名称，覆盖调用方传入的文件名。普通直链和上传文件仍使用原有命名规则，已有地图不会自动改名。

文件采用 LYTVPK 兼容的 JSON 字段，保存 Workshop ID、标题、作者、说明、预览图、文件链接和下载时间。

- `/maps/summary` 的地图项目在 sidecar 存在且有效时包含 `workshop`。
- `/maps/detail` 在 sidecar 存在且有效时包含 `workshop`，否则该字段为 `null`。
- 手动上传或历史地图不会按文件名反查 Workshop ID，因此没有 `.meta` 时不会返回预览信息。
- 重命名和删除地图会同步处理同名 `.meta`。
- 删除和清空采用暂存、更新清单、最终清理的事务式流程；清单写入失败时会恢复 VPK 和 sidecar。
- 下载后的 sidecar 写入失败时，会撤销本次产生的 VPK、清单记录和检测记录。
