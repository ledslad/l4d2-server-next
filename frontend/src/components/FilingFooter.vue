<script setup lang="ts">
  import { onMounted } from 'vue';
  import { useSiteStore } from '../stores/site';

  const site = useSiteStore();
  onMounted(() => { site.load().catch(() => { /* 未配置时不显示备案信息。 */ }); });
</script>

<template>
  <footer class="w-full py-3 text-center text-xs text-gray-400 dark:text-gray-500">
    <a href="https://github.com/ledslad/l4d2-server-next" target="_blank" rel="noopener noreferrer" class="transition-colors hover:text-blue-500 dark:hover:text-blue-400">本分支源代码 · AGPL v3</a>
    <span v-if="site.config.icp_number" class="mx-2">·</span>
    <a
      v-if="site.config.icp_number"
      href="https://beian.miit.gov.cn/"
      target="_blank"
      rel="noopener noreferrer"
      class="transition-colors hover:text-blue-500 dark:hover:text-blue-400"
    >
      {{ site.config.icp_number }}
    </a>
  </footer>
</template>
