import { createI18n } from 'vue-i18n'
import zh from './zh'
import en from './en'
import kh from './kh'

const savedLang = localStorage.getItem('app-lang') || 'zh'

const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: savedLang,
  fallbackLocale: 'zh',
  messages: { zh, en, kh },
})

export default i18n

export function setLanguage(lang) {
  i18n.global.locale.value = lang
  localStorage.setItem('app-lang', lang)
}

export function getCurrentLanguage() {
  return i18n.global.locale.value
}
