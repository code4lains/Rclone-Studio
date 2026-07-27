import { ref } from 'vue'

export const alertState = ref({
  isOpen: false,
  message: '',
  resolve: null as ((value: void) => void) | null
})

export const confirmState = ref({
  isOpen: false,
  message: '',
  resolve: null as ((value: boolean) => void) | null
})

export const promptState = ref({
  isOpen: false,
  title: '',
  label: '',
  defaultValue: '',
  value: '',
  resolve: null as ((value: string | null) => void) | null
})

export const showAlert = (message: string): Promise<void> => {
  return new Promise(resolve => {
    alertState.value = { isOpen: true, message, resolve }
  })
}

export const showConfirm = (message: string): Promise<boolean> => {
  return new Promise(resolve => {
    confirmState.value = { isOpen: true, message, resolve }
  })
}

export const showPrompt = (title: string, label: string, defaultValue: string = ''): Promise<string | null> => {
  return new Promise(resolve => {
    promptState.value = { isOpen: true, title, label, defaultValue, value: defaultValue, resolve }
  })
}
