import { defineStore } from 'pinia';
import { ref } from 'vue';
import { api, type SiteConfig } from '../services/api';

export const useSiteStore = defineStore('site', () => {
  const config = ref<SiteConfig>({ icp_number: '', update_repository: '' });
  let requestId = 0;

  async function load() {
    const id = ++requestId;
    const result = await api.getSiteConfig();
    if (id === requestId) config.value = result;
    return result;
  }

  async function save(value: SiteConfig) {
    ++requestId;
    const result = await api.setSiteConfig(value);
    config.value = result;
    return result;
  }

  return { config, load, save };
});
