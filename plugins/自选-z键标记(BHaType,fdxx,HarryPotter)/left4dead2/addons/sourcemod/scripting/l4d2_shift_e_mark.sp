#pragma semicolon 1
#pragma newdecls required

#include <sourcemod>
#include <sdktools>

bool g_wasMarkComboDown[MAXPLAYERS + 1];

public Plugin myinfo =
{
	name = "L4D2 Shift+E Marker",
	author = "Codex",
	description = "Triggers l4d2_item_hint's sm_mark when a survivor presses Shift+E.",
	version = "1.0.1"
};

public void OnClientDisconnect(int client)
{
	g_wasMarkComboDown[client] = false;
}

public Action OnPlayerRunCmd(int client, int &buttons, int &impulse, float vel[3], float angles[3], int &weapon)
{
	if (!IsValidSurvivor(client))
	{
		if (client > 0 && client <= MaxClients)
		{
			g_wasMarkComboDown[client] = false;
		}
		return Plugin_Continue;
	}

	bool comboDown = (buttons & IN_SPEED) != 0 && (buttons & IN_USE) != 0;
	if (comboDown && !g_wasMarkComboDown[client])
	{
		CreateTimer(0.0, Timer_RunMarkCommand, GetClientUserId(client), TIMER_FLAG_NO_MAPCHANGE);
	}

	g_wasMarkComboDown[client] = comboDown;
	if (comboDown)
	{
		buttons &= ~IN_USE;
		return Plugin_Changed;
	}

	return Plugin_Continue;
}

public Action Timer_RunMarkCommand(Handle timer, any userid)
{
	int client = GetClientOfUserId(userid);
	if (IsValidSurvivor(client))
	{
		FakeClientCommand(client, "sm_mark");
	}

	return Plugin_Stop;
}

bool IsValidSurvivor(int client)
{
	return client > 0
		&& client <= MaxClients
		&& IsClientInGame(client)
		&& !IsFakeClient(client)
		&& GetClientTeam(client) == 2
		&& IsPlayerAlive(client);
}
