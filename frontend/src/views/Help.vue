<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, ref } from 'vue'

const { t, locale } = useI18n()

const os = ref<'windows' | 'linux' | 'macos'>('windows')

const rcloneLink = computed(() => {
  return locale.value === 'zh' 
    ? 'https://rclone.org/downloads/'
    : 'https://rclone.org/downloads/'
})

const winfspLink = computed(() => {
  return 'https://winfsp.dev/rel/'
})

const macfuseLink = computed(() => {
  return 'https://osxfuse.github.io/'
})

</script>

<template>
  <div class="animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-4xl mx-auto pb-12">
    <div class="mb-6 flex flex-col md:flex-row md:items-end justify-between gap-4">
      <div>
        <h2 class="text-2xl font-semibold tracking-tight">{{ t('message.help.title') }}</h2>
        <p class="text-gray-500 text-sm mt-1">{{ t('message.help.subtitle') }}</p>
      </div>
      
      <!-- OS Toggle -->
      <div class="flex space-x-1 bg-gray-200/50 dark:bg-gray-800 p-1 rounded-lg w-max shrink-0">
        <button @click="os = 'windows'" :class="os === 'windows' ? 'bg-white dark:bg-black shadow text-black dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'" class="px-4 py-1.5 rounded-md text-sm font-medium transition-all">
          {{ t('message.help.os_windows') }}
        </button>
        <button @click="os = 'linux'" :class="os === 'linux' ? 'bg-white dark:bg-black shadow text-black dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'" class="px-4 py-1.5 rounded-md text-sm font-medium transition-all">
          {{ t('message.help.os_linux') }}
        </button>
        <button @click="os = 'macos'" :class="os === 'macos' ? 'bg-white dark:bg-black shadow text-black dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'" class="px-4 py-1.5 rounded-md text-sm font-medium transition-all">
          {{ t('message.help.os_macos') }}
        </button>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
      <!-- Rclone Card -->
      <div class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col transition-all hover:shadow-md hover:border-blue-300 dark:hover:border-blue-700">
        <div class="p-6 border-b border-gray-100 dark:border-gray-800 flex items-center space-x-4 bg-gray-50/50 dark:bg-gray-900/30">
          <div class="w-12 h-12 bg-blue-100 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 rounded-xl flex items-center justify-center shrink-0">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
          </div>
          <div>
            <h3 class="text-lg font-bold">Rclone</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ t('message.help.rcloneDesc') }}</p>
          </div>
        </div>
        <div class="p-6 flex-1 flex flex-col">
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-6 leading-relaxed">
            {{ t('message.help.rcloneInfo') }}
          </p>
          
          <div class="mt-auto">
            <div class="bg-gray-100 dark:bg-gray-900 rounded-lg p-4 mb-6">
              <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{{ t('message.help.installSteps') }}</h4>
              <ol v-if="os === 'windows'" class="list-decimal list-inside text-sm space-y-1.5 text-gray-700 dark:text-gray-300">
                <li>{{ t('message.help.rcloneStep1') }}</li>
                <li>{{ t('message.help.rcloneStep2') }}</li>
                <li>{{ t('message.help.rcloneStep3') }}</li>
              </ol>
              <ol v-else-if="os === 'linux'" class="list-decimal list-inside text-sm space-y-1.5 text-gray-700 dark:text-gray-300">
                <li>{{ t('message.help.rcloneStep1_linux') }}</li>
                <li>{{ t('message.help.rcloneStep2_linux') }}</li>
                <li>{{ t('message.help.rcloneStep3_linux') }}</li>
              </ol>
              <ol v-else-if="os === 'macos'" class="list-decimal list-inside text-sm space-y-1.5 text-gray-700 dark:text-gray-300">
                <li>{{ t('message.help.rcloneStep1_macos') }}</li>
                <li>{{ t('message.help.rcloneStep2_macos') }}</li>
                <li>{{ t('message.help.rcloneStep3_macos') }}</li>
              </ol>
            </div>
            
            <a :href="rcloneLink" target="_blank" rel="noopener noreferrer" class="flex items-center justify-center w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors shadow-sm">
              {{ t('message.help.downloadBtn') }} Rclone
              <svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path></svg>
            </a>
          </div>
        </div>
      </div>

      <!-- WinFSP Card (Windows Only) -->
      <div v-if="os === 'windows'" class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col transition-all hover:shadow-md hover:border-green-300 dark:hover:border-green-700">
        <div class="p-6 border-b border-gray-100 dark:border-gray-800 flex items-center space-x-4 bg-gray-50/50 dark:bg-gray-900/30">
          <div class="w-12 h-12 bg-green-100 dark:bg-green-900/40 text-green-600 dark:text-green-400 rounded-xl flex items-center justify-center shrink-0">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
          </div>
          <div>
            <h3 class="text-lg font-bold">WinFSP</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ t('message.help.winfspDesc') }}</p>
          </div>
        </div>
        <div class="p-6 flex-1 flex flex-col">
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-6 leading-relaxed">
            {{ t('message.help.winfspInfo') }}
          </p>
          
          <div class="mt-auto">
            <div class="bg-gray-100 dark:bg-gray-900 rounded-lg p-4 mb-6">
              <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{{ t('message.help.installSteps') }}</h4>
              <ol class="list-decimal list-inside text-sm space-y-1.5 text-gray-700 dark:text-gray-300">
                <li>{{ t('message.help.winfspStep1') }}</li>
                <li>{{ t('message.help.winfspStep2') }}</li>
                <li>{{ t('message.help.winfspStep3') }}</li>
              </ol>
            </div>
            
            <a :href="winfspLink" target="_blank" rel="noopener noreferrer" class="flex items-center justify-center w-full py-2.5 px-4 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition-colors shadow-sm">
              {{ t('message.help.downloadBtn') }} WinFSP
              <svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path></svg>
            </a>
          </div>
        </div>
      </div>
      
      <!-- macFUSE Card (macOS Only) -->
      <div v-if="os === 'macos'" class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col transition-all hover:shadow-md hover:border-purple-300 dark:hover:border-purple-700">
        <div class="p-6 border-b border-gray-100 dark:border-gray-800 flex items-center space-x-4 bg-gray-50/50 dark:bg-gray-900/30">
          <div class="w-12 h-12 bg-purple-100 dark:bg-purple-900/40 text-purple-600 dark:text-purple-400 rounded-xl flex items-center justify-center shrink-0">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
          </div>
          <div>
            <h3 class="text-lg font-bold">macFUSE</h3>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ t('message.help.macfuseDesc') }}</p>
          </div>
        </div>
        <div class="p-6 flex-1 flex flex-col">
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-6 leading-relaxed">
            {{ t('message.help.macfuseInfo') }}
          </p>
          
          <div class="mt-auto">
            <div class="bg-gray-100 dark:bg-gray-900 rounded-lg p-4 mb-6">
              <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{{ t('message.help.installSteps') }}</h4>
              <ol class="list-decimal list-inside text-sm space-y-1.5 text-gray-700 dark:text-gray-300">
                <li>{{ t('message.help.macfuseStep1') }}</li>
                <li>{{ t('message.help.macfuseStep2') }}</li>
                <li>{{ t('message.help.macfuseStep3') }}</li>
              </ol>
            </div>
            
            <a :href="macfuseLink" target="_blank" rel="noopener noreferrer" class="flex items-center justify-center w-full py-2.5 px-4 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition-colors shadow-sm">
              {{ t('message.help.downloadBtn') }} macFUSE
              <svg class="w-4 h-4 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path></svg>
            </a>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>
