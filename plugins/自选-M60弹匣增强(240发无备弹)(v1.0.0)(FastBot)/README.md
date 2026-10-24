# M60 弹匣增强

将 M60 的初始弹匣容量改为可配置数值，默认 `240` 发。插件只修改 M60
武器定义中的弹匣容量，不增加后备弹药，不开放换弹，不允许从弹药堆补充，
也不改变伤害、射速或原版打空丢弃行为。

## 依赖

- SourceMod 1.11 或更高版本
- Left 4 DHooks Direct 1.126 或更高版本

本插件使用 Left 4 DHooks 的 `L4D2IWA_ClipSize` 和
`L4D2IWA_DefaultSize` 武器属性。仓库内置的 Left 4 DHooks 1.155 已满足要求。

## 安装路径

将包内 `left4dead2/` 覆盖到服务端游戏根目录：

```text
left4dead2/
|-- addons/sourcemod/plugins/l4d2_m60_clip_boost.smx
|-- addons/sourcemod/scripting/l4d2_m60_clip_boost.sp
`-- cfg/sourcemod/l4d2_m60_clip_boost.cfg
```

`.smx` 和 `.cfg` 为运行所需文件，`.sp` 仅为源码。安装后重启服务端或加载
插件并切换地图。插件不会给已经生成或正在使用的 M60 补弹。

## 配置

配置文件：`left4dead2/cfg/sourcemod/l4d2_m60_clip_boost.cfg`

```text
l4d2_m60_clip_boost_enable "1"
l4d2_m60_clip_boost_size "240"
```

- `l4d2_m60_clip_boost_enable`：`0` 关闭，`1` 开启。
- `l4d2_m60_clip_boost_size`：M60 初始弹匣容量，范围 `1` 到 `255`。

`255` 上限用于兼容 Windows 和 Linux 服务端。修改配置后切换地图，确保后续
生成的 M60 使用新容量。禁用或卸载插件时，会恢复插件加载前的武器属性。

服务端控制台或 RCON 可执行 `sm_m60_clip_boost_status`，查看开关、配置容量、
实际 `ClipSize` 和实际 `DefaultSize`，用于确认插件是否生效。

## 行为边界

- 不修改 `ammo_m60_max`，因此不会增加 M60 后备弹药。
- 不监听拾取事件，不会因丢下再拾取而自动回满。
- 不安装或替代 M60 NoDrop/AmmoPile Patch。
- 不修改现有 M60 实体，只修改之后由游戏生成的 M60 默认容量。

不要同时使用其他会修改 `weapon_rifle_m60` 的 `clip_size`、`DefaultSize` 或
拾取弹量的插件，例如 Gun Control、Info Editor 的 M60 配置、通用武器属性
覆盖或 M60 弹药补充插件；同时使用时最终结果取决于插件加载和执行顺序。
