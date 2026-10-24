## 简介

多人插件

## 可用指令

- `spec_next` 监听指令
- `sm_afk` 闲置
- `sm_teams` 团队菜单
- `sm_join` 加入生还者
- `sm_tkbot` 接管指定BOT
- `sm_away` 加入旁观者 (管理员指令)
- `sm_spec` 加入旁观者 (管理员指令)
- `sm_bot` 设置开局Bot的数量 (管理员指令)

## 本地增强

- `!afk` 支持冷却，并可选择禁用游戏自带闲置命令。
- `bots_join_respawn "1"` 使用 SteamID 防止同回合切队或退出重进创建新生命。
- `bots_reconnect_restore_state "1"` 记录退出状态：站立存活退出可在没有可用 Bot 时创建存活新 Bot；倒地、挂边或死亡退出不会创建或复活 Bot，但允许接管服务器现有的存活 Bot。
- 退出状态在团灭、回合结束或换图时清空，不记录血量和装备。
