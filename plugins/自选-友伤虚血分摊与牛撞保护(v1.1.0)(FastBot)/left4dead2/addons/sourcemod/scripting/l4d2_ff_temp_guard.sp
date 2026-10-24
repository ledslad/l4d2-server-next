#pragma semicolon 1
#pragma newdecls required

#include <sourcemod>
#include <sdkhooks>

#define PLUGIN_VERSION "1.1.0"
#define PENDING_SLOTS 16

ConVar g_cvVersion;
ConVar g_cvEnabled;
ConVar g_cvPermFraction;
ConVar g_cvChargerImmunity;
ConVar g_cvPillsDecay;

bool g_bPending[MAXPLAYERS + 1][PENDING_SLOTS];
int g_iPendingPermHealth[MAXPLAYERS + 1][PENDING_SLOTS];
int g_iPendingTempHealth[MAXPLAYERS + 1][PENDING_SLOTS];
int g_iPendingIndex[MAXPLAYERS + 1];
bool g_bCarryProtected[MAXPLAYERS + 1];
int g_iCarryGeneration[MAXPLAYERS + 1];

public Plugin myinfo =
{
    name = "L4D2 Friendly Fire Temp Health Guard",
    author = "FastBot",
    description = "Splits survivor damage between permanent and temporary health and blocks FF during Charger carries",
    version = PLUGIN_VERSION,
    url = ""
};

public void OnPluginStart()
{
    g_cvVersion = CreateConVar("l4d2_ff_temp_guard_version", PLUGIN_VERSION,
        "Friendly fire temporary health guard version.", FCVAR_NOTIFY | FCVAR_DONTRECORD);
    g_cvVersion.SetString(PLUGIN_VERSION);
    g_cvEnabled = CreateConVar("l4d2_ff_temp_guard_enable", "1",
        "Enable survivor damage splitting and Charger carry protection.", _, true, 0.0, true, 1.0);
    g_cvPermFraction = CreateConVar("l4d2_ff_temp_guard_perm_fraction", "0.5",
        "Fraction of survivor damage applied to permanent health.", _, true, 0.0, true, 1.0);
    g_cvChargerImmunity = CreateConVar("l4d2_ff_temp_guard_charger_immunity", "1",
        "Block survivor friendly fire while the victim is carried by a Charger.", _, true, 0.0, true, 1.0);
    AutoExecConfig(true, "l4d2_ff_temp_guard");

    g_cvPillsDecay = FindConVar("pain_pills_decay_rate");
    HookEvent("player_hurt", Event_PlayerHurt, EventHookMode_Post);
    HookEvent("player_incapacitated_start", Event_PlayerIncapacitatedStart, EventHookMode_Pre);
    HookEvent("charger_carry_start", Event_ChargerCarryStart, EventHookMode_Post);
    HookEvent("charger_carry_end", Event_ChargerCarryEnd, EventHookMode_Post);

    for (int client = 1; client <= MaxClients; client++)
    {
        if (IsClientInGame(client))
        {
            HookClient(client);
        }
    }
}

public void OnClientPutInServer(int client)
{
    HookClient(client);
}

public void OnClientDisconnect(int client)
{
    ResetClientState(client);
}

void HookClient(int client)
{
    ResetClientState(client);
    SDKHook(client, SDKHook_OnTakeDamage, OnTakeDamage);
}

void ResetClientState(int client)
{
    if (client <= 0 || client > MaxClients)
    {
        return;
    }

    g_iPendingIndex[client] = 0;
    g_bCarryProtected[client] = false;
    g_iCarryGeneration[client]++;
    for (int slot = 0; slot < PENDING_SLOTS; slot++)
    {
        g_bPending[client][slot] = false;
        g_iPendingPermHealth[client][slot] = 0;
        g_iPendingTempHealth[client][slot] = 0;
    }
}

public Action OnTakeDamage(int victim, int &attacker, int &inflictor, float &damage, int &damageType)
{
    if (!g_cvEnabled.BoolValue || !IsStandingSurvivor(victim))
    {
        return Plugin_Continue;
    }

    if (g_cvChargerImmunity.BoolValue && IsSurvivor(attacker) && attacker != victim && IsCarryProtected(victim))
    {
        damage = 0.0;
        return Plugin_Handled;
    }

    int totalDamage = RoundToFloor(damage);
    if (totalDamage <= 0)
    {
        return Plugin_Continue;
    }

    int permanentHealth = GetClientHealth(victim);
    int temporaryHealth = GetTemporaryHealth(victim);
    if (temporaryHealth <= 0 || totalDamage >= permanentHealth + temporaryHealth)
    {
        return Plugin_Continue;
    }

    int permanentDamage = RoundToCeil(float(totalDamage) * g_cvPermFraction.FloatValue);
    permanentDamage = ClampInt(permanentDamage, 0, totalDamage);
    int temporaryDamage = totalDamage - permanentDamage;

    if (temporaryDamage > temporaryHealth)
    {
        permanentDamage += temporaryDamage - temporaryHealth;
        temporaryDamage = temporaryHealth;
    }

    if (permanentDamage >= permanentHealth && totalDamage < permanentHealth + temporaryHealth)
    {
        int shiftToTemporary = permanentDamage - (permanentHealth - 1);
        int spareTemporary = temporaryHealth - temporaryDamage;
        if (shiftToTemporary > spareTemporary)
        {
            shiftToTemporary = spareTemporary;
        }
        permanentDamage -= shiftToTemporary;
        temporaryDamage += shiftToTemporary;
    }

    if (temporaryDamage <= 0)
    {
        return Plugin_Continue;
    }

    int slot = (g_iPendingIndex[victim] + 1) % PENDING_SLOTS;
    g_bPending[victim][slot] = true;
    g_iPendingPermHealth[victim][slot] = permanentHealth - permanentDamage;
    g_iPendingTempHealth[victim][slot] = temporaryHealth - temporaryDamage;

    return Plugin_Continue;
}

public void Event_PlayerHurt(Event event, const char[] name, bool dontBroadcast)
{
    int victim = GetClientOfUserId(event.GetInt("userid"));
    if (!IsSurvivor(victim))
    {
        return;
    }

    int slot = (g_iPendingIndex[victim] + 1) % PENDING_SLOTS;
    g_iPendingIndex[victim] = slot;
    if (!g_bPending[victim][slot])
    {
        return;
    }

    g_bPending[victim][slot] = false;
    if (IsIncapacitated(victim))
    {
        return;
    }

    int targetPermanentHealth = g_iPendingPermHealth[victim][slot];
    int targetTemporaryHealth = g_iPendingTempHealth[victim][slot];
    int permanentHealth = GetClientHealth(victim);
    int temporaryHealth = GetTemporaryHealth(victim);
    if (permanentHealth + temporaryHealth != targetPermanentHealth + targetTemporaryHealth)
    {
        return;
    }

    if (permanentHealth != targetPermanentHealth)
    {
        SetEntityHealth(victim, targetPermanentHealth);
    }
    if (temporaryHealth != targetTemporaryHealth)
    {
        SetTemporaryHealth(victim, targetTemporaryHealth);
    }
}

public void Event_PlayerIncapacitatedStart(Event event, const char[] name, bool dontBroadcast)
{
    int victim = GetClientOfUserId(event.GetInt("userid"));
    if (!IsSurvivor(victim))
    {
        return;
    }

    int slot = (g_iPendingIndex[victim] + 1) % PENDING_SLOTS;
    g_iPendingIndex[victim] = slot;
    g_bPending[victim][slot] = false;
}

public void Event_ChargerCarryStart(Event event, const char[] name, bool dontBroadcast)
{
    int victim = GetClientOfUserId(event.GetInt("victim"));
    if (!IsSurvivor(victim))
    {
        return;
    }

    g_iCarryGeneration[victim]++;
    g_bCarryProtected[victim] = true;
}

public void Event_ChargerCarryEnd(Event event, const char[] name, bool dontBroadcast)
{
    int victim = GetClientOfUserId(event.GetInt("victim"));
    if (!IsSurvivor(victim))
    {
        return;
    }

    int generation = ++g_iCarryGeneration[victim];
    DataPack pack = new DataPack();
    pack.WriteCell(GetClientUserId(victim));
    pack.WriteCell(generation);
    CreateTimer(1.0, Timer_EndCarryProtection, pack, TIMER_FLAG_NO_MAPCHANGE | TIMER_DATA_HNDL_CLOSE);
}

public Action Timer_EndCarryProtection(Handle timer, DataPack pack)
{
    pack.Reset();
    int victim = GetClientOfUserId(pack.ReadCell());
    int generation = pack.ReadCell();
    if (IsSurvivor(victim) && g_iCarryGeneration[victim] == generation)
    {
        g_bCarryProtected[victim] = false;
    }
    return Plugin_Stop;
}

bool IsCarryProtected(int client)
{
    return g_bCarryProtected[client] || GetEntPropEnt(client, Prop_Send, "m_carryAttacker") > 0;
}

bool IsStandingSurvivor(int client)
{
    return IsSurvivor(client) && IsPlayerAlive(client) && !IsIncapacitated(client);
}

bool IsSurvivor(int client)
{
    return client > 0 && client <= MaxClients && IsClientInGame(client) && GetClientTeam(client) == 2;
}

bool IsIncapacitated(int client)
{
    return GetEntProp(client, Prop_Send, "m_isIncapacitated", 1) != 0 ||
        GetEntProp(client, Prop_Send, "m_isHangingFromLedge", 1) != 0;
}

int GetTemporaryHealth(int client)
{
    if (g_cvPillsDecay == null)
    {
        return 0;
    }

    float buffer = GetEntPropFloat(client, Prop_Send, "m_healthBuffer");
    float bufferTime = GetEntPropFloat(client, Prop_Send, "m_healthBufferTime");
    float remaining = buffer - ((GetGameTime() - bufferTime) * g_cvPillsDecay.FloatValue);
    int health = RoundToCeil(remaining) - 1;
    return health > 0 ? health : 0;
}

void SetTemporaryHealth(int client, int health)
{
    SetEntPropFloat(client, Prop_Send, "m_healthBuffer", float(health));
    SetEntPropFloat(client, Prop_Send, "m_healthBufferTime", GetGameTime());
}

int ClampInt(int value, int minimum, int maximum)
{
    if (value < minimum)
    {
        return minimum;
    }
    if (value > maximum)
    {
        return maximum;
    }
    return value;
}
