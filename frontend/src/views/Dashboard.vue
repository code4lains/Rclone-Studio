<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { callRcloneAPI } from '../utils/api'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()

const stats = ref({
  uploadSpeed: 0,
  downloadSpeed: 0,
  activeTasks: 0,
  memoryUsageMB: 0
})

let timer: ReturnType<typeof setInterval> | null = null

const formatSpeed = (bytesPerSec: number) => {
  if (bytesPerSec === 0) return '0.00'
  const kb = bytesPerSec / 1024
  if (kb < 1024) return kb.toFixed(2) + ' KB/s'
  const mb = kb / 1024
  return mb.toFixed(2) + ' MB/s'
}

// Chart History Data
const maxDataPoints = 60
const timeLabels = ref<string[]>(Array(maxDataPoints).fill(''))
const uploadData = ref<number[]>(Array(maxDataPoints).fill(0))
const downloadData = ref<number[]>(Array(maxDataPoints).fill(0))
const memoryData = ref<number[]>(Array(maxDataPoints).fill(0))

const fetchData = async () => {
  try {
    const [statsRes, memRes] = await Promise.all([
      callRcloneAPI('core/stats', {}, true),
      callRcloneAPI('core/memstats', {}, true)
    ])

    // Rclone's core/stats returns a global 'speed' in bytes/sec. 
    // We map it to uploadSpeed as a primary indicator, unless we parse 'transferring' array deeply.
    stats.value.uploadSpeed = statsRes.speed || 0
    stats.value.downloadSpeed = 0 // Placeholder as rclone doesn't distinct globally in basic stats
    
    // statsRes.transfers is cumulative total completed. 
    // To get currently active tasks, we count the length of the transferring array.
    stats.value.activeTasks = (statsRes.transferring && Array.isArray(statsRes.transferring)) ? statsRes.transferring.length : 0
    
    // memRes.Sys is in bytes
    stats.value.memoryUsageMB = (memRes.Sys / (1024 * 1024)) || 0

    // Update Chart History
    uploadData.value = [...uploadData.value.slice(1), stats.value.uploadSpeed]
    downloadData.value = [...downloadData.value.slice(1), stats.value.downloadSpeed]
    memoryData.value = [...memoryData.value.slice(1), stats.value.memoryUsageMB]
    
    const now = new Date()
    timeLabels.value = [...timeLabels.value.slice(1), `${now.getHours()}:${now.getMinutes()}:${now.getSeconds()}`]
  } catch (error) {
    console.error('Failed to fetch dashboard stats:', error)
  }
}

onMounted(() => {
  fetchData()
  timer = setInterval(fetchData, 2000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})

// Network Chart Setup
const networkChartData = computed(() => ({
  labels: timeLabels.value,
  datasets: [
    {
      label: t('message.dashboard.uploadSpeed') || 'Upload',
      borderColor: '#3b82f6',
      backgroundColor: 'rgba(59, 130, 246, 0.1)',
      data: uploadData.value,
      fill: true,
      tension: 0.2,
      pointRadius: 0,
      borderWidth: 2
    },
    {
      label: t('message.dashboard.downloadSpeed') || 'Download',
      borderColor: '#22c55e',
      backgroundColor: 'rgba(34, 197, 94, 0.1)',
      data: downloadData.value,
      fill: true,
      tension: 0.2,
      pointRadius: 0,
      borderWidth: 2
    }
  ]
}))

const networkChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 0 }, // Disable animation for Task Manager feel
  scales: {
    x: { display: false },
    y: { 
      beginAtZero: true,
      grid: { color: 'rgba(156, 163, 175, 0.1)' },
      ticks: {
        callback: (value: string | number) => {
          const numValue = Number(value)
          if (numValue === 0) return '0'
          const kb = numValue / 1024
          if (kb < 1024) return kb.toFixed(0) + ' KB/s'
          return (kb / 1024).toFixed(1) + ' MB/s'
        }
      }
    }
  },
  plugins: {
    legend: { display: true, position: 'top' as const },
    tooltip: { enabled: false }
  },
  interaction: { intersect: false, mode: 'index' as const }
}

// Memory Chart Setup
const memoryChartData = computed(() => ({
  labels: timeLabels.value,
  datasets: [
    {
      label: t('message.dashboard.memoryUsage') || 'Memory',
      borderColor: '#f97316',
      backgroundColor: 'rgba(249, 115, 22, 0.1)',
      data: memoryData.value,
      fill: true,
      tension: 0.2,
      pointRadius: 0,
      borderWidth: 2
    }
  ]
}))

const memoryChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 0 },
  scales: {
    x: { display: false },
    y: { 
      beginAtZero: true,
      grid: { color: 'rgba(156, 163, 175, 0.1)' },
      ticks: {
        callback: (value: string | number) => Number(value) + ' MB'
      }
    }
  },
  plugins: {
    legend: { display: false },
    tooltip: { enabled: false }
  },
  interaction: { intersect: false, mode: 'index' as const }
}
</script>

<template>
  <div class="animate-in fade-in slide-in-from-bottom-4 duration-500 pb-8">
    <h2 class="text-2xl font-semibold tracking-tight mb-8">{{ t('message.dashboard.title') }}</h2>
    
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
      
      <!-- Upload Speed Card -->
      <div class="p-6 bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('message.dashboard.uploadSpeed') }}</h3>
          <div class="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg text-blue-600 dark:text-blue-400">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"></path></svg>
          </div>
        </div>
        <div class="flex items-baseline space-x-1">
          <span class="text-3xl font-bold tracking-tight text-blue-600 dark:text-blue-400">{{ formatSpeed(stats.uploadSpeed).split(' ')[0] }}</span>
          <span class="text-sm font-medium text-gray-500">{{ formatSpeed(stats.uploadSpeed).split(' ')[1] || t('message.dashboard.bytesPerSec') }}</span>
        </div>
      </div>

      <!-- Download Speed Card -->
      <div class="p-6 bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('message.dashboard.downloadSpeed') }}</h3>
          <div class="p-2 bg-green-50 dark:bg-green-900/20 rounded-lg text-green-600 dark:text-green-400">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path></svg>
          </div>
        </div>
        <div class="flex items-baseline space-x-1">
          <span class="text-3xl font-bold tracking-tight text-green-600 dark:text-green-400">{{ formatSpeed(stats.downloadSpeed).split(' ')[0] }}</span>
          <span class="text-sm font-medium text-gray-500">{{ formatSpeed(stats.downloadSpeed).split(' ')[1] || t('message.dashboard.bytesPerSec') }}</span>
        </div>
      </div>

      <!-- Active Tasks Card -->
      <div class="p-6 bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('message.dashboard.activeTasks') }}</h3>
          <div class="p-2 bg-purple-50 dark:bg-purple-900/20 rounded-lg text-purple-600 dark:text-purple-400">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg>
          </div>
        </div>
        <div class="flex items-baseline space-x-1">
          <span class="text-3xl font-bold tracking-tight">{{ stats.activeTasks }}</span>
        </div>
      </div>

      <!-- Memory Usage Card -->
      <div class="p-6 bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('message.dashboard.memoryUsage') }}</h3>
          <div class="p-2 bg-orange-50 dark:bg-orange-900/20 rounded-lg text-orange-600 dark:text-orange-400">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"></path></svg>
          </div>
        </div>
        <div class="flex items-baseline space-x-1">
          <span class="text-3xl font-bold tracking-tight text-orange-600 dark:text-orange-400">{{ stats.memoryUsageMB.toFixed(1) }}</span>
          <span class="text-sm font-medium text-gray-500">MB</span>
        </div>
      </div>
    </div>

    <!-- Charts Section -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      
      <!-- Network Chart -->
      <div class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
        <h3 class="text-base font-semibold mb-6 flex items-center space-x-2">
          <span class="w-2 h-6 bg-blue-500 rounded-sm"></span>
          <span>网络流量 (Network Traffic)</span>
        </h3>
        <div class="h-64 relative">
          <Line :data="networkChartData" :options="networkChartOptions" />
        </div>
      </div>

      <!-- Memory Chart -->
      <div class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
        <h3 class="text-base font-semibold mb-6 flex items-center space-x-2">
          <span class="w-2 h-6 bg-orange-500 rounded-sm"></span>
          <span>系统内存 (System Memory)</span>
        </h3>
        <div class="h-64 relative">
          <Line :data="memoryChartData" :options="memoryChartOptions" />
        </div>
      </div>

    </div>

  </div>
</template>
