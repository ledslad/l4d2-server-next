# 全伤害虚血分摊与牛撞保护

轻量 SourceMod 插件，只提供两个功能：

- 生还者受到的所有伤害按比例由实血和虚血共同承担，包括友伤、感染者、普通感染者和环境伤害。
- Charger 携带生还者期间，阻断其他生还者对受害者造成的友伤；携带结束后延迟 1 秒解除。

不包含距离保护、机器人保护、霰弹枪修正、无敌帧控制、聊天播报或手动撤销功能。

## 配置

首次运行生成 `cfg/sourcemod/l4d2_ff_temp_guard.cfg`：

```text
l4d2_ff_temp_guard_enable "1"
l4d2_ff_temp_guard_perm_fraction "0.5"
l4d2_ff_temp_guard_charger_immunity "1"
```

`l4d2_ff_temp_guard_perm_fraction` 表示实血承担比例：

- `1.0`：全部扣实血。
- `0.5`：实血 50%、虚血 50%。
- `0.0`：优先全部扣虚血，虚血不足再扣实血。

虚血不足时，未承担的部分自动转到实血。伤害总量不会因分摊而减少；插件只改变实血和虚血各自承担的部分。
奇数伤害无法平均拆分时，多出的 1 点由实血承担；例如 3 点伤害按 2 点实血、1 点虚血结算。

实现参考了 GPL-3.0 项目 [L4D2 Undo Friendly Fire Confogl Edition](https://github.com/Attano/L4D2-Competitive-Framework/blob/master/addons/sourcemod/scripting/l4d2_survivor_ff.sp) 的公开行为和边界设计；本插件源码为针对当前服务器独立实现的精简版本，随本项目使用 Apache-2.0。
