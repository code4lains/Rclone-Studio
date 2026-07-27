<script setup lang="ts">
import { showAlert } from '../composables/useModal'
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { callRcloneAPI } from '../utils/api'

const { t } = useI18n()

interface MountItem {
  MountPoint: string
  Fs: string
}

const isDesktop = computed(() => {
  return !!window.go?.main?.App
})

const availableRemotes = ref<string[]>([])
const activeMounts = ref<MountItem[]>([])

const isLoading = ref(false)
const showMountModal = ref(false)
const selectedRemote = ref('')
const mountPoint = ref('')
const isMounting = ref(false)

const fetchData = async () => {
  if (!isDesktop.value) return
  isLoading.value = true
  try {
    const [configRes, mountRes] = await Promise.all([
      callRcloneAPI('config/listremotes', {}),
      callRcloneAPI('mount/listmounts', {})
    ])
    availableRemotes.value = configRes.remotes || []
    activeMounts.value = mountRes.mountPoints || []
  } catch (error) {
    console.error('Failed to fetch mounts data', error)
  } finally {
    isLoading.value = false
  }
}

const openMountModal = (remote: string) => {
  selectedRemote.value = remote
  mountPoint.value = ''
  showMountModal.value = true
}

const confirmMount = async () => {
  if (!mountPoint.value) {
    showAlert(t('message.mounts.mountPointRequired'))
    return
  }
  isMounting.value = true
  try {
    const fs = selectedRemote.value.endsWith(':') ? selectedRemote.value : `${selectedRemote.value}:`
    await callRcloneAPI('mount/mount', {
      fs: fs,
      mountPoint: mountPoint.value
    })
    showAlert(t('message.mounts.mountSuccess'))
    showMountModal.value = false
    await fetchData()
  } catch (error: any) {
    showAlert('Mount failed: ' + error)
  } finally {
    isMounting.value = false
  }
}

const handleUnmount = async (mountPath: string) => {
  try {
    await callRcloneAPI('mount/unmount', {
      mountPoint: mountPath
    })
    showAlert(t('message.mounts.unmountSuccess'))
    await fetchData()
  } catch (error: any) {
    showAlert('Unmount failed: ' + error)
  }
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="animate-in fade-in slide-in-from-bottom-4 duration-500">
    <div class="mb-8">
      <h2 class="text-2xl font-semibold tracking-tight">{{ t('message.mounts.title') }}</h2>
    </div>

    <!-- Web Environment Warning -->
    <div v-if="!isDesktop" class="p-8 bg-yellow-50 dark:bg-yellow-900/20 rounded-2xl border border-yellow-200 dark:border-yellow-800 text-center">
      <svg class="w-12 h-12 text-yellow-500 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
      <h3 class="text-lg font-medium text-yellow-800 dark:text-yellow-200">{{ t('message.mounts.desktopOnly') }}</h3>
    </div>

    <!-- Desktop Split View -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-8">
      
      <!-- Left: Available Remotes -->
      <div class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col h-[calc(100vh-14rem)]">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/30 shrink-0">
          <h3 class="font-medium text-gray-900 dark:text-gray-100">{{ t('message.mounts.availableRemotes') }}</h3>
        </div>
        <div class="flex-1 overflow-y-auto p-4">
          <div v-if="isLoading" class="text-center text-sm text-gray-500 py-4">{{ t('message.files.loading') }}</div>
          <div v-else-if="availableRemotes.length === 0" class="text-center text-sm text-gray-500 py-4">
            {{ t('message.mounts.noRemotes') }}
          </div>
          <div v-else class="space-y-3">
            <div 
              v-for="r in availableRemotes" 
              :key="r"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 flex items-center justify-between group hover:border-blue-300 dark:hover:border-blue-700 transition-colors"
            >
              <div class="flex items-center space-x-3">
                <div class="p-2 bg-blue-50 dark:bg-blue-900/20 text-blue-500 rounded-lg">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"></path></svg>
                </div>
                <span class="font-medium">{{ r }}</span>
              </div>
              <button 
                @click="openMountModal(r)"
                class="opacity-0 group-hover:opacity-100 px-3 py-1.5 text-sm bg-black dark:bg-white text-white dark:text-black rounded-md font-medium transition-opacity"
              >
                {{ t('message.mounts.mountBtn') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Active Mounts -->
      <div class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col h-[calc(100vh-14rem)]">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/30 shrink-0">
          <h3 class="font-medium text-gray-900 dark:text-gray-100">{{ t('message.mounts.activeMounts') }}</h3>
        </div>
        <div class="flex-1 overflow-y-auto p-4">
          <div v-if="isLoading" class="text-center text-sm text-gray-500 py-4">{{ t('message.files.loading') }}</div>
          <div v-else-if="activeMounts.length === 0" class="text-center text-sm text-gray-500 py-4">
            {{ t('message.mounts.noMounts') }}
          </div>
          <div v-else class="space-y-3">
            <div 
              v-for="m in activeMounts" 
              :key="m.MountPoint"
              class="p-4 rounded-xl border border-green-200 dark:border-green-900/30 bg-green-50/50 dark:bg-green-900/10 flex items-center justify-between"
            >
              <div class="flex items-center space-x-3">
                <div class="p-2 bg-green-100 dark:bg-green-900/40 text-green-600 dark:text-green-400 rounded-lg shrink-0">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                </div>
                <div class="overflow-hidden">
                  <div class="font-medium text-gray-900 dark:text-gray-100 truncate">{{ m.MountPoint }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ m.Fs }}</div>
                </div>
              </div>
              <button 
                @click="handleUnmount(m.MountPoint)"
                class="px-3 py-1.5 text-sm border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md font-medium transition-colors shrink-0"
              >
                {{ t('message.mounts.unmountBtn') }}
              </button>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Mount Modal -->
    <div v-if="showMountModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div class="bg-white dark:bg-black w-full max-w-md rounded-2xl shadow-xl border border-gray-200 dark:border-gray-800 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800">
          <h3 class="text-lg font-semibold">{{ t('message.mounts.modalTitle') }}</h3>
        </div>
        <div class="p-6">
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
            Mounting <span class="font-bold text-gray-900 dark:text-gray-100">{{ selectedRemote }}</span>
          </p>
          <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('message.mounts.mountPointLabel') }}</label>
            <input 
              v-model="mountPoint" 
              type="text" 
              class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow"
              placeholder="e.g., Z: or /mnt/data"
              @keyup.enter="confirmMount"
            />
          </div>
        </div>
        <div class="px-6 py-4 bg-gray-50 dark:bg-gray-900/30 border-t border-gray-200 dark:border-gray-800 flex justify-end space-x-3">
          <button @click="showMountModal = false" class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 transition-colors">
            {{ t('message.mounts.cancel') }}
          </button>
          <button 
            @click="confirmMount" 
            :disabled="isMounting"
            class="px-4 py-2 text-sm font-medium bg-black dark:bg-white text-white dark:text-black rounded-lg hover:bg-gray-800 dark:hover:bg-gray-100 disabled:opacity-50 transition-colors"
          >
            {{ isMounting ? t('message.mounts.mounting') : t('message.mounts.confirm') }}
          </button>
        </div>
      </div>
    </div>

  </div>
</template>
