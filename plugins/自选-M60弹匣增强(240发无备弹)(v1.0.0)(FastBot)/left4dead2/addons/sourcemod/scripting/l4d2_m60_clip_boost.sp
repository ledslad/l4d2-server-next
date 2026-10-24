#pragma semicolon 1
#pragma newdecls required

#include <sourcemod>
#include <left4dhooks>

#define PLUGIN_VERSION "1.0.0"
#define M60_WEAPON_NAME "weapon_rifle_m60"
#define MAX_CROSS_PLATFORM_CLIP 255

ConVar g_cvEnable;
ConVar g_cvClipSize;

bool g_bConfigsExecuted;
bool g_bOriginalStored;
bool g_bApplied;
int g_iOriginalClipSize;
int g_iOriginalDefaultSize;

public APLRes AskPluginLoad2(Handle myself, bool late, char[] error, int errorMax)
{
	if (GetEngineVersion() != Engine_Left4Dead2)
	{
		strcopy(error, errorMax, "Plugin only supports Left 4 Dead 2");
		return APLRes_SilentFailure;
	}

	return APLRes_Success;
}

public Plugin myinfo =
{
	name = "[L4D2] M60 Clip Boost",
	author = "FastBot",
	description = "Sets the M60 initial clip size without adding reserve ammo",
	version = PLUGIN_VERSION,
	url = ""
};

public void OnPluginStart()
{
	CreateConVar(
		"l4d2_m60_clip_boost_version",
		PLUGIN_VERSION,
		"M60 Clip Boost version.",
		FCVAR_NOTIFY | FCVAR_DONTRECORD
	);

	g_cvEnable = CreateConVar(
		"l4d2_m60_clip_boost_enable",
		"1",
		"Enable the M60 clip boost. 0=Off, 1=On.",
		FCVAR_NOTIFY,
		true,
		0.0,
		true,
		1.0
	);

	g_cvClipSize = CreateConVar(
		"l4d2_m60_clip_boost_size",
		"240",
		"Initial M60 clip size. Limited to 255 for Windows and Linux compatibility.",
		FCVAR_NOTIFY,
		true,
		1.0,
		true,
		float(MAX_CROSS_PLATFORM_CLIP)
	);

	g_cvEnable.AddChangeHook(OnSettingChanged);
	g_cvClipSize.AddChangeHook(OnSettingChanged);
	RegServerCmd("sm_m60_clip_boost_status", CommandStatus);

	AutoExecConfig(true, "l4d2_m60_clip_boost");
}

public void OnConfigsExecuted()
{
	g_bConfigsExecuted = true;
	ApplySettings();
}

public void OnPluginEnd()
{
	RestoreOriginalAttributes();
}

public void OnSettingChanged(ConVar convar, const char[] oldValue, const char[] newValue)
{
	if (g_bConfigsExecuted)
		ApplySettings();
}

public Action CommandStatus(int args)
{
	int actualClipSize = L4D2_GetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_ClipSize);
	int actualDefaultSize = L4D2_GetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_DefaultSize);

	PrintToServer(
		"[M60 Clip Boost] enabled=%d configured=%d actual_clip=%d actual_default=%d applied=%d",
		g_cvEnable.BoolValue,
		g_cvClipSize.IntValue,
		actualClipSize,
		actualDefaultSize,
		g_bApplied
	);

	return Plugin_Handled;
}

void ApplySettings()
{
	if (!g_cvEnable.BoolValue)
	{
		RestoreOriginalAttributes();
		return;
	}

	StoreOriginalAttributes();

	int clipSize = g_cvClipSize.IntValue;
	L4D2_SetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_ClipSize, clipSize);
	L4D2_SetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_DefaultSize, clipSize);
	g_bApplied = true;

	PrintToServer("[M60 Clip Boost] Initial clip size set to %d; reserve ammo unchanged.", clipSize);
}

void StoreOriginalAttributes()
{
	if (g_bOriginalStored)
		return;

	g_iOriginalClipSize = L4D2_GetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_ClipSize);
	g_iOriginalDefaultSize = L4D2_GetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_DefaultSize);
	g_bOriginalStored = true;
}

void RestoreOriginalAttributes()
{
	if (!g_bOriginalStored || !g_bApplied)
		return;

	L4D2_SetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_ClipSize, g_iOriginalClipSize);
	L4D2_SetIntWeaponAttribute(M60_WEAPON_NAME, L4D2IWA_DefaultSize, g_iOriginalDefaultSize);
	g_bApplied = false;

	PrintToServer(
		"[M60 Clip Boost] Original clip attributes restored: clip=%d, default=%d.",
		g_iOriginalClipSize,
		g_iOriginalDefaultSize
	);
}
