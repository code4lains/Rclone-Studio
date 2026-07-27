<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { callRcloneAPI } from '../utils/api'
import { showAlert, showConfirm } from '../composables/useModal'

const { t, te } = useI18n()

const formatOptionName = (name: string) => {
  const key = `message.rcloneOptions.${name}`
  if (te(key)) return t(key)
  return name.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
}

const getProviderName = (provider: Provider) => {
  const key = `message.rcloneProviders.${provider.Prefix}`
  return te(key) ? t(key) : provider.Prefix
}

const getProviderDescription = (provider: Provider) => {
  const key = `message.rcloneProviderDescription.${provider.Prefix}`
  const desc = te(key) ? t(key) : provider.Description
  if (desc.toLowerCase() === provider.Prefix.toLowerCase() || desc === getProviderName(provider)) {
    return ''
  }
  return desc
}

const sanitizeKey = (key: string) => key.replace(/[^a-zA-Z0-9_-]/g, '_')

const formatExampleHelp = (providerPrefix: string, optName: string, exValue: string, originalHelp: string) => {
  const key = `message.rcloneExampleHelp.${providerPrefix}.${optName}.${sanitizeKey(exValue)}`
  if (te(key)) return t(key)
  return originalHelp
}

const formatOptionHelp = (providerPrefix: string, opt: ProviderOption) => {
  const key = `message.rcloneOptionHelp.${providerPrefix}.${opt.Name}`
  if (te(key)) return t(key)
  return opt.Help.split('\n')[0]
}

interface RemoteInfo {
  type: string
  [key: string]: string
}

interface ProviderOption {
  Name: string
  Help: string
  Required: boolean
  IsPassword: boolean
  Advanced: boolean
  Type: string
  Examples?: Array<{ Value: string; Help: string }>
}

interface Provider {
  Name: string
  Description: string
  Prefix: string
  Options: ProviderOption[]
}

const remotes = ref<Record<string, RemoteInfo>>({})
const isLoading = ref(true)

// Wizard State
const isAdding = ref(false)
const isEditing = ref(false)
const step = ref(1)
const providers = ref<Provider[]>([])
const selectedProvider = ref<Provider | null>(null)

// Form State
const formName = ref('')
const formConfig = ref<Record<string, string>>({})
const isSaving = ref(false)
const showAdvanced = ref(false)

const fetchRemotes = async () => {
  isLoading.value = true
  try {
    const res = await callRcloneAPI('config/dump', {})
    remotes.value = res || {}
  } catch (e) {
    console.error('Failed to fetch remotes', e)
  } finally {
    isLoading.value = false
  }
}

const fetchProviders = async () => {
  try {
    const res = await callRcloneAPI('config/providers', {})
    providers.value = res.providers || []
  } catch (e) {
    console.error('Failed to fetch providers', e)
  }
}

const startAdd = async () => {
  isAdding.value = true
  isEditing.value = false
  step.value = 1
  selectedProvider.value = null
  formName.value = ''
  formConfig.value = {}
  showAdvanced.value = false
  if (providers.value.length === 0) {
    await fetchProviders()
  }
}

const editConfig = async (name: string, info: RemoteInfo) => {
  if (providers.value.length === 0) await fetchProviders()
  const provider = providers.value.find(p => p.Prefix === info.type)
  if (provider) {
    selectedProvider.value = provider
    formName.value = name
    const cfg: Record<string, string> = {}
    for (const [k, v] of Object.entries(info)) {
      if (k !== 'type' && typeof v === 'string') cfg[k] = v
    }
    formConfig.value = cfg
    step.value = 2
    isAdding.value = true
    isEditing.value = true
    showAdvanced.value = false
  } else {
    showAlert(t('message.config.providerNotFound'))
    return
  }
}

const deleteConfig = async (name: string) => {
  if (!await showConfirm(t('message.config.deleteConfirm'))) return
  try {
    await callRcloneAPI('config/delete', { name })
    fetchRemotes()
  } catch (e: any) {
    showAlert(`${t('message.config.deleteFailed')}: ${e}`)
  }
}

const selectProvider = (provider: Provider) => {
  selectedProvider.value = provider
  formConfig.value = {}
  step.value = 2
}

const cancelAdd = () => {
  isAdding.value = false
  isEditing.value = false
}

const saveConfig = async () => {
  if (!formName.value) {
    showAlert(t('message.config.nameRequired'))
    return
  }
  isSaving.value = true
  try {
    // Filter out empty parameters
    const params: Record<string, string> = {}
    for (const [k, v] of Object.entries(formConfig.value)) {
      if (v !== '' && v !== undefined) {
        params[k] = v
      }
    }

    if (isEditing.value) {
      await callRcloneAPI('config/update', {
        name: formName.value,
        parameters: params
      })
    } else {
      await callRcloneAPI('config/create', {
        name: formName.value,
        type: selectedProvider.value!.Prefix,
        parameters: params
      })
    }

    showAlert(t('message.config.success'))
    step.value = 1
    fetchRemotes()
    isAdding.value = false
    isEditing.value = false
  } catch (e: any) {
    showAlert(`${t('message.config.saveFailed')}: ${e}`)
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  fetchRemotes()
})
</script>

<template>
  <div class="animate-in fade-in slide-in-from-bottom-4 duration-500">
    <div class="flex items-center justify-between mb-8">
      <h2 class="text-2xl font-semibold tracking-tight">{{ t('message.config.title') }}</h2>
      <button 
        v-if="!isAdding"
        @click="startAdd"
        class="px-4 py-2 bg-black dark:bg-white text-white dark:text-black font-medium rounded-lg hover:bg-gray-800 dark:hover:bg-gray-100 transition-colors shadow-sm text-sm"
      >
        {{ t('message.config.addRemote') }}
      </button>
    </div>

    <!-- View: List Remotes -->
    <div v-if="!isAdding" class="space-y-4">
      <div v-if="isLoading" class="text-gray-500 text-sm py-4">Loading...</div>
      <div v-else-if="Object.keys(remotes).length === 0" class="p-8 bg-white dark:bg-black rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm text-center">
        <p class="text-gray-500 dark:text-gray-400">{{ t('message.config.noRemotes') }}</p>
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div 
          v-for="(info, name) in remotes" 
          :key="name"
          class="p-6 bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group"
        >
          <div class="flex items-start justify-between mb-2">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white truncate pr-4">{{ name }}</h3>
            <span class="px-2.5 py-1 text-xs font-semibold rounded-full bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400 uppercase tracking-wider shrink-0">
              {{ info.type }}
            </span>
          </div>
          <!-- Show 2-3 keys from the config if available -->
          <div class="space-y-1 mt-4 flex-1">
            <div v-for="(v, k) in Object.entries(info).slice(1, 3)" :key="k" class="text-sm truncate">
              <span class="text-gray-400">{{ formatOptionName(v[0]) }}:</span> 
              <span class="text-gray-700 dark:text-gray-300 ml-1 font-mono text-xs">{{ typeof v[1] === 'string' && v[1].length > 20 ? v[1].slice(0,20)+'...' : v[1] }}</span>
            </div>
          </div>
          <!-- Action Buttons -->
          <div class="mt-6 flex items-center justify-end space-x-2 border-t border-gray-100 dark:border-gray-800 pt-4">
            <button @click="editConfig(name, info)" class="px-3 py-1.5 text-sm font-medium text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-md transition-colors">{{ t('message.config.edit') }}</button>
            <button @click="deleteConfig(name)" class="px-3 py-1.5 text-sm font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md transition-colors">{{ t('message.config.delete') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- View: Add Remote Wizard -->
    <div v-else class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
      
      <!-- Step 1: Choose Provider -->
      <div v-if="step === 1" class="p-8">
        <div class="flex justify-between items-center mb-6">
          <h3 class="text-xl font-semibold">{{ t('message.config.step1') }}</h3>
          <button @click="cancelAdd" class="text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">{{ t('message.config.cancel') }}</button>
        </div>
        
        <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 max-h-[60vh] overflow-y-auto p-1">
          <button 
            v-for="provider in providers" 
            :key="provider.Prefix"
            @click="selectProvider(provider)"
            class="flex flex-col p-4 border border-gray-200 dark:border-gray-800 rounded-xl hover:border-blue-500 hover:ring-1 hover:ring-blue-500 hover:bg-blue-50/50 dark:hover:bg-blue-900/10 transition-all text-left"
          >
            <span class="font-bold text-gray-900 dark:text-white mb-1">{{ getProviderName(provider) }}</span>
            <span class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 leading-relaxed" :title="getProviderDescription(provider)">{{ getProviderDescription(provider) }}</span>
          </button>
        </div>
      </div>

      <!-- Step 2: Configure Parameters -->
      <div v-if="step === 2 && selectedProvider" class="p-8">
        <div class="flex justify-between items-center mb-6 border-b border-gray-200 dark:border-gray-800 pb-4">
          <div>
            <h3 class="text-xl font-semibold">{{ t('message.config.step2') }}</h3>
            <p class="text-sm text-gray-500 mt-1">{{ t('message.config.configuring') }} <strong class="text-blue-500">{{ getProviderName(selectedProvider) }}</strong></p>
          </div>
          <button @click="step = 1" class="text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">{{ t('message.config.back') }}</button>
        </div>

        <form @submit.prevent="saveConfig" class="space-y-6 max-w-3xl">
          <!-- Remote Name -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('message.config.nameLabel') }} <span class="text-red-500">*</span></label>
            <input 
              v-model="formName" 
              type="text" 
              required
              :disabled="isEditing"
              class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow disabled:opacity-50 disabled:cursor-not-allowed" 
            />
          </div>

          <!-- Dynamic Options -->
          <div class="space-y-5 border-t border-gray-200 dark:border-gray-800 pt-5">
            <template v-for="opt in selectedProvider.Options" :key="opt.Name">
              <div v-if="!opt.Advanced || showAdvanced">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {{ formatOptionName(opt.Name) }} 
                  <span v-if="opt.Required" class="text-red-500">*</span>
                  <span v-if="opt.Advanced" class="ml-2 text-[10px] uppercase tracking-wide bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400 px-1.5 py-0.5 rounded">Advanced</span>
                </label>
                
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-2 leading-relaxed" v-if="opt.Help">{{ formatOptionHelp(selectedProvider.Prefix, opt) }}</p>
                
                <!-- Select for Examples -->
                <select 
                  v-if="opt.Examples && opt.Examples.length > 0"
                  v-model="formConfig[opt.Name]"
                  :required="opt.Required"
                  class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow text-sm"
                >
                  <option value="" class="dark:bg-gray-800">-- {{ t('message.config.select') || 'Select' }} --</option>
                  <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value" class="dark:bg-gray-800">
                    {{ ex.Value }} {{ ex.Help ? `- ${formatExampleHelp(selectedProvider.Prefix, opt.Name, ex.Value, ex.Help)}` : '' }}
                  </option>
                </select>
                
                <!-- Password Input -->
                <input 
                  v-else-if="opt.IsPassword"
                  v-model="formConfig[opt.Name]"
                  type="password"
                  :required="opt.Required"
                  class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow"
                />

                <!-- Standard Input -->
                <input 
                  v-else
                  v-model="formConfig[opt.Name]"
                  type="text"
                  :required="opt.Required"
                  class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow"
                />
              </div>
            </template>
          </div>

          <div class="flex items-center justify-between pt-6">
            <button type="button" @click="showAdvanced = !showAdvanced" class="text-sm text-blue-600 dark:text-blue-400 font-medium hover:underline">
              {{ showAdvanced ? t('message.config.hideAdvanced') : t('message.config.showAdvanced') }}
            </button>
            <div class="space-x-3">
              <button type="button" @click="cancelAdd" class="px-5 py-2.5 rounded-lg border border-gray-300 dark:border-gray-700 font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
                {{ t('message.config.cancel') }}
              </button>
              <button type="submit" :disabled="isSaving" class="px-5 py-2.5 bg-black dark:bg-white text-white dark:text-black font-medium rounded-lg hover:bg-gray-800 dark:hover:bg-gray-100 transition-colors disabled:opacity-50 min-w-[120px]">
                {{ isSaving ? t('message.config.creating') : t('message.config.save') }}
              </button>
            </div>
          </div>
        </form>
      </div>

    </div>
  </div>
</template>
