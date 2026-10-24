import type { MapWorkshopMeta } from '../services/api';

export function getWorkshopPreviewURL(workshop?: MapWorkshopMeta | null): string {
  try {
    const url = new URL(workshop?.preview_url || '');
    return ['https:', 'http:'].includes(url.protocol) ? url.href : '';
  } catch {
    return '';
  }
}

export function getWorkshopSteamURL(workshop?: MapWorkshopMeta | null): string {
  return /^\d+$/.test(workshop?.workshop_id || '')
    ? `https://steamcommunity.com/sharedfiles/filedetails/?id=${workshop!.workshop_id}`
    : '';
}
