<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { setLanguage } from '../i18n'
import { onMounted, onUnmounted, ref, watch } from 'vue'

import { useConnectionStore } from '../stores'

import { callRcloneAPI } from '../utils/api'

const { t, locale } = useI18n()
const route = useRoute()
const connectionStore = useConnectionStore()

// Load config on layout mount
connectionStore.loadConfig()

const toggleLanguage = () => {
  setLanguage(locale.value === 'zh' ? 'en' : 'zh')
}

let pingTimer: ReturnType<typeof setInterval> | null = null

const startPing = () => {
  if (!pingTimer) {
    pingTimer = setInterval(checkConnection, 3000)
  }
}

const stopPing = () => {
  if (pingTimer) {
    clearInterval(pingTimer)
    pingTimer = null
  }
}

const isInitializing = ref(true)

const checkConnection = async () => {
  if (!connectionStore.config.ip) {
    connectionStore.setConnected(false)
    isInitializing.value = false
    return
  }
  try {
    await callRcloneAPI('rc/noop', {}, true)
    if (!connectionStore.isConnected) {
      connectionStore.setConnected(true)
    }
  } catch (e) {
    connectionStore.setConnected(false)
  } finally {
    isInitializing.value = false
  }
}

watch(() => connectionStore.isConnected, (connected) => {
  if (connected) {
    startPing()
  } else {
    stopPing()
  }
})

onMounted(() => {
  checkConnection()
})

onUnmounted(() => {
  stopPing()
})

const navItems = [
  { name: 'dashboard', path: '/' },
  { name: 'instances', path: '/instances' },
  { name: 'config', path: '/config' },
  { name: 'files', path: '/files' },
  { name: 'mounts', path: '/mounts' },
  { name: 'help', path: '/help' },
]
</script>

<template>
  <div class="flex h-screen bg-white text-black dark:bg-black dark:text-gray-100 font-sans selection:bg-gray-300 dark:selection:bg-gray-700">
    <!-- Sidebar -->
    <aside class="w-64 border-r border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-black flex flex-col">
      <div class="h-16 flex items-center px-6 border-b border-gray-200 dark:border-gray-800">
        <div class="w-8 h-8 rounded-lg bg-black dark:bg-white flex items-center justify-center mr-3">
          <svg class="w-5 h-5 text-white dark:text-black" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
        </div>
        <span class="font-semibold tracking-tight text-lg">Rclone Studio</span>
      </div>
      
      <nav class="flex-1 overflow-y-auto py-4 px-3 space-y-1">
        <router-link
          v-for="item in navItems"
          :key="item.name"
          :to="item.path"
          class="flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors"
          :class="[
            route.path === item.path 
              ? 'bg-gray-200 text-black dark:bg-gray-800 dark:text-white' 
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-900 dark:hover:text-gray-100'
          ]"
        >
          {{ t('message.menu.' + item.name) }}
        </router-link>
      </nav>
      
      <!-- Bottom section of sidebar -->
      <div class="p-4 border-t border-gray-200 dark:border-gray-800 text-xs text-gray-500 dark:text-gray-400">
        v1.0.0
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <!-- Header -->
      <header class="h-16 flex items-center justify-between px-8 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-black shrink-0">
        
        <!-- Status Placeholder -->
        <div class="flex items-center space-x-2">
          <div class="relative flex h-3 w-3">
            <span v-if="connectionStore.isConnected" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-3 w-3" :class="connectionStore.isConnected ? 'bg-green-500' : 'bg-red-500'"></span>
          </div>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ connectionStore.isConnected ? t('message.status.connected') : t('message.status.disconnected') }}
          </span>
        </div>

        <!-- Right actions -->
        <div class="flex items-center space-x-4">
          <button 
            @click="toggleLanguage" 
            class="text-sm px-3 py-1.5 rounded-md border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700">
            {{ locale === 'zh' ? 'EN' : '中文' }}
          </button>
        </div>
      </header>

      <!-- Router View Container -->
      <div class="flex-1 overflow-auto bg-gray-50 dark:bg-black p-8">
        <div class="mx-auto max-w-5xl">
          
          <div v-if="isInitializing" class="flex flex-col items-center justify-center py-32">
            <svg class="animate-spin h-8 w-8 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
          </div>

          <div v-else-if="!connectionStore.isConnected && route.name !== 'instances' && route.name !== 'help'" class="flex flex-col items-center justify-center py-20 text-center animate-in fade-in zoom-in-95 duration-300">
            <div class="w-20 h-20 bg-red-50 dark:bg-red-900/20 text-red-500 rounded-full flex items-center justify-center mb-6 border border-red-100 dark:border-red-900/30">
              <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
            </div>
            <h2 class="text-2xl font-bold mb-3">{{ t('message.status.notConnectedTitle') }}</h2>
            <p class="text-gray-500 dark:text-gray-400 mb-8 max-w-md leading-relaxed">{{ t('message.status.notConnectedDesc') }}</p>
            <router-link to="/instances" class="px-6 py-3 bg-black dark:bg-white text-white dark:text-black font-medium rounded-xl hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors shadow-sm flex items-center space-x-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7"></path></svg>
              <span>{{ t('message.status.goToInstances') }}</span>
            </router-link>
          </div>

          <router-view v-else v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
          
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
