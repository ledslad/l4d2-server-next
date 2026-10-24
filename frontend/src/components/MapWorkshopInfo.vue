<script setup lang="ts">
  import { computed } from 'vue';
  import { LinkOutlined } from '@ant-design/icons-vue';
  import type { MapWorkshopMeta } from '../services/api';
  import { getWorkshopPreviewURL, getWorkshopSteamURL } from '../utils/mapWorkshop';

  const props = withDefaults(defineProps<{
    name: string;
    title?: string;
    workshop?: MapWorkshopMeta | null;
    previewFailed?: boolean;
    loading?: boolean;
    disabled?: boolean;
    detailed?: boolean;
    size?: string;
    chapterCount?: number;
  }>(), { title: '', chapterCount: 0 });
  const emit = defineEmits<{ details: []; imageError: [url: string] }>();
  const displayTitle = computed(() => props.workshop?.title?.trim() || props.title || props.name);
  const previewURL = computed(() => getWorkshopPreviewURL(props.workshop));
  const steamURL = computed(() => getWorkshopSteamURL(props.workshop));
  // Render cleaned Workshop descriptions as text without interpreting HTML.
  const description = computed(() => String(props.workshop?.description || '')
    .replace(/\[url=([^\]]+)]([^[]+)\[\/url]/gi, '$2 ($1)')
    .replace(/\[img][\s\S]*?\[\/img]/gi, '')
    .replace(/\[\*]/g, '• ')
    .replace(/\[\/?.+?]/g, '')
    .replace(/\r/g, '')
    .replace(/\n{3,}/g, '\n\n')
    .trim());
  const formatDate = (value?: string) => {
    if (!value) return '-';
    const date = /^\d+$/.test(value) ? new Date(Number(value) * 1000) : new Date(value);
    if (Number.isNaN(date.valueOf())) return '-';
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric', month: '2-digit', day: '2-digit',
    }).format(date);
  };
</script>

<template>
  <button v-if="!detailed" class="map-main" type="button" :disabled="disabled"
    :aria-label="`查看 ${displayTitle} 详情`" @click="emit('details')">
    <span class="map-preview">
      <img v-if="previewURL && !previewFailed" :src="previewURL" alt="" loading="lazy"
        @error="emit('imageError', previewURL)">
      <span v-else>VPK</span>
    </span>
    <span class="map-copy">
      <strong :title="displayTitle">{{ displayTitle }}</strong>
      <span v-if="loading">读取中...</span>
      <template v-else>
        <span v-if="workshop?.author" :title="`作者：${workshop.author}`">作者：{{ workshop.author }}</span>
        <span v-if="name !== displayTitle" :title="name">{{ name }}</span>
      </template>
    </span>
  </button>
  <div v-else class="workshop-detail">
    <section class="detail-summary">
      <div class="detail-preview">
        <img v-if="previewURL && !previewFailed" :src="previewURL" :alt="`${displayTitle} 预览图`"
          @error="emit('imageError', previewURL)">
        <span v-else>VPK</span>
      </div>
      <div class="detail-overview">
        <h2>{{ displayTitle }}</h2>
        <p v-if="workshop?.author">作者：{{ workshop.author }}</p>
        <div class="detail-stats">
          <div><span>文件大小</span><strong>{{ size || '-' }}</strong></div>
          <div><span>章节数</span><strong>{{ chapterCount || '-' }}</strong></div>
        </div>
        <a-button v-if="steamURL" :href="steamURL" target="_blank" rel="noopener noreferrer">
          <template #icon><link-outlined /></template>Steam 创意工坊
        </a-button>
      </div>
    </section>
    <section v-if="description" class="description-section">
      <h3>工坊说明</h3><div>{{ description }}</div>
    </section>
    <a-descriptions :column="1" size="small" class="detail-metadata">
      <a-descriptions-item label="VPK 文件">{{ name }}</a-descriptions-item>
      <template v-if="workshop">
        <a-descriptions-item label="Workshop ID">{{ workshop.workshop_id }}</a-descriptions-item>
        <a-descriptions-item label="下载时间">{{ formatDate(workshop.downloaded_at) }}</a-descriptions-item>
        <a-descriptions-item label="工坊更新时间">{{ formatDate(workshop.time_updated) }}</a-descriptions-item>
      </template>
    </a-descriptions>
  </div>
</template>

<style scoped>
  .map-main {
    min-width: 0; width: 100%; padding: 0; border: 0;
    display: flex; align-items: center; gap: 12px;
    background: transparent; color: inherit; text-align: left; cursor: pointer;
  }
  .map-main:disabled { cursor: default; }
  .map-main:focus-visible { outline: 2px solid #1677ff; outline-offset: 4px; border-radius: 5px; }
  .map-preview {
    width: 64px; height: 42px; flex: 0 0 64px; overflow: hidden;
    border: 1px solid rgba(148, 163, 184, .28); border-radius: 5px;
    display: grid; place-items: center; background: rgba(148, 163, 184, .1);
    color: #d97706; font-size: 10px; font-weight: 700;
  }
  .map-preview img { width: 100%; height: 100%; object-fit: cover; }
  .map-copy { min-width: 0; flex: 1 1 auto; overflow: hidden; }
  .map-copy strong, .map-copy > span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .map-copy strong { margin-bottom: 4px; font-size: 14px; }
  .map-copy > span { color: var(--ant-color-text-secondary, #6b7280); font-size: 12px; }
  .detail-summary { display: grid; grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr); gap: 20px; }
  .detail-preview {
    width: 100%; aspect-ratio: 16 / 9; overflow: hidden;
    border: 1px solid rgba(148, 163, 184, .28); border-radius: 7px;
    display: grid; place-items: center; background: rgba(148, 163, 184, .1);
    color: #d97706; font-size: 18px; font-weight: 700;
  }
  .detail-preview img { width: 100%; height: 100%; object-fit: contain; }
  .detail-overview { min-width: 0; display: flex; flex-direction: column; align-items: flex-start; gap: 12px; }
  .detail-overview h2 { margin: 0; font-size: 22px; line-height: 1.4; overflow-wrap: anywhere; }
  .detail-overview p { margin: 0; color: var(--ant-color-text-secondary, #6b7280); overflow-wrap: anywhere; }
  .detail-stats { width: 100%; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
  .detail-stats span, .detail-stats strong { display: block; }
  .detail-stats span { margin-bottom: 4px; color: var(--ant-color-text-secondary, #6b7280); font-size: 12px; }
  .description-section { margin-top: 18px; padding-top: 16px; border-top: 1px solid rgba(148, 163, 184, .24); }
  .description-section h3 { margin: 0 0 10px; font-size: 14px; }
  .description-section div { max-height: 210px; overflow: auto; white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.7; }
  .detail-metadata { margin-top: 18px; overflow-wrap: anywhere; }
  @media (max-width: 640px) {
    .detail-summary { grid-template-columns: 1fr; }
    .map-main { gap: 8px; }
    .map-preview { width: 56px; height: 38px; flex-basis: 56px; }
  }
</style>
