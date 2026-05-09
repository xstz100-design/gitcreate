import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { isMobile } from '@/utils/device'
import { isTelegramMiniApp, getInitData } from '@/utils/telegram'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false },
    },
    // 移动端商户路由 (Vant UI)
    {
      path: '/m',
      component: () => import('@/layouts/MobileLayout.vue'),
      meta: { requiresAuth: true, roles: ['merchant', 'picker', 'delivery'], mobile: true },
      children: [
        {
          path: '',
          redirect: '/m/shop',
        },
        {
          path: 'shop',
          name: 'MobileShop',
          component: () => import('@/views/mobile/Shop.vue'),
        },
        {
          path: 'cart',
          name: 'MobileCart',
          component: () => import('@/views/mobile/Cart.vue'),
        },
        {
          path: 'orders',
          name: 'MobileOrders',
          component: () => import('@/views/mobile/Orders.vue'),
        },
        {
          path: 'profile',
          name: 'MobileProfile',
          component: () => import('@/views/mobile/Profile.vue'),
        },
      ],
    },
    // 管理端路由 (Element Plus) - 仅 admin 角色
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true, roles: ['admin'] },
      children: [
        {
          path: '',
          redirect: '/admin/dashboard',
        },
        {
          path: 'dashboard',
          name: 'AdminDashboard',
          component: () => import('@/views/admin/Dashboard.vue'),
        },
        {
          path: 'products',
          name: 'AdminProducts',
          component: () => import('@/views/admin/Products.vue'),
        },
        {
          path: 'orders',
          name: 'AdminOrders',
          component: () => import('@/views/admin/Orders.vue'),
        },
        {
          path: 'settings',
          name: 'AdminSettings',
          component: () => import('@/views/admin/Settings.vue'),
        },

        {
          path: 'merchants',
          name: 'AdminMerchants',
          component: () => import('@/views/admin/Merchants.vue'),
        },
        {
          path: 'approvals',
          redirect: '/admin/merchants',
        },
        {
          path: 'categories',
          name: 'AdminCategories',
          component: () => import('@/views/admin/Categories.vue'),
        },
        {
          path: 'announcements',
          name: 'AdminAnnouncements',
          component: () => import('@/views/admin/Announcements.vue'),
        },
        {
          path: 'profile',
          name: 'AdminProfile',
          component: () => import('@/views/admin/Profile.vue'),
        },
      ],
    },
    // 配送员路由
    {
      path: '/',
      redirect: '/m/shop',
    },
  ],
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  const inTelegram = isTelegramMiniApp()

  // 未登录时
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    // 在 Telegram 环境中，尝试自动登录
    if (inTelegram) {
      try {
        await userStore.telegramLogin(getInitData())
        // 登录成功，管理员使用管理端界面，商户使用移动端界面
        if (userStore.isAdmin && !to.path.startsWith('/admin')) {
          next('/admin/dashboard')
          return
        }
        // 商户继续导航
      } catch (e) {
        console.error('TG 自动登录失败:', e)
        next('/login')
        return
      }
    } else {
      next('/login')
      return
    }
  }

  // 已登录访问 /login → 按角色重定向
  if (to.path === '/login' && userStore.isLoggedIn) {
    if (userStore.isAdmin) {
      next('/admin/dashboard')
    } else if (userStore.isMerchant) {
      next('/m/shop')
    } else {
      next('/m/shop')
    }
    return
  }

  // 在 Telegram 中访问 /login → 尝试自动登录后重定向
  if (to.path === '/login' && inTelegram && !userStore.isLoggedIn) {
    try {
      await userStore.telegramLogin(getInitData())
      if (userStore.isAdmin) {
        next('/admin/dashboard')
      } else {
        next('/m/shop')
      }
      return
    } catch {
      // 登录失败，继续显示登录页
    }
  }

  // 角色不匹配 → 重定向
  // 支持 meta.role (单角色) 和 meta.roles (多角色数组)
  const allowedRoles = to.meta.roles || (to.meta.role ? [to.meta.role] : null)
  if (allowedRoles && !allowedRoles.includes(userStore.userRole)) {
    // 管理员在 Telegram miniapp 中可以访问 /m/* 移动端路由
    if (userStore.isAdmin && inTelegram && (to.path.startsWith('/m') || to.path === '/')) {
      next()
      return
    }
    if (userStore.isAdmin) {
      next('/admin/dashboard')
    } else if (userStore.isMerchant) {
      next('/m/shop')
    } else {
      next('/login')
    }
    return
  }

  next()
})

export default router
