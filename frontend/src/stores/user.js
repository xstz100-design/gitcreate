import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, getCurrentUser } from '@/api'

export const useUserStore = defineStore(
  'user',
  () => {
    const token = ref('')
    const userInfo = ref(null)

    // 是否已登录（简化逻辑，避免多次计算）
    const isLoggedIn = computed(() => {
      return !!token.value && !!userInfo.value
    })

    // 用户角色
    const userRole = computed(() => userInfo.value?.role || '')

    // 是否是管理员
    const isAdmin = computed(() => userRole.value === 'admin')

    // 是否是商户
    const isMerchant = computed(() => userRole.value === 'merchant')

    // 登录
    async function login(username, password) {
      const data = await apiLogin(username, password)
      token.value = data.access_token
      userInfo.value = data.user
      return data
    }

    // 登出
    function logout() {
      token.value = ''
      userInfo.value = null
    }

    // 获取用户信息
    async function fetchUserInfo() {
      if (!token.value) return
      try {
        const data = await getCurrentUser()
        userInfo.value = data
      } catch (error) {
        console.error('获取用户信息失败:', error)
        logout()
      }
    }

    return {
      token,
      userInfo,
      isLoggedIn,
      userRole,
      isAdmin,
      isMerchant,
      login,
      logout,
      fetchUserInfo,
    }
  },
  {
    persist: {
      key: 'user',
      storage: localStorage,
      pick: ['token', 'userInfo'],
    },
  }
)
