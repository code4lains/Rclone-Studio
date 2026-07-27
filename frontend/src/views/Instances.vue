<script setup lang="ts">
import { showAlert, showConfirm } from '../composables/useModal'
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useConnectionStore, useInstanceStore, type Instance } from '../stores'
import { callRcloneAPI } from '../utils/api'

const { t } = useI18n()
const router = useRouter()
const connectionStore = useConnectionStore()
const instanceStore = useInstanceStore()

const isLoading = ref(true)
const isDesktop = instanceStore.isDesktop

// Modal state
const showModal = ref(false)
const modalMode = ref<'add' | 'edit'>('add')
const form = ref<Instance>({
  id: '',
  name: '',
  type: isDesktop ? 'managed' : 'external',
  command: [],
  url: 'http://127.0.0.1:5572',
  user: '',
  pass: '',
  pid: 0,
  status: 'stopped'
})
const formCommandStr = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

const isTesting = ref(false)

const actionStates = ref<Record<string, 'starting' | 'stopping' | 'restarting' | 'entering'>>({})

const fetchInstances = async () => {
  await instanceStore.fetchInstances()
  isLoading.value = false
}

onMounted(() => {
  fetchInstances()
  if (isDesktop) {
    pollTimer = setInterval(fetchInstances, 2000)
  }
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

const openAddModal = () => {
  modalMode.value = 'add'
  form.value = {
    id: crypto.randomUUID(),
    name: '',
    type: isDesktop ? 'managed' : 'external',
    command: [],
    url: 'http://127.0.0.1:5572',
    user: '',
    pass: '',
    pid: 0,
    status: 'stopped'
  }
  formCommandStr.value = 'rcd --rc-no-auth --rc-serve --rc-allow-origin * --rc-addr=127.0.0.1:5572'
  showModal.value = true
}

const openEditModal = (inst: Instance) => {
  modalMode.value = 'edit'
  form.value = { ...inst }
  formCommandStr.value = inst.command?.join(' ') || ''
  showModal.value = true
}

const saveInstance = async () => {
  if (form.value.type === 'managed') {
    form.value.command = formCommandStr.value.split(' ').filter(s => s.trim() !== '')
  } else {
    form.value.command = []
    
    // Connectivity test for external instances
    isTesting.value = true
    try {
      const testUrl = form.value.url.replace(/\/$/, '') + '/rc/noop'
      
      if (isDesktop) {
        const u = new URL(form.value.url)
        const ip = u.hostname
        const port = u.port || (u.protocol === 'https:' ? '443' : '80')
        await window.go!.main!.App!.CallRcloneAPI('rc/noop', {}, ip, port, form.value.user, form.value.pass)
      } else {
        const headers: HeadersInit = { 'Content-Type': 'application/json' }
        if (form.value.user || form.value.pass) {
          headers['Authorization'] = `Basic ${btoa(`${form.value.user}:${form.value.pass}`)}`
        }
        const res = await fetch(testUrl, { method: 'POST', headers, body: JSON.stringify({}) })
        if (!res.ok) throw new Error(`HTTP error ${res.status}`)
      }
    } catch (e: any) {
      isTesting.value = false
      showAlert(t('message.instances.testFailed') + '\n' + e)
      return
    }
    isTesting.value = false
  }

  try {
    await instanceStore.saveInstance(form.value)
    showModal.value = false
    
    if (form.value.type === 'external' && !connectionStore.isConnected) {
      const savedInst = instanceStore.instances.find(i => i.name === form.value.name)
      if (savedInst) {
        enterDashboard(savedInst)
      }
    }
  } catch (e: any) {
    showAlert('Save failed: ' + e)
  }
}

const deleteInstance = async (id: string) => {
  if (!await showConfirm(t('message.instances.deleteConfirm'))) return
  try {
    const inst = instanceStore.instances.find(i => i.id === id)
    await instanceStore.deleteInstance(id)
    if (inst) {
      let ip = '127.0.0.1'
      let port = '5572'
      if (inst.url) {
        try {
          const u = new URL(inst.url)
          ip = u.hostname
          port = u.port || (u.protocol === 'https:' ? '443' : '80')
        } catch(e) {}
      }
      if (connectionStore.config.ip === ip && connectionStore.config.port === port) {
        connectionStore.setConfig({ ip: '', port: '', user: '', pass: '' })
        connectionStore.setConnected(false)
      }
    }
  } catch (e: any) {
    showAlert('Delete failed: ' + e)
  }
}

const startInstance = async (id: string) => {
  if (!isDesktop) return
  actionStates.value[id] = 'starting'
  try {
    await window.go!.main!.App!.StartInstance(id)
    await fetchInstances()
    
    // Auto connect if not connected to anything (retry up to 3 times since process just started)
    if (!connectionStore.isConnected) {
      const inst = instanceStore.instances.find(i => i.id === id)
      if (inst) enterDashboard(inst, 3)
    }
  } catch (e: any) {
    showAlert('Start failed: ' + e)
  } finally {
    delete actionStates.value[id]
  }
}

const stopInstance = async (inst: Instance) => {
  actionStates.value[inst.id] = 'stopping'
  try {
    if (isDesktop) {
      await window.go!.main!.App!.StopInstance(inst.id)
    } else {
      await instanceStore.stopExternalInstanceWeb(inst)
    }
    await fetchInstances()
  } catch (e: any) {
    showAlert('Stop failed: ' + e)
  } finally {
    delete actionStates.value[inst.id]
  }
}

const restartInstance = async (id: string) => {
  if (!isDesktop) return
  actionStates.value[id] = 'restarting'
  try {
    await window.go!.main!.App!.RestartInstance(id)
    await fetchInstances()
  } catch (e: any) {
    showAlert('Restart failed: ' + e)
  } finally {
    delete actionStates.value[id]
  }
}

const enterDashboard = async (inst: Instance, retries = 0) => {
  actionStates.value[inst.id] = 'entering'
  try {
    const u = new URL(inst.url)
    connectionStore.setConfig({
      ip: u.hostname,
      port: u.port || (u.protocol === 'https:' ? '443' : '80'),
      user: inst.user,
      pass: inst.pass
    })

    // test connection with retry
    let success = false
    let lastErr: any = null
    for (let i = 0; i <= retries; i++) {
      try {
        const res = await callRcloneAPI('rc/noop', {})
        if (res) {
          success = true
          break
        }
      } catch (e) {
        lastErr = e
        if (i < retries) await new Promise(r => setTimeout(r, 1000))
      }
    }
    
    if (success) {
      connectionStore.setConnected(true)
      router.push('/')
    } else {
      throw lastErr || new Error('Connection failed')
    }
  } catch (e) {
    connectionStore.setConnected(false)
    showAlert(t('message.instances.testFailed') + '\n' + e)
  } finally {
    delete actionStates.value[inst.id]
  }
}
</script>

<template>
  <div class="animate-in fade-in slide-in-from-bottom-4 duration-500">
    
    <div>
      <div class="flex items-center justify-between mb-8">
        <div>
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('message.instances.title') }}</h2>
          <p class="text-gray-500 text-sm mt-1">{{ t('message.instances.subtitle') }}</p>
        </div>
        <button 
          @click="openAddModal"
          class="px-4 py-2 bg-black dark:bg-white text-white dark:text-black font-medium rounded-lg hover:bg-gray-800 dark:hover:bg-gray-100 transition-colors shadow-sm text-sm"
        >
          {{ t('message.instances.addInstance') }}
        </button>
      </div>

      <div class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
        <div v-if="isLoading" class="p-8 text-center text-gray-500 text-sm">
          Loading...
        </div>
        <div v-else-if="instanceStore.instances.length === 0" class="p-12 text-center text-gray-500">
          {{ t('message.instances.noInstances') }}
        </div>
        <table v-else class="w-full text-left border-collapse">
          <thead class="bg-gray-50/50 dark:bg-gray-900/30 border-b border-gray-200 dark:border-gray-800">
            <tr>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ t('message.instances.name') }}</th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ t('message.instances.type') }}</th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ t('message.instances.status') }}</th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ t('message.instances.url') }}</th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider text-right">{{ t('message.instances.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="inst in instanceStore.instances" :key="inst.id" class="hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors">
              <td class="px-6 py-4 font-medium text-gray-900 dark:text-gray-100 text-sm">{{ inst.name }}</td>
              <td class="px-6 py-4 text-sm">
                <span class="px-2.5 py-1 rounded-md text-xs font-medium bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400">
                  {{ inst.type === 'managed' ? t('message.instances.managed') : t('message.instances.external') }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm">
                <div v-if="inst.type === 'managed' || inst.status === 'running'" class="flex items-center space-x-2">
                  <span class="relative flex h-2.5 w-2.5">
                    <span v-if="inst.status === 'running'" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                    <span class="relative inline-flex rounded-full h-2.5 w-2.5" :class="inst.status === 'running' ? 'bg-green-500' : 'bg-red-500'"></span>
                  </span>
                  <span class="text-gray-700 dark:text-gray-300">{{ inst.status === 'running' ? t('message.instances.running') : t('message.instances.stopped') }}</span>
                  <span v-if="inst.status === 'running' && inst.pid" class="text-xs text-gray-400 ml-1">(PID: {{ inst.pid }})</span>
                </div>
                <div v-else class="text-gray-400">-</div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 font-mono">{{ inst.url }}</td>
              <td class="px-6 py-4 text-right space-x-2">
                <!-- Managed specific logic -->
                <template v-if="inst.type === 'managed'">
                  <template v-if="inst.status === 'stopped'">
                    <button v-if="isDesktop" @click="startInstance(inst.id)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm bg-black dark:bg-white text-white dark:text-black rounded-md font-medium hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      <svg v-if="actionStates[inst.id] === 'starting'" class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                      {{ actionStates[inst.id] === 'starting' ? (t('message.instances.starting') || '启动中...') : t('message.instances.btnStart') }}
                    </button>
                    <button @click="openEditModal(inst)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm border border-gray-200 dark:border-gray-700 rounded-md font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      {{ t('message.instances.btnEdit') }}
                    </button>
                    <button @click="deleteInstance(inst.id)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-md font-medium hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      {{ t('message.instances.btnDelete') }}
                    </button>
                  </template>
                  <template v-else>
                    <button @click="enterDashboard(inst)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      <svg v-if="actionStates[inst.id] === 'entering'" class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                      {{ actionStates[inst.id] === 'entering' ? (t('message.instances.entering') || '连接中...') : t('message.instances.btnEnter') }}
                    </button>
                    <button v-if="isDesktop" @click="restartInstance(inst.id)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm border border-gray-200 dark:border-gray-700 rounded-md font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      <svg v-if="actionStates[inst.id] === 'restarting'" class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                      {{ actionStates[inst.id] === 'restarting' ? (t('message.instances.restarting') || '重启中...') : t('message.instances.btnRestart') }}
                    </button>
                    <button v-if="isDesktop" @click="stopInstance(inst)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-md font-medium hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      <svg v-if="actionStates[inst.id] === 'stopping'" class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                      {{ actionStates[inst.id] === 'stopping' ? (t('message.instances.stopping') || '停止中...') : t('message.instances.btnStop') }}
                    </button>
                  </template>
                </template>

                <!-- External specific logic -->
                <template v-else>
                  <button @click="enterDashboard(inst)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                    <svg v-if="actionStates[inst.id] === 'entering'" class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                    {{ actionStates[inst.id] === 'entering' ? (t('message.instances.entering') || '连接中...') : t('message.instances.btnEnter') }}
                  </button>
                  <button @click="openEditModal(inst)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm border border-gray-200 dark:border-gray-700 rounded-md font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                    {{ t('message.instances.btnEdit') }}
                  </button>
                  <button v-if="isDesktop" @click="stopInstance(inst)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-md font-medium hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                    <svg v-if="actionStates[inst.id] === 'stopping'" class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                    {{ actionStates[inst.id] === 'stopping' ? (t('message.instances.stopping') || '停止中...') : t('message.instances.btnStop') }}
                  </button>
                  <button @click="deleteInstance(inst.id)" :disabled="actionStates[inst.id] !== undefined" class="inline-flex items-center justify-center px-3 py-1.5 text-sm border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-md font-medium hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                    {{ t('message.instances.btnDelete') }}
                  </button>
                </template>

              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div class="bg-white dark:bg-black w-full max-w-lg rounded-2xl shadow-xl border border-gray-200 dark:border-gray-800 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800">
          <h3 class="text-lg font-semibold">{{ modalMode === 'add' ? t('message.instances.modalTitleAdd') : t('message.instances.modalTitleEdit') }}</h3>
        </div>
        
        <form @submit.prevent="saveInstance">
          <div class="p-6 space-y-4 max-h-[70vh] overflow-y-auto">
            
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('message.instances.formName') }} *</label>
              <input v-model="form.name" type="text" required class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow" />
            </div>

            <!-- Hide Type selector in Web environment, force External -->
            <div v-if="isDesktop">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{{ t('message.instances.formType') }}</label>
              <div class="flex space-x-4">
                <label class="flex items-center space-x-2 cursor-pointer">
                  <input type="radio" v-model="form.type" value="managed" class="text-blue-600 focus:ring-blue-500" :disabled="form.status === 'running'" />
                  <span>{{ t('message.instances.managed') }}</span>
                </label>
                <label class="flex items-center space-x-2 cursor-pointer">
                  <input type="radio" v-model="form.type" value="external" class="text-blue-600 focus:ring-blue-500" :disabled="form.status === 'running'" />
                  <span>{{ t('message.instances.external') }}</span>
                </label>
              </div>
            </div>

            <template v-if="form.type === 'managed'">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('message.instances.formCommand') }} *</label>
                <textarea v-model="formCommandStr" rows="2" required class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow font-mono text-sm"></textarea>
              </div>
            </template>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('message.instances.formUrl') }} *</label>
              <input v-model="form.url" type="url" required class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow font-mono text-sm" />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('message.instances.formUser') }}</label>
                <input v-model="form.user" type="text" class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('message.instances.formPass') }}</label>
                <input v-model="form.pass" type="password" class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow" />
              </div>
            </div>

          </div>
          
          <div class="px-6 py-4 bg-gray-50 dark:bg-gray-900/30 border-t border-gray-200 dark:border-gray-800 flex justify-end space-x-3">
            <button type="button" @click="showModal = false" :disabled="isTesting" class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 disabled:opacity-50 transition-colors">
              {{ t('message.instances.cancel') }}
            </button>
            <button type="submit" :disabled="isTesting" class="px-5 py-2 text-sm font-medium bg-black dark:bg-white text-white dark:text-black rounded-lg hover:bg-gray-800 dark:hover:bg-gray-100 disabled:opacity-50 transition-colors flex items-center">
              <svg v-if="isTesting" class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
              {{ isTesting ? t('message.instances.connecting') || 'Testing...' : t('message.instances.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

  </div>
</template>
