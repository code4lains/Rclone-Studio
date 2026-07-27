import { createI18n } from 'vue-i18n'
import en from '../locales/en.json'
import zh from '../locales/zh.json'

const messages = {
  en,
  zh,
}

const savedLang = localStorage.getItem('language') || 'zh'

const i18n = createI18n({
  legacy: false,
  locale: savedLang,
  fallbackLocale: 'en',
  messages,
})

export function setLanguage(lang: string) {
  if (i18n.global.locale.value !== lang) {
    i18n.global.locale.value = lang as any
    localStorage.setItem('language', lang)
  }
}

export default i18n
