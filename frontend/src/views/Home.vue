<script setup lang="ts">
import { showAlert } from '../composables/useModal'
import { useI18n } from 'vue-i18n'
import { setLanguage } from '../i18n'
import { callRcloneAPI } from '../utils/api'

const { t, locale } = useI18n()

const toggleLanguage = () => {
  setLanguage(locale.value === 'zh' ? 'en' : 'zh')
}

const testApi = async () => {
  try {
    const res = await callRcloneAPI('core/version', {})
    showAlert('Rclone Version: ' + JSON.stringify(res))
  } catch (err) {
    showAlert('API Error: ' + err)
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-900 text-white flex flex-col items-center justify-center space-y-8 relative">
    
    <!-- Language Toggle Button -->
    <div class="absolute top-6 right-6">
      <button 
        @click="toggleLanguage" 
        class="px-4 py-2 bg-gray-800 text-gray-300 hover:text-white hover:bg-gray-700 rounded-lg transition-colors border border-gray-700 shadow-md">
        {{ t('message.switchLang') }}
      </button>
    </div>

    <div class="p-10 bg-gray-800 rounded-2xl shadow-xl shadow-blue-900/20 border border-gray-700 text-center relative overflow-hidden group">
      <!-- Glow effect -->
      <div class="absolute -inset-1 bg-gradient-to-r from-blue-600 to-teal-500 rounded-2xl blur opacity-10 group-hover:opacity-20 transition duration-1000 group-hover:duration-200"></div>
      
      <div class="relative">
        <div class="mb-6 flex justify-center">
          <div class="w-24 h-24 rounded-full bg-gradient-to-tr from-blue-500 to-teal-400 flex items-center justify-center shadow-lg shadow-blue-500/30 animate-pulse">
            <svg class="w-12 h-12 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
          </div>
        </div>
        <h1 class="text-4xl font-extrabold tracking-tight mb-4 bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-teal-300">
          {{ t('message.hello') }}
        </h1>
        <p class="text-gray-400 text-lg mb-8 max-w-md mx-auto leading-relaxed">
          Your modern, beautiful, and dynamic graphical interface for Rclone. Built with Wails, Vue 3, Vite, and Tailwind CSS.
        </p>
        <div class="flex justify-center space-x-4">
          <button @click="testApi" class="px-6 py-3 rounded-xl bg-blue-600 hover:bg-blue-500 active:scale-95 transition-all font-semibold shadow-lg shadow-blue-600/30">
            {{ t('message.getStarted') }}
          </button>
          <button class="px-6 py-3 rounded-xl bg-gray-700 hover:bg-gray-600 active:scale-95 transition-all font-semibold border border-gray-600 hover:border-gray-500 shadow-md">
            {{ t('message.configureRemote') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
