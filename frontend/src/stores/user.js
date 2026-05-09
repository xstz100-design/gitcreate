import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, getCurrentUser, telegramAuth as apiTelegramAuth } from '@/api'

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

    const isSuperAdmin = computed(() => !!userInfo.value?.is_super_admin)

    // 是否是商户（picker/delivery 均归为商户处理）
    const isMerchant = computed(() => ['merchant', 'picker', 'delivery'].includes(userRole.value))

    // 资料是否完整
    const profileCompleted = computed(() => !!userInfo.value?.profile_completed)

    // 审核状态
    const approvalStatus = computed(() => userInfo.value?.approval_status || '')

    // 是否已审核通过
    const isApproved = computed(() => approvalStatus.value === 'approved')

    // 下单受限状态：资料未完善 / 审核中 / 审核拒绝 / 已通过
    const orderAccessState = computed(() => {
      if (!isMerchant.value) return 'approved'
      if (!profileCompleted.value) return 'incomplete'
      if (approvalStatus.value === 'rejected') return 'rejected'
      if (approvalStatus.value === 'pending') return 'pending'
      if (approvalStatus.value === 'approved') return 'approved'
      return 'pending'
    })

    // 能否下单（资料完整 + 审核通过）
    const canOrder = computed(() => profileCompleted.value && isApproved.value)

    // 登录
    async function login(username, password) {
      const data = await apiLogin(username, password)
      token.value = data.access_token
      userInfo.value = data.user
      return data
    }

    // Telegram Mini App 免登录
    async function telegramLogin(initData) {
      const data = await apiTelegramAuth(initData)
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
      isSuperAdmin,
      isMerchant,
      profileCompleted,
      approvalStatus,
      isApproved,
      orderAccessState,
      canOrder,
      login,
      telegramLogin,
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
