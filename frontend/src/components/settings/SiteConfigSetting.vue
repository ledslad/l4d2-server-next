<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { message } from 'ant-design-vue';
  import { useSiteStore } from '../../stores/site';

  const site = useSiteStore();
  const icpNumber = ref('');
  const updateRepository = ref('');
  const loading = ref(true);
  const saving = ref(false);
  const loadError = ref('');
  onMounted(async () => {
    try {
      const config = await site.load();
      icpNumber.value = config.icp_number;
      updateRepository.value = config.update_repository;
    } catch (error) { loadError.value = String(error); }
    finally { loading.value = false; }
  });
  async function save() {
    saving.value = true;
    try {
      const result = await site.save({ icp_number: icpNumber.value, update_repository: updateRepository.value });
      icpNumber.value = result.icp_number;
      updateRepository.value = result.update_repository;
      message.success('站点配置已保存；环境变量覆盖项以实际生效值为准');
    } catch (error) { message.error(String(error)); }
    finally { saving.value = false; }
  }
</script>

<template>
  <section class="setting-section">
    <h3 class="setting-title">站点信息与更新来源</h3>
    <p class="setting-description">备案号为空时隐藏备案信息；检查更新仅查询所填 GitHub 仓库的 Releases。</p>
    <a-alert v-if="loadError" :message="loadError" type="error" class="my-3" />
    <a-form layout="vertical" class="mt-3" @finish="save">
      <a-form-item label="ICP备案号"><a-input v-model:value="icpNumber" :maxlength="64" placeholder="留空则不显示" :disabled="loading || !!loadError" /></a-form-item>
      <a-form-item label="GitHub 更新仓库"><a-input v-model:value="updateRepository" placeholder="用户名/仓库名" :disabled="loading || !!loadError" /></a-form-item>
      <a-button type="primary" html-type="submit" :loading="saving" :disabled="loading || !!loadError">保存站点配置</a-button>
    </a-form>
  </section>
</template>
