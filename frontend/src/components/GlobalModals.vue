<script setup lang="ts">
import { alertState, confirmState, promptState } from '../composables/useModal'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const closeAlert = () => {
  if (alertState.value.resolve) alertState.value.resolve()
  alertState.value.isOpen = false
}

const confirmYes = () => {
  if (confirmState.value.resolve) confirmState.value.resolve(true)
  confirmState.value.isOpen = false
}

const confirmNo = () => {
  if (confirmState.value.resolve) confirmState.value.resolve(false)
  confirmState.value.isOpen = false
}

const promptSubmit = () => {
  if (promptState.value.resolve) promptState.value.resolve(promptState.value.value)
  promptState.value.isOpen = false
}

const promptCancel = () => {
  if (promptState.value.resolve) promptState.value.resolve(null)
  promptState.value.isOpen = false
}
</script>

<template>
  <!-- Alert Modal -->
  <div v-if="alertState.isOpen" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
    <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-sm overflow-hidden animate-in zoom-in-95 duration-200">
      <div class="px-6 py-5">
        <p class="text-gray-700 dark:text-gray-300">{{ alertState.message }}</p>
      </div>
      <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800/50 flex justify-end">
        <button @click="closeAlert" class="px-4 py-2 text-sm font-medium bg-black dark:bg-white text-white dark:text-black rounded-lg transition-colors">{{ t('message.files.confirm') || 'OK' }}</button>
      </div>
    </div>
  </div>

  <!-- Confirm Modal -->
  <div v-if="confirmState.isOpen" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
    <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-sm overflow-hidden animate-in zoom-in-95 duration-200">
      <div class="px-6 py-5">
        <p class="text-gray-700 dark:text-gray-300">{{ confirmState.message }}</p>
      </div>
      <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800/50 flex justify-end space-x-3">
        <button @click="confirmNo" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors">{{ t('message.files.cancel') || 'Cancel' }}</button>
        <button @click="confirmYes" class="px-4 py-2 text-sm font-medium bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">{{ t('message.files.confirm') || 'Confirm' }}</button>
      </div>
    </div>
  </div>

  <!-- Prompt Modal -->
  <div v-if="promptState.isOpen" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
    <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in zoom-in-95 duration-200">
      <div class="px-6 py-5 border-b border-gray-100 dark:border-gray-800">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ promptState.title }}</h3>
      </div>
      <div class="px-6 py-5 space-y-4">
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">{{ promptState.label }}</label>
        <input 
          type="text" 
          v-model="promptState.value" 
          @keyup.enter="promptSubmit"
          class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500" 
        />
      </div>
      <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800/50 flex justify-end space-x-3">
        <button @click="promptCancel" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors">{{ t('message.files.cancel') || 'Cancel' }}</button>
        <button @click="promptSubmit" class="px-4 py-2 text-sm font-medium bg-black dark:bg-white text-white dark:text-black rounded-lg hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors">{{ t('message.files.confirm') || 'Confirm' }}</button>
      </div>
    </div>
  </div>
</template>
