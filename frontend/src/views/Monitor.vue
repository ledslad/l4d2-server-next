<template>
  <div class="p-4 md:p-6 space-y-6">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-center gap-4 bg-white dark:bg-gray-900 p-4 rounded-lg shadow-sm transition-colors duration-300"
    >
      <div>
        <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100 flex items-center gap-2">
          <LineChartOutlined /> 性能监控
        </h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">实时监控服务器资源使用情况</p>
      </div>

      <!-- Center: Time Range Selector -->
      <div v-if="historyEnabled && isAdmin" class="flex-1 flex justify-center">
        <a-segmented
          :value="viewMode"
          :options="[
            { label: '实时', value: 'realtime' },
            { label: '1小时', value: '1h' },
            { label: '24小时', value: '24h' },
            { label: '3天', value: '3d' },
          ]"
          :disabled="historyLoading"
          @change="(val: any) => setViewMode(val)"
        />
      </div>

      <div class="flex gap-3">
        <a-button
          v-if="viewMode === 'realtime'"
          :type="isMonitoring ? 'primary' : 'default'"
          :danger="isMonitoring"
          @click="toggleMonitor"
          :loading="loading"
          class="!flex !items-center !justify-center"
        >
          <template #icon>
            <StopOutlined v-if="isMonitoring" />
            <PlayCircleOutlined v-else />
          </template>
          <span>{{ isMonitoring ? '停止监控' : '开始监控' }}</span>
        </a-button>
        <a-button
          v-else
          @click="refreshHistory"
          :loading="historyLoading"
          class="!flex !items-center !justify-center"
        >
          <template #icon>
            <SyncOutlined />
          </template>
          <span>刷新数据</span>
        </a-button>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- CPU -->
      <div
        class="bg-white dark:bg-gray-900 p-4 rounded-lg shadow-sm transition-colors duration-300"
      >
        <div
          class="flex justify-between items-center mb-4 border-b border-gray-100 dark:border-gray-800 pb-2"
        >
          <h3 class="font-bold text-gray-700 dark:text-gray-200">CPU</h3>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-mono">
            <span v-if="cpuData.length > 0">
              Total: {{ cpuData[cpuData.length - 1] }}% | Core Max:
              {{ cpuMaxCoreData[cpuMaxCoreData.length - 1] }}%
            </span>
            <span v-else>Waiting for data...</span>
          </div>
        </div>
        <div class="relative w-full h-72">
          <div
            v-if="!isMonitoring && cpuData.length === 0"
            class="absolute inset-0 flex flex-col items-center justify-center text-gray-300 dark:text-gray-600"
          >
            <LineChartOutlined class="text-6xl mb-2 opacity-50" />
            <span class="text-sm">点击开始监控以查看数据</span>
          </div>
          <div
            ref="cpuChartRef"
            class="w-full h-full"
            :class="{ 'opacity-0': !isMonitoring && cpuData.length === 0 }"
          ></div>
        </div>
      </div>

      <!-- Memory -->
      <div
        class="bg-white dark:bg-gray-900 p-4 rounded-lg shadow-sm transition-colors duration-300"
      >
        <div
          class="flex justify-between items-center mb-4 border-b border-gray-100 dark:border-gray-800 pb-2"
        >
          <h3 class="font-bold text-gray-700 dark:text-gray-200">内存</h3>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-mono flex flex-col items-end">
            <span v-if="memUsedData.length > 0">
              Phy: {{ memUsedData[memUsedData.length - 1] }} GB /
              {{ (memTotal / 1024 / 1024 / 1024).toFixed(1) }} GB
            </span>
            <span v-if="swapUsedData.length > 0">
              Swap: {{ swapUsedData[swapUsedData.length - 1] }} GB /
              {{ (swapTotal / 1024 / 1024 / 1024).toFixed(1) }} GB
            </span>
            <span v-if="memUsedData.length === 0">Waiting for data...</span>
          </div>
        </div>
        <div class="relative w-full h-72">
          <div
            v-if="!isMonitoring && memUsedData.length === 0"
            class="absolute inset-0 flex flex-col items-center justify-center text-gray-300 dark:text-gray-600"
          >
            <AppstoreAddOutlined class="text-6xl mb-2 opacity-50" />
            <span class="text-sm">点击开始监控以查看数据</span>
          </div>
          <div
            ref="memChartRef"
            class="w-full h-full"
            :class="{ 'opacity-0': !isMonitoring && memUsedData.length === 0 }"
          ></div>
        </div>
      </div>

      <!-- Network -->
      <div
        class="bg-white dark:bg-gray-900 p-4 rounded-lg shadow-sm transition-colors duration-300"
      >
        <div
          class="flex justify-between items-center mb-4 border-b border-gray-100 dark:border-gray-800 pb-2"
        >
          <h3 class="font-bold text-gray-700 dark:text-gray-200">网络</h3>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-mono">
            <span v-if="netUpData.length > 0">
              ↑ {{ netUpData[netUpData.length - 1] }} KB/s | ↓
              {{ netDownData[netDownData.length - 1] }} KB/s
            </span>
            <span v-else>Waiting for data...</span>
          </div>
        </div>
        <div class="relative w-full h-72">
          <div
            v-if="!isMonitoring && netUpData.length === 0"
            class="absolute inset-0 flex flex-col items-center justify-center text-gray-300 dark:text-gray-600"
          >
            <GlobalOutlined class="text-6xl mb-2 opacity-50" />
            <span class="text-sm">点击开始监控以查看数据</span>
          </div>
          <div
            ref="netChartRef"
            class="w-full h-full"
            :class="{ 'opacity-0': !isMonitoring && netUpData.length === 0 }"
          ></div>
        </div>
      </div>

      <!-- Disk -->
      <div
        class="bg-white dark:bg-gray-900 p-4 rounded-lg shadow-sm transition-colors duration-300"
      >
        <div
          class="flex justify-between items-center mb-4 border-b border-gray-100 dark:border-gray-800 pb-2"
        >
          <h3 class="font-bold text-gray-700 dark:text-gray-200">硬盘</h3>
          <div class="text-xs text-gray-500 dark:text-gray-400 font-mono">
            <span v-if="diskUsedData.length > 0">
              {{ diskUsedData[diskUsedData.length - 1] }} GB
              <span v-if="diskTotal > 0">
                / {{ (diskTotal / 1024 / 1024 / 1024).toFixed(1) }} GB
              </span>
            </span>
            <span v-else>Waiting for data...</span>
          </div>
        </div>
        <div class="relative w-full h-72">
          <div
            v-if="!isMonitoring && diskUsedData.length === 0"
            class="absolute inset-0 flex flex-col items-center justify-center text-gray-300 dark:text-gray-600"
          >
            <DatabaseOutlined class="text-6xl mb-2 opacity-50" />
            <span class="text-sm">点击开始监控以查看数据</span>
          </div>
          <div
            ref="diskChartRef"
            class="w-full h-full"
            :class="{ 'opacity-0': !isMonitoring && diskUsedData.length === 0 }"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue';
  import {
    LineChartOutlined,
    PlayCircleOutlined,
    StopOutlined,
    AppstoreAddOutlined,
    GlobalOutlined,
    DatabaseOutlined,
    SyncOutlined,
  } from '@ant-design/icons-vue';

  // ECharts Core
  import * as echarts from 'echarts/core';
  import { LineChart } from 'echarts/charts';
  import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
  import { ToolboxComponent, DataZoomComponent, BrushComponent } from 'echarts/components';
  import { CanvasRenderer } from 'echarts/renderers';

  echarts.use([
    LineChart,
    GridComponent,
    TooltipComponent,
    LegendComponent,
    ToolboxComponent,
    DataZoomComponent,
    BrushComponent,
    CanvasRenderer,
  ]);

  import { useThemeStore } from '../stores/theme';
  import { useMonitorStore } from '../stores/monitor';
  import { useAuthStore } from '../stores/auth';
  import { storeToRefs } from 'pinia';

  const themeStore = useThemeStore();
  const monitorStore = useMonitorStore();
  const authStore = useAuthStore();
  const isAdmin = computed(() => authStore.isAdmin);

  const {
    isMonitoring,
    historyEnabled,
    historyLoading,
    viewMode,
    timestamps,
    hRawTimestamps,
    cpuData,
    cpuMaxCoreData,
    memUsedData,
    memTotal,
    swapUsedData,
    swapTotal,
    netUpData,
    netDownData,
    diskUsedData,
    diskTotal,
  } = storeToRefs(monitorStore);

  const { toggleMonitor, fetchConfig, setViewMode, refreshHistory, fetchCustomHistory } =
    monitorStore;

  const loading = ref(false);

  // Chart Refs
  const cpuChartRef = ref<HTMLElement | null>(null);
  const memChartRef = ref<HTMLElement | null>(null);
  const netChartRef = ref<HTMLElement | null>(null);
  const diskChartRef = ref<HTMLElement | null>(null);

  let cpuChart: echarts.ECharts | null = null;
  let memChart: echarts.ECharts | null = null;
  let netChart: echarts.ECharts | null = null;
  let diskChart: echarts.ECharts | null = null;

  const getChartThemeColor = () => {
    return themeStore.isDark ? '#e5e7eb' : '#374151';
  };

  const getSplitLineColor = () => {
    return themeStore.isDark ? '#374151' : '#e5e7eb';
  };

  const initCharts = () => {
    const commonOption = {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
          label: { show: false }, // Hide axis pointer labels
        },
      },
      toolbox: {
        show: false,
        feature: {
          brush: {
            type: ['lineX'],
          },
        },
      },
      brush: {
        xAxisIndex: 'all',
        brushLink: 'all',
        outOfBrush: {
          colorAlpha: 0.1,
        },
        brushStyle: {
          borderWidth: 1,
          color: 'rgba(120,140,180,0.3)',
          borderColor: 'rgba(120,140,180,0.8)',
        },
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '12%', // Adjusted bottom padding for legend
        top: '10%', // Reduced top padding since title is removed
        containLabel: true,
      },
      legend: {
        show: true,
        bottom: '0',
        textStyle: {
          color: getChartThemeColor(),
        },
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: timestamps.value,
        axisLabel: {
          color: getChartThemeColor(),
          rotate: 45,
        },
        axisLine: { lineStyle: { color: getChartThemeColor() } },
      },
      yAxis: {
        type: 'value',
        splitLine: {
          lineStyle: {
            color: getSplitLineColor(),
          },
        },
        axisLabel: { color: getChartThemeColor() },
      },
    };

    if (cpuChartRef.value) {
      cpuChart = echarts.init(cpuChartRef.value);
      cpuChart.group = 'monitorGroup';
      bindChartEvents(cpuChart);
      cpuChart.setOption({
        ...commonOption,
        series: [
          {
            name: '总占用',
            type: 'line',
            smooth: true,
            data: cpuData.value,
            showSymbol: false,
            areaStyle: { opacity: 0.3 },
          },
          {
            name: '最高核心',
            type: 'line',
            smooth: true,
            data: cpuMaxCoreData.value,
            showSymbol: false,
          },
        ],
      });
      if (viewMode.value !== 'realtime') enableBrushSelection(cpuChart);
    }

    if (memChartRef.value) {
      memChart = echarts.init(memChartRef.value);
      memChart.group = 'monitorGroup';
      bindChartEvents(memChart);
      memChart.setOption({
        ...commonOption,
        series: [
          {
            name: '已用内存',
            type: 'line',
            smooth: true,
            data: memUsedData.value,
            showSymbol: false,
            areaStyle: { opacity: 0.3 },
          },
          {
            name: 'Swap已用',
            type: 'line',
            smooth: true,
            data: swapUsedData.value,
            showSymbol: false,
            lineStyle: { type: 'dashed' },
          },
        ],
      });
      if (viewMode.value !== 'realtime') enableBrushSelection(memChart);
    }

    if (netChartRef.value) {
      netChart = echarts.init(netChartRef.value);
      netChart.group = 'monitorGroup';
      bindChartEvents(netChart);
      netChart.setOption({
        ...commonOption,
        series: [
          { name: '上行', type: 'line', smooth: true, data: netUpData.value, showSymbol: false },
          { name: '下行', type: 'line', smooth: true, data: netDownData.value, showSymbol: false },
        ],
      });
      if (viewMode.value !== 'realtime') enableBrushSelection(netChart);
    }

    if (diskChartRef.value) {
      diskChart = echarts.init(diskChartRef.value);
      diskChart.group = 'monitorGroup';
      bindChartEvents(diskChart);
      diskChart.setOption({
        ...commonOption,
        series: [
          {
            name: '已用空间',
            type: 'line',
            smooth: true,
            data: diskUsedData.value,
            showSymbol: false,
            areaStyle: { opacity: 0.3 },
          },
        ],
      });
      if (viewMode.value !== 'realtime') enableBrushSelection(diskChart);
    }

    // Connect all charts in the group
    echarts.connect('monitorGroup');
  };

  const updateCharts = () => {
    const commonUpdate = {
      xAxis: { data: timestamps.value },
    };

    cpuChart?.setOption({
      ...commonUpdate,
      series: [{ data: cpuData.value }, { data: cpuMaxCoreData.value }],
    });

    memChart?.setOption({
      ...commonUpdate,
      series: [{ data: memUsedData.value }, { data: swapUsedData.value }],
    });

    netChart?.setOption({
      ...commonUpdate,
      series: [{ data: netUpData.value }, { data: netDownData.value }],
    });

    diskChart?.setOption({
      ...commonUpdate,
      series: [{ data: diskUsedData.value }],
    });
  };

  watch(
    timestamps,
    () => {
      updateCharts();
    },
    { deep: true }
  );

  watch(
    () => themeStore.isDark,
    () => {
      cpuChart?.dispose();
      memChart?.dispose();
      netChart?.dispose();
      diskChart?.dispose();
      nextTick(() => {
        initCharts();
      });
    }
  );

  watch(viewMode, (newMode) => {
    if (newMode !== 'realtime') {
      nextTick(() => {
        if (cpuChart) enableBrushSelection(cpuChart);
        if (memChart) enableBrushSelection(memChart);
        if (netChart) enableBrushSelection(netChart);
        if (diskChart) enableBrushSelection(diskChart);
      });
    } else {
      const resetCursor = (chart: echarts.ECharts | null) => {
        if (!chart) return;
        clearBrush(chart);
        chart.dispatchAction({
          type: 'takeGlobalCursor',
          key: 'dataZoomSelect',
          dataZoomSelectActive: false, // Ensure other modes are off
        });
        // Reset to default cursor
        chart.dispatchAction({
          type: 'takeGlobalCursor',
        });
      };

      resetCursor(cpuChart);
      resetCursor(memChart);
      resetCursor(netChart);
      resetCursor(diskChart);
    }
  });

  // Debounce helper
  const debounce = (fn: Function, delay: number) => {
    let timeoutId: any;
    return (...args: any[]) => {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => fn(...args), delay);
    };
  };

  const handleBrushEnd = debounce((params: any) => {
    // Only handle if it comes from brush action
    if (!params.areas || params.areas.length === 0) return;

    const range = params.areas[0].coordRange;
    // range is [startIndex, endIndex] for category axis
    let startIndex = Math.floor(range[0]);
    let endIndex = Math.floor(range[1]);

    // Ensure indices are within bounds
    if (startIndex < 0) startIndex = 0;
    if (endIndex >= hRawTimestamps.value.length) endIndex = hRawTimestamps.value.length - 1;

    // Get timestamps using indices
    let startTime = hRawTimestamps.value[startIndex];
    let endTime = hRawTimestamps.value[endIndex];

    // Handle case where selection is too small or invalid
    if (startTime && endTime) {
      if (startTime >= endTime) {
        // If same point or invalid, expand range slightly (e.g. 1 minute window)
        // But we need valid boundaries.
        // If it's a single point, we can try to get more data around it.
        // Let's just ensure endTime > startTime by at least 1 second if they are equal
        if (startTime === endTime) {
          endTime = startTime + 60; // Add 60 seconds
          startTime = startTime - 60; // Minus 60 seconds
        }
      }

      loading.value = true;
      fetchCustomHistory(startTime, endTime).finally(() => {
        loading.value = false;
        // Re-enable selection mode after data reload
        const resetCharts = () => {
          if (cpuChart) {
            clearBrush(cpuChart);
            enableBrushSelection(cpuChart);
          }
          if (memChart) {
            clearBrush(memChart);
            enableBrushSelection(memChart);
          }
          if (netChart) {
            clearBrush(netChart);
            enableBrushSelection(netChart);
          }
          if (diskChart) {
            clearBrush(diskChart);
            enableBrushSelection(diskChart);
          }
        };
        // Reset immediately
        resetCharts();
      });
    } else {
      // Invalid selection or no data, just clear brush
      if (cpuChart) clearBrush(cpuChart);
      if (memChart) clearBrush(memChart);
      if (netChart) clearBrush(netChart);
      if (diskChart) clearBrush(diskChart);
    }
  }, 100);

  const clearBrush = (chart: echarts.ECharts) => {
    chart.dispatchAction({
      type: 'brush',
      areas: [],
    });
  };

  const enableBrushSelection = (chart: echarts.ECharts) => {
    // We use setTimeout to ensure the chart is ready and rendered
    setTimeout(() => {
      chart.dispatchAction({
        type: 'takeGlobalCursor',
        key: 'brush',
        brushOption: {
          brushType: 'lineX',
          brushMode: 'single',
        },
      });
    }, 100);
  };

  const bindChartEvents = (chart: echarts.ECharts) => {
    chart.on('brushEnd', handleBrushEnd);
  };

  const unbindChartEvents = (chart: echarts.ECharts) => {
    chart.off('brushEnd');
  };

  onMounted(() => {
    fetchConfig();
    initCharts();
    window.addEventListener('resize', handleResize);
  });

  const handleResize = () => {
    cpuChart?.resize();
    memChart?.resize();
    netChart?.resize();
    diskChart?.resize();
  };

  onUnmounted(() => {
    // Do NOT stop monitor here to keep it running in background
    // stopMonitor();
    window.removeEventListener('resize', handleResize);

    if (cpuChart) {
      unbindChartEvents(cpuChart);
      cpuChart.dispose();
    }
    if (memChart) {
      unbindChartEvents(memChart);
      memChart.dispose();
    }
    if (netChart) {
      unbindChartEvents(netChart);
      netChart.dispose();
    }
    if (diskChart) {
      unbindChartEvents(diskChart);
      diskChart.dispose();
    }
  });
</script>
