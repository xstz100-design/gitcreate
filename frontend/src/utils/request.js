import axios from 'axios'
import { ElMessage } from 'element-plus/es/components/message/index'
import { useUserStore } from '@/stores/user'
import i18n from '@/i18n'

const request = axios.create({
  baseURL: '',
  timeout: 30000,
})

let _userStore = null
const getUserStore = () => {
  if (!_userStore) _userStore = useUserStore()
  return _userStore
}

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const store = getUserStore()
    if (store.token) {
      config.headers.Authorization = `Bearer ${store.token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.response?.data?.detail || error.message || i18n.global.t('common.requestFailed')

    if (error.response?.status === 401 && !error.config?.url?.includes('/api/auth/login')) {
      getUserStore().logout()
    }

    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export default request
