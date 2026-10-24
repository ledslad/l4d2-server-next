<script setup lang="ts">
  import { theme } from 'ant-design-vue';
  import { useThemeStore } from './stores/theme';
  import { ConfigProvider } from 'ant-design-vue';

  const themeStore = useThemeStore();

  /**
   * 解决 Ant Design Vue 组件（如 Tooltip, Select 等）的浮层在页面切换时
   * 可能残留或位置错误导致页面出现滚动条的问题。
   * 将挂载节点设置为触发节点的父元素，使其随组件销毁而销毁。
   */
  const getPopupContainer = (el?: HTMLElement) => {
    if (el && el.parentNode) {
      return el.parentNode as HTMLElement;
    }
    return document.body;
  };
</script>

<template>
  <ConfigProvider
    :getPopupContainer="getPopupContainer"
    :theme="{
      algorithm: themeStore.isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
      token: {
        colorBgLayout: themeStore.isDark ? '#020617' : '#f9fafb', // gray-950 : gray-50
        colorBgContainer: themeStore.isDark ? '#0f172a' : '#ffffff', // gray-900 : white
        colorBgElevated: themeStore.isDark ? '#1e293b' : '#ffffff', // gray-800 : white
      },
      components: {
        Layout: {
          colorBgHeader: themeStore.isDark ? '#0f172a' : '#ffffff', // gray-900 : white
          colorBgBody: themeStore.isDark ? '#020617' : '#f9fafb', // gray-950 : gray-50
          colorBgTrigger: themeStore.isDark ? '#1e293b' : '#ffffff', // gray-800 : white
        },
        Menu: {
          colorItemBg: themeStore.isDark ? '#0f172a' : '#ffffff',
          colorItemBgSelected: themeStore.isDark ? 'rgba(30, 58, 138, 0.3)' : '#e6f7ff', // blue-900/30
        },
      },
    }"
  >
    <router-view />
  </ConfigProvider>
</template>
