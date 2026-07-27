<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { callRcloneAPI } from '../utils/api'
import { useConnectionStore } from '../stores'
import { showAlert, showConfirm, showPrompt } from '../composables/useModal'

const { t } = useI18n()
const connectionStore = useConnectionStore()

interface RcloneItem {
  Path: string
  Name: string
  Size: number
  MimeType: string
  ModTime: string
  IsDir: boolean
}

const remotes = ref<string[]>([])
const selectedRemote = ref('')
const currentPath = ref<string[]>([])
const files = ref<RcloneItem[]>([])
const isLoading = ref(false)
const searchQuery = ref('')
const previewModal = ref(false)
const previewItem = ref<RcloneItem | null>(null)
const previewUrl = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const getFs = () => selectedRemote.value.endsWith(':') ? selectedRemote.value : `${selectedRemote.value}:`
const getRemotePath = (name: string) => currentPath.value.length > 0 ? currentPath.value.join('/') + '/' + name : name

const getFileUrl = (fs: string, path: string) => {
  const c = connectionStore.config
  const auth = (c.user || c.pass) ? `${encodeURIComponent(c.user)}:${encodeURIComponent(c.pass)}@` : ''
  return `http://${auth}${c.ip}:${c.port}/%5B${encodeURIComponent(fs)}%5D/${path.split('/').map(encodeURIComponent).join('/')}`
}

const isImage = (item: RcloneItem | null) => {
  if (!item) return false
  if (item.MimeType?.startsWith('image/')) return true
  const ext = item.Name.split('.').pop()?.toLowerCase() || ''
  return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'ico'].includes(ext)
}

const isVideo = (item: RcloneItem | null) => {
  if (!item) return false
  if (item.MimeType?.startsWith('video/')) return true
  const ext = item.Name.split('.').pop()?.toLowerCase() || ''
  return ['mp4', 'webm', 'ogg', 'mov', 'mkv', 'avi', 'm4v'].includes(ext)
}

const previewFile = (item: RcloneItem) => {
  previewItem.value = item
  previewUrl.value = getFileUrl(getFs(), getRemotePath(item.Name))
  previewModal.value = true
}

const downloadFile = (item: RcloneItem) => {
  window.open(getFileUrl(getFs(), getRemotePath(item.Name)), '_blank')
}

const deleteFile = async (item: RcloneItem) => {
  if (!await showConfirm(t('message.files.confirmDelete', { name: item.Name }))) return
  try {
    if (item.IsDir) {
      await callRcloneAPI('operations/purge', { fs: getFs(), remote: getRemotePath(item.Name) })
    } else {
      await callRcloneAPI('operations/deletefile', { fs: getFs(), remote: getRemotePath(item.Name) })
    }
    fetchFiles()
  } catch(e) { showAlert('Delete failed: ' + e) }
}

const renameFile = async (item: RcloneItem) => {
  const newName = await showPrompt(t('message.files.renameTitle'), t('message.files.newName'), item.Name)
  if (newName && newName !== item.Name) {
    try {
      if (item.IsDir) {
        await callRcloneAPI('sync/move', { 
          srcFs: getFs() + getRemotePath(item.Name), 
          dstFs: getFs() + getRemotePath(newName) 
        })
      } else {
        await callRcloneAPI('operations/movefile', {
          srcFs: getFs(),
          srcRemote: getRemotePath(item.Name),
          dstFs: getFs(),
          dstRemote: getRemotePath(newName)
        })
      }
      fetchFiles()
    } catch(e) { showAlert('Rename failed: ' + e) }
  }
}

const copyFile = async (item: RcloneItem) => {
  const dst = await showPrompt(t('message.files.copyTitle'), t('message.files.destinationPath'), getRemotePath(item.Name))
  if (dst) {
    try {
      if (item.IsDir) {
        await callRcloneAPI('sync/copy', { 
          srcFs: getFs() + getRemotePath(item.Name), 
          dstFs: getFs() + dst 
        })
      } else {
        await callRcloneAPI('operations/copyfile', {
          srcFs: getFs(),
          srcRemote: getRemotePath(item.Name),
          dstFs: getFs(),
          dstRemote: dst
        })
      }
      fetchFiles()
    } catch(e) { showAlert('Copy failed: ' + e) }
  }
}

const moveFile = async (item: RcloneItem) => {
  const dst = await showPrompt(t('message.files.moveTitle'), t('message.files.destinationPath'), getRemotePath(item.Name))
  if (dst) {
    try {
      if (item.IsDir) {
        await callRcloneAPI('sync/move', { 
          srcFs: getFs() + getRemotePath(item.Name), 
          dstFs: getFs() + dst 
        })
      } else {
        await callRcloneAPI('operations/movefile', {
          srcFs: getFs(),
          srcRemote: getRemotePath(item.Name),
          dstFs: getFs(),
          dstRemote: dst
        })
      }
      fetchFiles()
    } catch(e) { showAlert('Move failed: ' + e) }
  }
}

const newFolder = async () => {
  const name = await showPrompt(t('message.files.newFolderTitle'), t('message.files.newFolderName'), '')
  if (name) {
    try {
      await callRcloneAPI('operations/mkdir', { fs: getFs(), remote: getRemotePath(name) })
      fetchFiles()
    } catch(e) { showAlert('Mkdir failed: ' + e) }
  }
}

const handleUpload = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return
  const file = input.files[0]
  
  const fs = getFs()
  const remote = getRemotePath(file.name)

  const c = connectionStore.config
  const url = `http://${c.ip}:${c.port}/operations/uploadfile?fs=${encodeURIComponent(fs)}&remote=${encodeURIComponent(remote)}`
  
  const formData = new FormData()
  formData.append('file', file)

  const headers = new Headers()
  if (c.user || c.pass) {
    headers.set('Authorization', 'Basic ' + btoa(`${c.user}:${c.pass}`))
  }

  try {
    isLoading.value = true
    const res = await fetch(url, {
      method: 'POST',
      headers,
      body: formData
    })
    if (!res.ok) throw new Error('Upload failed with status ' + res.status)
    fetchFiles()
  } catch (e) {
    showAlert('Upload failed: ' + e)
  } finally {
    isLoading.value = false
    input.value = ''
  }
}

const breadcrumbs = computed(() => {
  const crumbs = [{ name: t('message.files.root'), path: [] as string[] }]
  let builtPath: string[] = []
  currentPath.value.forEach(part => {
    builtPath = [...builtPath, part]
    crumbs.push({ name: part, path: builtPath })
  })
  return crumbs
})

const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  const lower = searchQuery.value.toLowerCase()
  return files.value.filter(f => f.Name.toLowerCase().includes(lower))
})

const fetchRemotes = async () => {
  try {
    const res = await callRcloneAPI('config/listremotes', {})
    remotes.value = res.remotes || []
    if (remotes.value.length > 0 && !selectedRemote.value) {
      selectedRemote.value = remotes.value[0]
    }
  } catch (e) {
    console.error('Failed to fetch remotes', e)
  }
}

const fetchFiles = async () => {
  if (!selectedRemote.value) return
  isLoading.value = true
  files.value = []
  try {
    // Note: rclone operations/list needs fs like "remote:" and remote as "path"
    // config/listremotes returns names with or without ':', typically without but rclone 
    // requires appending ':' for the fs param if it doesn't have it.
    const fs = getFs()
    const pathStr = currentPath.value.join('/')
    
    const res = await callRcloneAPI('operations/list', {
      fs: fs,
      remote: pathStr
    })
    
    // Sort directories first, then alphabetical
    let list: RcloneItem[] = res.list || []
    list.sort((a, b) => {
      if (a.IsDir && !b.IsDir) return -1
      if (!a.IsDir && b.IsDir) return 1
      return a.Name.localeCompare(b.Name)
    })
    
    files.value = list
  } catch (e) {
    console.error('Failed to fetch files', e)
    showAlert('Failed to load directory.')
  } finally {
    isLoading.value = false
  }
}

watch(selectedRemote, () => {
  currentPath.value = []
  fetchFiles()
})

const openFolder = (item: RcloneItem) => {
  if (item.IsDir) {
    currentPath.value = [...currentPath.value, item.Name]
    fetchFiles()
  }
}

const navigateTo = (pathArr: string[]) => {
  currentPath.value = pathArr
  fetchFiles()
}

const formatBytes = (bytes: number) => {
  if (bytes < 0) return '-'
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString()
}

onMounted(() => {
  fetchRemotes()
})
</script>

<template>
  <div class="animate-in fade-in slide-in-from-bottom-4 duration-500">
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-2xl font-semibold tracking-tight">{{ t('message.files.title') }}</h2>
      
      <!-- Remote Selector & Upload & Search -->
      <div class="flex items-center space-x-4">
        <div class="relative w-48 hidden sm:block">
          <input 
            type="text" 
            v-model="searchQuery"
            :placeholder="t('message.files.searchPlaceholder')" 
            class="w-full pl-9 pr-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-black focus:outline-none focus:ring-2 focus:ring-blue-500 shadow-sm text-sm"
          />
          <svg class="w-4 h-4 text-gray-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
        </div>
        <div class="w-48">
          <select 
            v-model="selectedRemote"
            class="w-full px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-black focus:outline-none focus:ring-2 focus:ring-blue-500 shadow-sm text-sm"
          >
            <option value="" disabled>{{ remotes.length === 0 ? t('message.files.noRemotes') : t('message.files.selectRemote') }}</option>
            <option v-for="r in remotes" :key="r" :value="r">{{ r }}</option>
          </select>
        </div>
        <button v-if="selectedRemote" @click="newFolder" class="px-4 py-2 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 font-medium rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors shadow-sm text-sm shrink-0">
          {{ t('message.files.newFolder') }}
        </button>
        <button v-if="selectedRemote" @click="fileInput?.click()" class="px-4 py-2 bg-black dark:bg-white text-white dark:text-black font-medium rounded-lg hover:bg-gray-800 dark:hover:bg-gray-100 transition-colors shadow-sm text-sm shrink-0">
          {{ t('message.files.upload') }}
        </button>
        <input type="file" ref="fileInput" class="hidden" @change="handleUpload" />
      </div>
    </div>

    <!-- Main Card -->
    <div class="bg-white dark:bg-black rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col h-[calc(100vh-14rem)]">
      
      <!-- Breadcrumb Header -->
      <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/30 shrink-0">
        <div class="flex items-center space-x-2 text-sm">
          <template v-for="(crumb, idx) in breadcrumbs" :key="idx">
            <button 
              @click="navigateTo(crumb.path)"
              class="hover:text-blue-600 dark:hover:text-blue-400 font-medium transition-colors"
              :class="idx === breadcrumbs.length - 1 ? 'text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400'"
            >
              {{ crumb.name }}
            </button>
            <span v-if="idx < breadcrumbs.length - 1" class="text-gray-400">/</span>
          </template>
        </div>
      </div>

      <!-- File Table -->
      <div class="flex-1 overflow-y-auto">
        <div v-if="isLoading" class="p-8 text-center text-gray-500 text-sm">
          {{ t('message.files.loading') }}
        </div>
        <div v-else-if="files.length === 0" class="p-12 text-center text-gray-500">
          <div class="w-16 h-16 mx-auto bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mb-4">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path></svg>
          </div>
          {{ t('message.files.emptyDir') }}
        </div>
        <table v-else class="w-full text-left border-collapse">
          <thead class="sticky top-0 bg-white dark:bg-black shadow-sm z-10">
            <tr class="border-b border-gray-200 dark:border-gray-800">
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider w-12"></th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ t('message.files.name') }}</th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider w-32">{{ t('message.files.size') }}</th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider w-48 hidden sm:table-cell">{{ t('message.files.modTime') }}</th>
              <th class="px-6 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider text-right w-32">{{ t('message.files.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr 
              v-for="item in filteredFiles" 
              :key="item.Path"
              @click="openFolder(item)"
              class="hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors"
              :class="{ 'cursor-pointer': item.IsDir }"
            >
              <td class="px-6 py-4">
                <svg v-if="item.IsDir" class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"></path></svg>
                <svg v-else class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>
              </td>
              <td class="px-6 py-4 font-medium text-gray-900 dark:text-gray-100 text-sm">
                {{ item.Name }}
              </td>
              <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                {{ item.IsDir ? '-' : formatBytes(item.Size) }}
              </td>
              <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 hidden sm:table-cell">
                {{ formatDate(item.ModTime) }}
              </td>
              <td class="px-6 py-4 text-right text-sm font-medium relative">
                <div class="group inline-block relative">
                  <button class="px-3 py-1.5 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors inline-flex items-center space-x-1">
                    <span>{{ t('message.files.actions') }}</span>
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                  </button>
                  <!-- Dropdown Menu -->
                  <div class="absolute right-0 mt-1 w-32 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-20 overflow-hidden text-left flex flex-col py-1">
                    <button v-if="!item.IsDir && (isImage(item) || isVideo(item))" @click.stop="previewFile(item)" class="w-full px-4 py-2 text-sm text-blue-600 hover:bg-gray-100 dark:hover:bg-gray-700 text-left transition-colors">{{ t('message.files.preview') }}</button>
                    <button v-if="!item.IsDir" @click.stop="downloadFile(item)" class="w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 text-left transition-colors">{{ t('message.files.download') }}</button>
                    <button @click.stop="renameFile(item)" class="w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 text-left transition-colors">{{ t('message.files.rename') }}</button>
                    <button @click.stop="copyFile(item)" class="w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 text-left transition-colors">{{ t('message.files.copy') }}</button>
                    <button @click.stop="moveFile(item)" class="w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 text-left transition-colors">{{ t('message.files.move') }}</button>
                    <button @click.stop="deleteFile(item)" class="w-full px-4 py-2 text-sm text-red-600 hover:bg-gray-100 dark:hover:bg-gray-700 text-left transition-colors">{{ t('message.files.delete') }}</button>
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </div>

    <!-- Preview Modal -->
    <div v-if="previewModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4" @click.self="previewModal = false">
      <div class="bg-transparent relative max-w-5xl w-full max-h-[90vh] flex flex-col items-center justify-center">
        <button @click="previewModal = false" class="absolute -top-10 right-0 text-white hover:text-gray-300 font-medium tracking-wide flex items-center space-x-1">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          <span>Close</span>
        </button>
        <img v-if="isImage(previewItem)" :src="previewUrl" class="max-w-full max-h-full rounded-lg shadow-2xl object-contain animate-in zoom-in-95 duration-300" />
        <video v-else-if="isVideo(previewItem)" :src="previewUrl" controls autoplay class="max-w-full max-h-full rounded-lg shadow-2xl outline-none bg-black animate-in zoom-in-95 duration-300"></video>
      </div>
    </div>
  </div>
</template>
