#pragma semicolon 1
#pragma newdecls required

#include <sourcemod>
#include <sdktools>
#include <l4d2_source_keyvalues>

#define PLUGIN_VERSION "1.2.0"
#define DATA_FILE "data/l4d2_custom_campaign_list.cfg"
#define CONFIG_FILE "l4d2_custom_campaign_list"
#define MENU_TIME 30

enum struct CampaignEntry
{
	char mission[64];
	char title[128];
	char firstMap[64];
}

Address g_pMatchExtL4D;
Handle g_hSDKGetAllMissions;
ArrayList g_Queue;
ArrayList g_Draft;
CampaignEntry g_CurrentCampaign;
ConVar g_cvGameMode;
ConVar g_cvAdvertInterval;
ConVar g_cvEmptyConfirmDelay;
ConVar g_cvVersion;
Handle g_hAdvertTimer;

bool g_bHasList;
bool g_bActive;
bool g_bPaused;
bool g_bServerEmpty;
bool g_bMapChanger;
bool g_bCurrentCustom;
bool g_bHasCurrentCampaign;
bool g_bCurrentCompleted;
bool g_bDraftRegenerated;
bool g_bLegacyState;
int g_iEditor;
int g_iMapReadyAttempts;
int g_iMapGeneration;
int g_iPopulationGeneration;
char g_sGameMode[32];
char g_sCurrentMission[64];
char g_sCurrentTitle[128];
char g_sCurrentFirstMap[64];
char g_sTrackedMap[64];
char g_sResumeMission[64];
char g_sResumeTitle[128];
char g_sResumeMap[64];

native bool MC_SetNextMap(const char[] map);
native void MC_ClearNextMap();

public Plugin myinfo =
{
	name = "L4D2 Custom Campaign List",
	author = "FastBot",
	description = "Admin-managed one-shot custom campaign rotation",
	version = PLUGIN_VERSION,
	url = ""
};

public APLRes AskPluginLoad2(Handle myself, bool late, char[] error, int errMax)
{
	MarkNativeAsOptional("MC_SetNextMap");
	MarkNativeAsOptional("MC_ClearNextMap");
	RegPluginLibrary("l4d2_custom_campaign_list");
	return APLRes_Success;
}

public void OnLibraryAdded(const char[] name)
{
	if (strcmp(name, "map_changer") == 0)
	{
		g_bMapChanger = true;
		if (g_bPaused)
			ClearMapChangerNextMap();
		else if (g_Queue != null && g_bActive)
			ApplyQueueNextMap();
	}
}

public void OnLibraryRemoved(const char[] name)
{
	if (strcmp(name, "map_changer") == 0)
		g_bMapChanger = false;
}

public void OnPluginStart()
{
	InitMissionSDK();

	g_Queue = new ArrayList(sizeof CampaignEntry);
	g_Draft = new ArrayList(sizeof CampaignEntry);
	g_bMapChanger = LibraryExists("map_changer");

	g_cvVersion = CreateConVar("l4d2_custom_campaign_list_version", PLUGIN_VERSION,
		"Custom campaign list plugin version.", FCVAR_NOTIFY | FCVAR_DONTRECORD);
	g_cvVersion.SetString(PLUGIN_VERSION);
	g_cvAdvertInterval = CreateConVar("clist_ad_interval", "300",
		"下一张三方战役聊天提醒间隔（秒），0 关闭。", FCVAR_NONE, true, 0.0);
	g_cvEmptyConfirmDelay = CreateConVar("clist_empty_confirm_delay", "30",
		"服务器无真人后确认暂停轮换的等待时间（秒）。", FCVAR_NONE, true, 5.0, true, 300.0);
	g_cvAdvertInterval.AddChangeHook(OnAdvertIntervalChanged);
	AutoExecConfig(true, CONFIG_FILE);

	g_cvGameMode = FindConVar("mp_gamemode");
	if (g_cvGameMode != null)
	{
		g_cvGameMode.GetString(g_sGameMode, sizeof g_sGameMode);
		g_cvGameMode.AddChangeHook(OnGameModeChanged);
	}
	else
	{
		strcopy(g_sGameMode, sizeof g_sGameMode, "coop");
	}

	RegAdminCmd("sm_clist", Command_CList, ADMFLAG_CHANGEMAP,
		"生成、查看并管理三方战役轮换列表");
	HookEvent("finale_win", Event_CurrentCampaignCompleted, EventHookMode_PostNoCopy);
	HookEvent("finale_vehicle_leaving", Event_CurrentCampaignCompleted, EventHookMode_PostNoCopy);

	LoadQueue();
	g_bServerEmpty = false;
	if (g_bPaused)
		ClearMapChangerNextMap();
	g_iMapGeneration = 1;
	CreateTimer(1.0, Timer_MapReady, g_iMapGeneration);
	ScheduleServerPopulationUpdate(g_cvEmptyConfirmDelay.FloatValue);
}

public void OnConfigsExecuted()
{
	if (g_cvGameMode != null)
		g_cvGameMode.GetString(g_sGameMode, sizeof g_sGameMode);
	RestartAdvertTimer();
}

public void OnMapStart()
{
	g_iEditor = 0;
	g_iMapReadyAttempts = 0;
	g_iMapGeneration++;
	g_Draft.Clear();
	g_sCurrentMission[0] = '\0';
	g_sCurrentTitle[0] = '\0';
	g_sCurrentFirstMap[0] = '\0';
	g_bCurrentCustom = false;
	if (g_bPaused)
		ClearMapChangerNextMap();
	CreateTimer(1.0, Timer_MapReady, g_iMapGeneration);
	ScheduleServerPopulationUpdate(g_cvEmptyConfirmDelay.FloatValue);
	RestartAdvertTimer();
}

public void OnClientDisconnect(int client)
{
	if (client == g_iEditor)
		ReleaseEditor(client);
	if (!IsFakeClient(client))
		ScheduleServerPopulationUpdate(g_cvEmptyConfirmDelay.FloatValue);
}

public void OnClientConnected(int client)
{
	if (!IsFakeClient(client))
		ScheduleServerPopulationUpdate(0.2);
}

public void OnClientPutInServer(int client)
{
	if (!IsFakeClient(client))
		ScheduleServerPopulationUpdate(0.2);
}

public void OnGameModeChanged(ConVar convar, const char[] oldValue, const char[] newValue)
{
	strcopy(g_sGameMode, sizeof g_sGameMode, newValue);
	g_iMapReadyAttempts = 0;
	CreateTimer(0.5, Timer_MapReady, g_iMapGeneration);
}

public void OnAdvertIntervalChanged(ConVar convar, const char[] oldValue, const char[] newValue)
{
	RestartAdvertTimer();
}

public Action Timer_MapReady(Handle timer, any generation)
{
	if (generation != g_iMapGeneration)
		return Plugin_Stop;
	g_iMapReadyAttempts++;
	RefreshCurrentCampaign();
	ConsumeResumePoint();
	UpdateCurrentCampaignProgress();
	if (g_bActive && !g_bPaused)
		ApplyQueueNextMap();
	if (g_sCurrentMission[0] == '\0' && g_iMapReadyAttempts < 10)
		CreateTimer(1.0, Timer_MapReady, g_iMapGeneration);
	return Plugin_Stop;
}

void ScheduleServerPopulationUpdate(float delay)
{
	g_iPopulationGeneration++;
	CreateTimer(delay, Timer_UpdateServerEmpty, g_iPopulationGeneration, TIMER_FLAG_NO_MAPCHANGE);
}

public Action Timer_UpdateServerEmpty(Handle timer, any generation)
{
	if (generation != g_iPopulationGeneration)
		return Plugin_Stop;

	bool empty = !HasHumanPlayers();
	if (empty == g_bServerEmpty)
		return Plugin_Stop;

	g_bServerEmpty = empty;
	if (empty)
	{
		PauseRotationForEmptyServer();
	}
	else if (g_bPaused)
	{
		PrintToChatAll("\x04[轮换]\x01 三方战役轮换仍处于暂停状态，管理员输入 !clist 可手动继续。");
	}
	return Plugin_Stop;
}

public Action Command_CList(int client, int args)
{
	if (client == 0)
	{
		PrintCampaignScanToServer();
		return Plugin_Handled;
	}
	if (!IsAuthorizedAdmin(client))
		return Plugin_Handled;

	ShowRootMenu(client);
	return Plugin_Handled;
}

void PrintCampaignScanToServer()
{
	RefreshCurrentCampaign();
	ArrayList campaigns = new ArrayList(sizeof CampaignEntry);
	BuildCustomCampaignList(campaigns, false);
	PrintToServer("[轮换] 当前战役：%s（%s）", g_sCurrentTitle,
		g_bCurrentCustom ? "三方" : "官方或未知");
	PrintToServer("[轮换] 扫描到 %d 个有效三方战役：", campaigns.Length);

	CampaignEntry entry;
	for (int i = 0; i < campaigns.Length; i++)
	{
		campaigns.GetArray(i, entry);
		PrintToServer("  %d. %s | mission=%s | first=%s", i + 1,
			entry.title, entry.mission, entry.firstMap);
	}
	delete campaigns;
}

void ShowRootMenu(int client)
{
	Menu menu = new Menu(RootMenuHandler);
	char title[256];
	if (g_bCurrentCustom)
		Format(title, sizeof title, "三方战役轮换\n当前：%s\n待播放：%d%s", g_sCurrentTitle, g_Queue.Length,
			g_bPaused ? "\n状态：已暂停" : "");
	else
		Format(title, sizeof title, "三方战役轮换\n待播放：%d%s", g_Queue.Length,
			g_bPaused ? "\n状态：已暂停" : "");
	menu.SetTitle(title);
	menu.AddItem("generate", "生成/重置列表");
	menu.AddItem("view", "查看/调整列表", g_bHasList ? ITEMDRAW_DEFAULT : ITEMDRAW_DISABLED);
	if (g_bPaused)
	{
		char pauseLabel[192];
		if (g_sResumeMap[0] != '\0')
			Format(pauseLabel, sizeof pauseLabel, "继续未完成战役：%s", g_sResumeTitle);
		else
			strcopy(pauseLabel, sizeof pauseLabel, "继续轮换");
		menu.AddItem("resume", pauseLabel);
	}
	else if (g_bActive || (g_bHasList && g_bCurrentCustom))
	{
		menu.AddItem("pause", "暂停轮换");
	}

	if (!g_bPaused && g_bHasList && g_Queue.Length > 0)
	{
		CampaignEntry next;
		g_Queue.GetArray(0, next);
		char label[192];
		if (g_bActive && g_bCurrentCustom)
			Format(label, sizeof label, "跳过当前，立即切换到：%s", next.title);
		else
			Format(label, sizeof label, "开始轮换，立即切换到：%s", next.title);
		menu.AddItem("start", label);
	}
	else if (g_bHasList && g_Queue.Length == 0 && !g_bPaused)
	{
		menu.AddItem("done", "本轮三方战役已全部完成", ITEMDRAW_DISABLED);
	}

	menu.ExitButton = true;
	menu.Display(client, MENU_TIME);
}

public int RootMenuHandler(Menu menu, MenuAction action, int client, int item)
{
	if (action == MenuAction_End)
	{
		delete menu;
		return 0;
	}
	if (action != MenuAction_Select || !IsAuthorizedAdmin(client))
		return 0;

	char info[32];
	menu.GetItem(item, info, sizeof info);
	if (strcmp(info, "generate") == 0)
		ShowGenerateConfirmMenu(client);
	else if (strcmp(info, "view") == 0)
		BeginEditing(client, false);
	else if (strcmp(info, "start") == 0)
		ShowImmediateSwitchConfirm(client);
	else if (strcmp(info, "pause") == 0)
		PauseRotation(client);
	else if (strcmp(info, "resume") == 0)
		ResumeRotation(client);
	return 0;
}

void PauseRotation(int client)
{
	CaptureCurrentResumePoint();
	g_bPaused = true;
	SaveQueue();
	ClearMapChangerNextMap();
	PrintToChatAll("\x04[轮换]\x01 管理员已暂停三方战役轮换，当前未完成战役已保留。");
	ShowRootMenu(client);
}

void PauseRotationForEmptyServer()
{
	if (g_bPaused || (!g_bActive && !(g_bHasList && g_bCurrentCustom)))
		return;

	CaptureCurrentResumePoint();
	g_bPaused = true;
	SaveQueue();
	ClearMapChangerNextMap();
	PrintToServer("[轮换] 服务器已无真人，轮换自动暂停，等待管理员手动继续。");
}

void ResumeRotation(int client)
{
	char currentMap[64];
	GetCurrentMap(currentMap, sizeof currentMap);
	NormalizeMapName(currentMap, sizeof currentMap);

	if (g_sResumeMap[0] != '\0' && strcmp(currentMap, g_sResumeMap, false) != 0)
	{
		if (!IsMapValidEx(g_sResumeMap))
		{
			PrintToChat(client, "\x04[轮换]\x01 未完成战役的小关 %s 当前不可用，未恢复轮换。", g_sResumeMap);
			return;
		}

		g_bPaused = false;
		if (g_Queue.Length > 0)
			g_bActive = true;
		SaveQueue();
		if (g_bActive)
			ApplyQueueNextMap();
		PrintToChatAll("\x04[轮换]\x01 管理员继续未完成战役：\x05%s\x01。", g_sResumeTitle);
		ServerCommand("changelevel \"%s\"", g_sResumeMap);
		return;
	}

	ClearResumePoint();
	g_bPaused = false;
	if (g_Queue.Length > 0)
		g_bActive = true;
	SaveQueue();
	if (g_bActive && ApplyQueueNextMap())
		AnnounceNextCampaign("轮换已手动继续；完成当前战役后进入");
	else
		PrintToChat(client, "\x04[轮换]\x01 轮换已继续，当前没有待播放的下一张三方战役。");
	ShowRootMenu(client);
}

void ShowGenerateConfirmMenu(int client)
{
	Menu menu = new Menu(GenerateConfirmMenuHandler);
	menu.SetTitle("重新生成将覆盖当前顺序和进度，并恢复本轮已删除或玩过的战役。\n是否继续？");
	menu.AddItem("yes", "确认重新生成");
	menu.AddItem("no", "返回");
	menu.ExitButton = false;
	menu.Display(client, MENU_TIME);
}

public int GenerateConfirmMenuHandler(Menu menu, MenuAction action, int client, int item)
{
	if (action == MenuAction_End)
	{
		delete menu;
		return 0;
	}
	if (action != MenuAction_Select || !IsAuthorizedAdmin(client))
		return 0;

	char info[16];
	menu.GetItem(item, info, sizeof info);
	if (strcmp(info, "yes") == 0)
		BeginEditing(client, true);
	else
		ShowRootMenu(client);
	return 0;
}

void BeginEditing(int client, bool regenerate)
{
	if (g_iEditor != 0 && g_iEditor != client && IsClientInGame(g_iEditor))
	{
		PrintToChat(client, "\x04[轮换]\x01 另一位管理员正在调整列表，请稍后再试。");
		return;
	}

	g_iEditor = client;
	g_bDraftRegenerated = regenerate;
	g_Draft.Clear();
	RefreshCurrentCampaign();
	if (regenerate)
	{
		BuildCustomCampaignList(g_Draft, true);
		if (g_Draft.Length == 0)
			PrintToChat(client, "\x04[轮换]\x01 没有找到当前战役之外的有效三方战役。");
	}
	else
	{
		CopyCampaignList(g_Queue, g_Draft);
	}
	ShowDraftMenu(client);
}

void ShowDraftMenu(int client)
{
	if (!OwnsEditor(client))
		return;

	Menu menu = new Menu(DraftMenuHandler);
	char title[192];
	Format(title, sizeof title, "待播放三方战役（顶部为下一张）\n共 %d 张", g_Draft.Length);
	menu.SetTitle(title);

	CampaignEntry entry;
	char info[32];
	for (int i = 0; i < g_Draft.Length; i++)
	{
		g_Draft.GetArray(i, entry);
		Format(info, sizeof info, "campaign:%d", i);
		menu.AddItem(info, entry.title);
	}
	menu.AddItem("confirm", "确认列表", g_Draft.Length > 0 ? ITEMDRAW_DEFAULT : ITEMDRAW_DISABLED);
	menu.ExitButton = true;
	menu.Display(client, MENU_TIME);
}

public int DraftMenuHandler(Menu menu, MenuAction action, int client, int item)
{
	if (action == MenuAction_End)
	{
		delete menu;
		return 0;
	}
	if (action == MenuAction_Cancel)
	{
		if (OwnsEditor(client))
			ReleaseEditor(client);
		return 0;
	}
	if (action != MenuAction_Select || !OwnsEditor(client) || !IsAuthorizedAdmin(client))
		return 0;

	char info[32];
	menu.GetItem(item, info, sizeof info);
	if (strcmp(info, "confirm") == 0)
	{
		ShowListConfirmMenu(client);
		return 0;
	}

	int index = ParseCampaignIndex(info);
	if (index >= 0 && index < g_Draft.Length)
		ShowCampaignActionMenu(client, index);
	return 0;
}

void ShowCampaignActionMenu(int client, int index)
{
	CampaignEntry entry;
	g_Draft.GetArray(index, entry);
	Menu menu = new Menu(CampaignActionMenuHandler);
	char title[192];
	Format(title, sizeof title, "%s\n选择操作", entry.title);
	menu.SetTitle(title);

	char info[32];
	Format(info, sizeof info, "start:%d", index);
	menu.AddItem(info, "立即开始此战役");
	Format(info, sizeof info, "top:%d", index);
	menu.AddItem(info, "移到最前（设为下一张）", index == 0 ? ITEMDRAW_DISABLED : ITEMDRAW_DEFAULT);
	Format(info, sizeof info, "bottom:%d", index);
	menu.AddItem(info, "移到队尾", index == g_Draft.Length - 1 ? ITEMDRAW_DISABLED : ITEMDRAW_DEFAULT);
	Format(info, sizeof info, "delete:%d", index);
	menu.AddItem(info, "从本轮列表删除");
	menu.ExitBackButton = true;
	menu.Display(client, MENU_TIME);
}

public int CampaignActionMenuHandler(Menu menu, MenuAction action, int client, int item)
{
	if (action == MenuAction_End)
	{
		delete menu;
		return 0;
	}
	if (action == MenuAction_Cancel)
	{
		if (item == MenuCancel_ExitBack && OwnsEditor(client))
			ShowDraftMenu(client);
		else if (OwnsEditor(client))
			ReleaseEditor(client);
		return 0;
	}
	if (action != MenuAction_Select || !OwnsEditor(client) || !IsAuthorizedAdmin(client))
		return 0;

	char info[32];
	menu.GetItem(item, info, sizeof info);
	char parts[2][16];
	if (ExplodeString(info, ":", parts, sizeof parts, sizeof parts[]) != 2)
		return 0;
	int index = StringToInt(parts[1]);
	if (index < 0 || index >= g_Draft.Length)
		return 0;

	CampaignEntry entry;
	g_Draft.GetArray(index, entry);
	if (strcmp(parts[0], "start") == 0)
	{
		ShowSelectedCampaignStartConfirm(client, index);
		return 0;
	}
	else if (strcmp(parts[0], "top") == 0)
	{
		g_Draft.Erase(index);
		g_Draft.ShiftUp(0);
		g_Draft.SetArray(0, entry);
	}
	else if (strcmp(parts[0], "bottom") == 0)
	{
		g_Draft.Erase(index);
		g_Draft.PushArray(entry);
	}
	else if (strcmp(parts[0], "delete") == 0)
	{
		g_Draft.Erase(index);
	}
	ShowDraftMenu(client);
	return 0;
}

void ShowSelectedCampaignStartConfirm(int client, int index)
{
	if (!OwnsEditor(client) || index < 0 || index >= g_Draft.Length)
		return;

	CampaignEntry entry;
	g_Draft.GetArray(index, entry);
	Menu menu = new Menu(SelectedCampaignStartConfirmMenuHandler);
	char title[256];
	Format(title, sizeof title, "立即开始此战役？\n%s\n首章：%s", entry.title, entry.firstMap);
	menu.SetTitle(title);

	char info[32];
	Format(info, sizeof info, "yes:%d", index);
	menu.AddItem(info, "确认立即切换");
	menu.AddItem("back", "返回调整");
	menu.ExitButton = false;
	menu.Display(client, MENU_TIME);
}

public int SelectedCampaignStartConfirmMenuHandler(Menu menu, MenuAction action, int client, int item)
{
	if (action == MenuAction_End)
	{
		delete menu;
		return 0;
	}
	if (action != MenuAction_Select || !OwnsEditor(client) || !IsAuthorizedAdmin(client))
		return 0;

	char info[32];
	menu.GetItem(item, info, sizeof info);
	if (strcmp(info, "back") == 0)
	{
		ShowDraftMenu(client);
		return 0;
	}

	char parts[2][16];
	if (ExplodeString(info, ":", parts, sizeof parts, sizeof parts[]) != 2 || strcmp(parts[0], "yes") != 0)
		return 0;
	int index = StringToInt(parts[1]);
	if (index < 0 || index >= g_Draft.Length)
		return 0;

	CopyCampaignList(g_Draft, g_Queue);
	g_bHasList = true;
	g_bActive = true;
	g_bPaused = false;
	ClearResumePoint();
	if (g_bDraftRegenerated)
		ResetTrackedCampaignFromRuntime();
	SaveQueue();
	ReleaseEditor(client);
	ExecuteCampaignSwitch(client, index);
	return 0;
}

void ExecuteCampaignSwitch(int client, int index)
{
	if (index < 0 || index >= g_Queue.Length)
	{
		PrintToChat(client, "\x04[轮换]\x01 选择的战役已不在列表中。");
		return;
	}

	CampaignEntry selected;
	g_Queue.GetArray(index, selected);
	if (!IsMapValidEx(selected.firstMap))
	{
		PrintToChat(client, "\x04[轮换]\x01 地图 %s 当前不可用，未执行切换。", selected.firstMap);
		return;
	}

	if (g_bHasCurrentCampaign)
	{
		g_bCurrentCompleted = true;
		SaveQueue();
	}
	PrintToChatAll("\x04[轮换]\x01 管理员选择立即开始：\x05%s\x01。", selected.title);
	ServerCommand("changelevel \"%s\"", selected.firstMap);
}

void ShowListConfirmMenu(int client)
{
	if (!OwnsEditor(client) || g_Draft.Length == 0)
		return;

	CampaignEntry first;
	g_Draft.GetArray(0, first);
	Menu menu = new Menu(ListConfirmMenuHandler);
	char title[256];
	Format(title, sizeof title, "是否从 %s 的首章开始轮换？\n首章：%s", first.title, first.firstMap);
	menu.SetTitle(title);
	menu.AddItem("now", "立即切换");
	menu.AddItem("later", "稍后切换（当前战役结束后）");
	menu.AddItem("back", "返回调整");
	menu.ExitButton = false;
	menu.Display(client, MENU_TIME);
}

public int ListConfirmMenuHandler(Menu menu, MenuAction action, int client, int item)
{
	if (action == MenuAction_End)
	{
		delete menu;
		return 0;
	}
	if (action != MenuAction_Select || !OwnsEditor(client) || !IsAuthorizedAdmin(client))
		return 0;

	char info[16];
	menu.GetItem(item, info, sizeof info);
	if (strcmp(info, "back") == 0)
	{
		ShowDraftMenu(client);
		return 0;
	}

	CopyCampaignList(g_Draft, g_Queue);
	g_bHasList = true;
	g_bActive = true;
	g_bPaused = false;
	ClearResumePoint();
	if (g_bDraftRegenerated)
		ResetTrackedCampaignFromRuntime();
	SaveQueue();
	ReleaseEditor(client);

	if (!ApplyQueueNextMap())
	{
		PrintToChat(client, "\x04[轮换]\x01 无法设置下一张战役，请检查 map_changer 和地图文件。");
		return 0;
	}

	if (strcmp(info, "now") == 0)
		ExecuteImmediateSwitch(client);
	else
		AnnounceNextCampaign("轮换已启用，当前战役结束后进入");
	return 0;
}

void ShowImmediateSwitchConfirm(int client)
{
	if (g_Queue.Length == 0)
	{
		PrintToChat(client, "\x04[轮换]\x01 待播放列表为空。");
		return;
	}

	CampaignEntry next;
	g_Queue.GetArray(0, next);
	Menu menu = new Menu(ImmediateSwitchConfirmMenuHandler);
	char title[256];
	if (g_bCurrentCustom)
		Format(title, sizeof title, "当前：%s\n下一张：%s\n确认跳过当前战役并立即切换？", g_sCurrentTitle, next.title);
	else
		Format(title, sizeof title, "下一张：%s\n确认结束当前地图并立即切换？", next.title);
	menu.SetTitle(title);
	menu.AddItem("yes", "确认立即切换");
	menu.AddItem("no", "返回");
	menu.ExitButton = false;
	menu.Display(client, MENU_TIME);
}

public int ImmediateSwitchConfirmMenuHandler(Menu menu, MenuAction action, int client, int item)
{
	if (action == MenuAction_End)
	{
		delete menu;
		return 0;
	}
	if (action != MenuAction_Select || !IsAuthorizedAdmin(client))
		return 0;

	char info[16];
	menu.GetItem(item, info, sizeof info);
	if (strcmp(info, "yes") != 0)
	{
		ShowRootMenu(client);
		return 0;
	}

	ExecuteImmediateSwitch(client);
	return 0;
}

void ExecuteImmediateSwitch(int client)
{
	if (g_Queue.Length == 0)
	{
		PrintToChat(client, "\x04[轮换]\x01 待播放列表已经为空。");
		return;
	}

	CampaignEntry next;
	g_Queue.GetArray(0, next);
	if (!IsMapValidEx(next.firstMap))
	{
		PrintToChat(client, "\x04[轮换]\x01 地图 %s 当前不可用，未执行切换。", next.firstMap);
		return;
	}

	g_bActive = true;
	g_bPaused = false;
	ClearResumePoint();
	if (g_bHasCurrentCampaign)
		g_bCurrentCompleted = true;
	SaveQueue();
	PrintToChatAll("\x04[轮换]\x01 管理员选择立即切换到：\x05%s\x01。", next.title);
	ServerCommand("changelevel \"%s\"", next.firstMap);
}

void RefreshCurrentCampaign()
{
	g_sCurrentMission[0] = '\0';
	g_sCurrentTitle[0] = '\0';
	g_sCurrentFirstMap[0] = '\0';
	g_bCurrentCustom = false;

	char currentMap[64];
	GetCurrentMap(currentMap, sizeof currentMap);
	NormalizeMapName(currentMap, sizeof currentMap);

	SourceKeyValues missions = SDKCall(g_hSDKGetAllMissions, g_pMatchExtL4D);
	for (SourceKeyValues mission = missions.GetFirstTrueSubKey(); !mission.IsNull(); mission = mission.GetNextTrueSubKey())
	{
		SourceKeyValues mode = FindMissionMode(mission);
		if (mode.IsNull())
			continue;

		char firstMap[64];
		for (SourceKeyValues chapter = mode.GetFirstTrueSubKey(); !chapter.IsNull(); chapter = chapter.GetNextTrueSubKey())
		{
			char map[64];
			chapter.GetString("Map", map, sizeof map, "");
			NormalizeMapName(map, sizeof map);
			if (firstMap[0] == '\0')
				strcopy(firstMap, sizeof firstMap, map);
			if (strcmp(map, currentMap, false) != 0)
				continue;

			mission.GetName(g_sCurrentMission, sizeof g_sCurrentMission);
			mission.GetString("DisplayTitle", g_sCurrentTitle, sizeof g_sCurrentTitle, g_sCurrentMission);
			strcopy(g_sCurrentFirstMap, sizeof g_sCurrentFirstMap, firstMap);
			g_bCurrentCustom = mission.GetInt("builtin") == 0;
			return;
		}
	}

	strcopy(g_sCurrentTitle, sizeof g_sCurrentTitle, currentMap);
}

void UpdateCurrentCampaignProgress()
{
	if (!g_bHasList || g_sCurrentMission[0] == '\0')
		return;

	char currentMap[64];
	GetCurrentMap(currentMap, sizeof currentMap);
	NormalizeMapName(currentMap, sizeof currentMap);

	if (!g_bCurrentCustom)
	{
		HandleMapOutsideTrackedCampaign();
		return;
	}

	if (g_bHasCurrentCampaign && strcmp(g_CurrentCampaign.mission, g_sCurrentMission, false) == 0)
	{
		if (strcmp(g_sTrackedMap, currentMap, false) != 0)
		{
			strcopy(g_sTrackedMap, sizeof g_sTrackedMap, currentMap);
			g_bCurrentCompleted = false;
			SaveQueue();
			PrintToServer("[轮换] 继续 %s，当前小关 %s。", g_CurrentCampaign.title, g_sTrackedMap);
		}
		return;
	}
	if (g_bPaused && g_sResumeMap[0] != '\0' && strcmp(currentMap, g_sResumeMap, false) != 0)
		return;

	int index = -1;
	index = FindCampaignByMission(g_Queue, g_sCurrentMission);
	if (index == -1)
	{
		CampaignEntry candidate;
		for (int i = 0; i < g_Queue.Length; i++)
		{
			g_Queue.GetArray(i, candidate);
			if (strcmp(candidate.firstMap, currentMap, false) == 0)
			{
				index = i;
				break;
			}
		}
	}

	if (index != -1)
	{
		CampaignEntry entered;
		g_Queue.GetArray(index, entered);
		char previousTitle[128];
		if (g_bHasCurrentCampaign)
			strcopy(previousTitle, sizeof previousTitle, g_CurrentCampaign.title);

		strcopy(g_CurrentCampaign.mission, sizeof g_CurrentCampaign.mission, entered.mission);
		strcopy(g_CurrentCampaign.title, sizeof g_CurrentCampaign.title, entered.title);
		strcopy(g_CurrentCampaign.firstMap, sizeof g_CurrentCampaign.firstMap, entered.firstMap);
		strcopy(g_sTrackedMap, sizeof g_sTrackedMap, currentMap);
		g_bHasCurrentCampaign = true;
		g_bCurrentCompleted = false;
		g_Queue.Erase(index);
		if (g_Queue.Length == 0)
			g_bActive = false;
		SaveQueue();

		if (previousTitle[0] != '\0')
			PrintToServer("[轮换] 已结束 %s，进入 %s；后续剩余 %d 张三方战役。", previousTitle, entered.title, g_Queue.Length);
		else
			PrintToServer("[轮换] 已进入 %s 并记录为进行中；后续剩余 %d 张三方战役。", entered.title, g_Queue.Length);
		return;
	}

	if (!g_bHasCurrentCampaign && (g_bLegacyState || g_bActive || g_bPaused || g_Queue.Length > 0))
	{
		ResetTrackedCampaignFromRuntime();
		SaveQueue();
		PrintToServer("[轮换] 已恢复进行中的战役 %s，当前小关 %s。", g_CurrentCampaign.title, g_sTrackedMap);
		return;
	}

	HandleMapOutsideTrackedCampaign();
}

void ResetTrackedCampaignFromRuntime()
{
	ClearTrackedCampaign();
	if (!g_bCurrentCustom || g_sCurrentMission[0] == '\0')
		return;

	strcopy(g_CurrentCampaign.mission, sizeof g_CurrentCampaign.mission, g_sCurrentMission);
	strcopy(g_CurrentCampaign.title, sizeof g_CurrentCampaign.title, g_sCurrentTitle);
	strcopy(g_CurrentCampaign.firstMap, sizeof g_CurrentCampaign.firstMap, g_sCurrentFirstMap);
	GetCurrentMap(g_sTrackedMap, sizeof g_sTrackedMap);
	NormalizeMapName(g_sTrackedMap, sizeof g_sTrackedMap);
	g_bHasCurrentCampaign = true;
	g_bCurrentCompleted = false;
}

void ClearTrackedCampaign()
{
	g_bHasCurrentCampaign = false;
	g_bCurrentCompleted = false;
	g_CurrentCampaign.mission[0] = '\0';
	g_CurrentCampaign.title[0] = '\0';
	g_CurrentCampaign.firstMap[0] = '\0';
	g_sTrackedMap[0] = '\0';
}

void HandleMapOutsideTrackedCampaign()
{
	if (!g_bHasCurrentCampaign)
		return;

	if (g_bCurrentCompleted)
	{
		PrintToServer("[轮换] 已确认完成 %s。", g_CurrentCampaign.title);
		ClearTrackedCampaign();
		SaveQueue();
		return;
	}

	if (g_bPaused)
		return;

	strcopy(g_sResumeMission, sizeof g_sResumeMission, g_CurrentCampaign.mission);
	strcopy(g_sResumeTitle, sizeof g_sResumeTitle, g_CurrentCampaign.title);
	strcopy(g_sResumeMap, sizeof g_sResumeMap, g_sTrackedMap);
	g_bPaused = true;
	SaveQueue();
	ClearMapChangerNextMap();
	PrintToServer("[轮换] 当前地图离开了未完成战役 %s，已暂停并保留恢复小关 %s。", g_CurrentCampaign.title, g_sTrackedMap);
}

void Event_CurrentCampaignCompleted(Event event, const char[] name, bool dontBroadcast)
{
	if (!g_bHasCurrentCampaign || g_bCurrentCompleted)
		return;

	char currentMap[64];
	GetCurrentMap(currentMap, sizeof currentMap);
	NormalizeMapName(currentMap, sizeof currentMap);
	if (strcmp(currentMap, g_sTrackedMap, false) != 0)
		return;

	g_bCurrentCompleted = true;
	SaveQueue();
	PrintToServer("[轮换] 已记录 %s 终局完成，等待进入下一张战役。", g_CurrentCampaign.title);
}

bool ApplyQueueNextMap()
{
	if (!g_bActive || g_bPaused || g_Queue.Length == 0)
		return false;
	if (!g_bMapChanger)
	{
		LogError("map_changer library is unavailable; cannot set queued next map");
		return false;
	}

	bool changed;
	CampaignEntry next;
	while (g_Queue.Length > 0)
	{
		g_Queue.GetArray(0, next);
		if (IsMapValidEx(next.firstMap) && MC_SetNextMap(next.firstMap))
		{
			if (changed)
				SaveQueue();
			return true;
		}

		LogError("removing unavailable queued campaign %s (%s)", next.title, next.firstMap);
		g_Queue.Erase(0);
		changed = true;
	}

	g_bActive = false;
	SaveQueue();
	return false;
}

void BuildCustomCampaignList(ArrayList output, bool excludeCurrent)
{
	output.Clear();
	SourceKeyValues missions = SDKCall(g_hSDKGetAllMissions, g_pMatchExtL4D);
	for (SourceKeyValues mission = missions.GetFirstTrueSubKey(); !mission.IsNull(); mission = mission.GetNextTrueSubKey())
	{
		if (mission.GetInt("builtin") != 0)
			continue;

		CampaignEntry entry;
		mission.GetName(entry.mission, sizeof entry.mission);
		if (excludeCurrent && strcmp(entry.mission, g_sCurrentMission, false) == 0)
			continue;
		if (FindCampaignByMission(output, entry.mission) != -1)
			continue;

		SourceKeyValues mode = FindMissionMode(mission);
		if (mode.IsNull())
			continue;
		SourceKeyValues chapter = mode.GetFirstTrueSubKey();
		if (chapter.IsNull())
			continue;
		chapter.GetString("Map", entry.firstMap, sizeof entry.firstMap, "");
		NormalizeMapName(entry.firstMap, sizeof entry.firstMap);
		if (!IsMapValidEx(entry.firstMap))
			continue;

		mission.GetString("DisplayTitle", entry.title, sizeof entry.title, entry.mission);
		output.PushArray(entry);
	}
	SortCampaignList(output);
}

SourceKeyValues FindMissionMode(SourceKeyValues mission)
{
	char path[64];
	Format(path, sizeof path, "modes/%s", g_sGameMode);
	SourceKeyValues mode = mission.FindKey(path);
	if (mode.IsNull() && strcmp(g_sGameMode, "coop", false) != 0)
		mode = mission.FindKey("modes/coop");
	return mode;
}

void SortCampaignList(ArrayList list)
{
	CampaignEntry left;
	CampaignEntry right;
	for (int i = 1; i < list.Length; i++)
	{
		for (int j = i; j > 0; j--)
		{
			list.GetArray(j - 1, left);
			list.GetArray(j, right);
			if (strcmp(left.title, right.title, false) <= 0)
				break;
			list.SwapAt(j - 1, j);
		}
	}
}

int FindCampaignByMission(ArrayList list, const char[] mission)
{
	CampaignEntry entry;
	for (int i = 0; i < list.Length; i++)
	{
		list.GetArray(i, entry);
		if (strcmp(entry.mission, mission, false) == 0)
			return i;
	}
	return -1;
}

void CopyCampaignList(ArrayList source, ArrayList target)
{
	target.Clear();
	CampaignEntry entry;
	for (int i = 0; i < source.Length; i++)
	{
		source.GetArray(i, entry);
		target.PushArray(entry);
	}
}

int ParseCampaignIndex(const char[] info)
{
	char parts[2][16];
	if (ExplodeString(info, ":", parts, sizeof parts, sizeof parts[]) != 2 || strcmp(parts[0], "campaign") != 0)
		return -1;
	return StringToInt(parts[1]);
}

bool HasHumanPlayers()
{
	for (int client = 1; client <= MaxClients; client++)
	{
		if (IsClientConnected(client) && !IsFakeClient(client))
			return true;
	}
	return false;
}

void CaptureCurrentResumePoint()
{
	RefreshCurrentCampaign();
	if (g_bHasCurrentCampaign && g_bCurrentCustom
		&& strcmp(g_CurrentCampaign.mission, g_sCurrentMission, false) == 0)
	{
		char currentMap[64];
		GetCurrentMap(currentMap, sizeof currentMap);
		NormalizeMapName(currentMap, sizeof currentMap);
		if (strcmp(g_sTrackedMap, currentMap, false) != 0)
		{
			strcopy(g_sTrackedMap, sizeof g_sTrackedMap, currentMap);
			g_bCurrentCompleted = false;
		}
	}

	if (g_bHasCurrentCampaign && g_sTrackedMap[0] != '\0')
	{
		strcopy(g_sResumeMission, sizeof g_sResumeMission, g_CurrentCampaign.mission);
		strcopy(g_sResumeTitle, sizeof g_sResumeTitle, g_CurrentCampaign.title);
		strcopy(g_sResumeMap, sizeof g_sResumeMap, g_sTrackedMap);
		return;
	}

	if (!g_bCurrentCustom || g_sCurrentMission[0] == '\0')
		return;

	strcopy(g_sResumeMission, sizeof g_sResumeMission, g_sCurrentMission);
	strcopy(g_sResumeTitle, sizeof g_sResumeTitle, g_sCurrentTitle);
	GetCurrentMap(g_sResumeMap, sizeof g_sResumeMap);
	NormalizeMapName(g_sResumeMap, sizeof g_sResumeMap);
}

void ConsumeResumePoint()
{
	if (g_bPaused || g_sResumeMap[0] == '\0')
		return;

	char currentMap[64];
	GetCurrentMap(currentMap, sizeof currentMap);
	NormalizeMapName(currentMap, sizeof currentMap);
	if (strcmp(currentMap, g_sResumeMap, false) == 0)
	{
		ClearResumePoint();
		SaveQueue();
	}
}

void ClearResumePoint()
{
	g_sResumeMission[0] = '\0';
	g_sResumeTitle[0] = '\0';
	g_sResumeMap[0] = '\0';
}

void ClearMapChangerNextMap()
{
	if (!g_bMapChanger || GetFeatureStatus(FeatureType_Native, "MC_ClearNextMap") != FeatureStatus_Available)
		return;
	MC_ClearNextMap();
}

bool IsAuthorizedAdmin(int client)
{
	if (client <= 0 || client > MaxClients || !IsClientInGame(client))
		return false;
	if (!CheckCommandAccess(client, "sm_clist", ADMFLAG_CHANGEMAP, true))
	{
		PrintToChat(client, "\x04[轮换]\x01 只有管理员可以使用此功能。");
		return false;
	}
	return true;
}

bool OwnsEditor(int client)
{
	return client > 0 && client == g_iEditor && IsClientInGame(client);
}

void ReleaseEditor(int client)
{
	if (client == g_iEditor)
	{
		g_iEditor = 0;
		g_bDraftRegenerated = false;
		g_Draft.Clear();
	}
}

void RestartAdvertTimer()
{
	if (g_hAdvertTimer != null)
	{
		delete g_hAdvertTimer;
		g_hAdvertTimer = null;
	}

	float interval = g_cvAdvertInterval.FloatValue;
	if (interval > 0.0)
		g_hAdvertTimer = CreateTimer(interval, Timer_Advertise, _, TIMER_REPEAT);
}

public Action Timer_Advertise(Handle timer)
{
	if (timer != g_hAdvertTimer)
		return Plugin_Stop;
	if (g_bActive && !g_bPaused && g_Queue.Length > 0)
		AnnounceNextCampaign("下一张三方战役");
	return Plugin_Continue;
}

void AnnounceNextCampaign(const char[] prefix)
{
	if (g_Queue.Length == 0)
		return;
	CampaignEntry next;
	g_Queue.GetArray(0, next);
	PrintToChatAll("\x04[轮换]\x01 %s：\x05%s\x01。管理员输入 !clist 可查看或调整。", prefix, next.title);
}

void LoadQueue()
{
	g_Queue.Clear();
	ClearTrackedCampaign();
	char path[PLATFORM_MAX_PATH];
	BuildPath(Path_SM, path, sizeof path, DATA_FILE);
	if (!FileExists(path))
		return;

	KeyValues kv = new KeyValues("CustomCampaignList");
	if (!kv.ImportFromFile(path))
	{
		delete kv;
		LogError("failed to read campaign list: %s", path);
		return;
	}

	g_bHasList = kv.GetNum("has_list", 0) != 0;
	g_bActive = kv.GetNum("active", 0) != 0;
	g_bPaused = kv.GetNum("paused", 0) != 0;
	g_bLegacyState = kv.GetNum("state_version", 1) < 2;
	kv.GetString("resume_mission", g_sResumeMission, sizeof g_sResumeMission, "");
	kv.GetString("resume_title", g_sResumeTitle, sizeof g_sResumeTitle, g_sResumeMission);
	kv.GetString("resume_map", g_sResumeMap, sizeof g_sResumeMap, "");
	if (kv.JumpToKey("Current"))
	{
		kv.GetString("mission", g_CurrentCampaign.mission, sizeof g_CurrentCampaign.mission, "");
		kv.GetString("title", g_CurrentCampaign.title, sizeof g_CurrentCampaign.title, g_CurrentCampaign.mission);
		kv.GetString("first_map", g_CurrentCampaign.firstMap, sizeof g_CurrentCampaign.firstMap, "");
		kv.GetString("map", g_sTrackedMap, sizeof g_sTrackedMap, "");
		g_bCurrentCompleted = kv.GetNum("completed", 0) != 0;
		g_bHasCurrentCampaign = g_CurrentCampaign.mission[0] != '\0' && g_sTrackedMap[0] != '\0';
		kv.GoBack();
		if (!g_bHasCurrentCampaign)
			ClearTrackedCampaign();
	}
	if (kv.JumpToKey("Pending") && kv.GotoFirstSubKey())
	{
		do
		{
			CampaignEntry entry;
			kv.GetString("mission", entry.mission, sizeof entry.mission, "");
			kv.GetString("title", entry.title, sizeof entry.title, entry.mission);
			kv.GetString("first_map", entry.firstMap, sizeof entry.firstMap, "");
			if (entry.mission[0] != '\0' && entry.firstMap[0] != '\0')
				g_Queue.PushArray(entry);
		}
		while (kv.GotoNextKey());
	}
	delete kv;
}

void SaveQueue()
{
	char path[PLATFORM_MAX_PATH];
	char tempPath[PLATFORM_MAX_PATH];
	BuildPath(Path_SM, path, sizeof path, DATA_FILE);
	Format(tempPath, sizeof tempPath, "%s.tmp", path);
	DeleteFile(tempPath);

	KeyValues kv = new KeyValues("CustomCampaignList");
	kv.SetNum("state_version", 2);
	kv.SetNum("has_list", g_bHasList ? 1 : 0);
	kv.SetNum("active", g_bActive ? 1 : 0);
	kv.SetNum("paused", g_bPaused ? 1 : 0);
	kv.SetString("resume_mission", g_sResumeMission);
	kv.SetString("resume_title", g_sResumeTitle);
	kv.SetString("resume_map", g_sResumeMap);
	if (g_bHasCurrentCampaign)
	{
		kv.JumpToKey("Current", true);
		kv.SetString("mission", g_CurrentCampaign.mission);
		kv.SetString("title", g_CurrentCampaign.title);
		kv.SetString("first_map", g_CurrentCampaign.firstMap);
		kv.SetString("map", g_sTrackedMap);
		kv.SetNum("completed", g_bCurrentCompleted ? 1 : 0);
		kv.GoBack();
	}
	kv.JumpToKey("Pending", true);

	CampaignEntry entry;
	char key[16];
	for (int i = 0; i < g_Queue.Length; i++)
	{
		g_Queue.GetArray(i, entry);
		IntToString(i, key, sizeof key);
		kv.JumpToKey(key, true);
		kv.SetString("mission", entry.mission);
		kv.SetString("title", entry.title);
		kv.SetString("first_map", entry.firstMap);
		kv.GoBack();
	}
	kv.Rewind();

	if (!kv.ExportToFile(tempPath) || !RenameFile(path, tempPath))
		LogError("failed to save campaign list: %s", path);
	else
		g_bLegacyState = false;
	delete kv;
}

void InitMissionSDK()
{
	GameData data = new GameData("l4d2_custom_campaign_list");
	if (data == null)
		SetFailState("Failed to load l4d2_custom_campaign_list gamedata");

	g_pMatchExtL4D = data.GetAddress("g_pMatchExtL4D");
	if (g_pMatchExtL4D == Address_Null)
		SetFailState("Failed to find g_pMatchExtL4D address");

	StartPrepSDKCall(SDKCall_Raw);
	PrepSDKCall_SetVirtual(0);
	PrepSDKCall_SetReturnInfo(SDKType_PlainOldData, SDKPass_Plain);
	g_hSDKGetAllMissions = EndPrepSDKCall();
	if (g_hSDKGetAllMissions == null)
		SetFailState("Failed to prepare MatchExtL4D::GetAllMissions SDKCall");
	delete data;
}

void NormalizeMapName(char[] map, int maxLength)
{
	TrimString(map);
	ReplaceString(map, maxLength, "\\", "/");
	if (strncmp(map, "maps/", 5, false) == 0)
		strcopy(map, maxLength, map[5]);
	int length = strlen(map);
	if (length > 4 && strcmp(map[length - 4], ".bsp", false) == 0)
		map[length - 4] = '\0';
}

bool IsMapValidEx(const char[] map)
{
	if (map[0] == '\0')
		return false;
	char found[PLATFORM_MAX_PATH];
	return FindMap(map, found, sizeof found) == FindMap_Found;
}
