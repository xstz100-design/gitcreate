import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { isMobile } from '@/utils/device'

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
      meta: { requiresAuth: true, role: 'merchant', mobile: true },
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
    // PC端商户路由 (Element Plus)
    {
      path: '/merchant',
      component: () => import('@/layouts/MerchantLayout.vue'),
      meta: { requiresAuth: true, role: 'merchant' },
      children: [
        {
          path: '',
          redirect: '/merchant/products',
        },
        {
          path: 'products',
          name: 'MerchantProducts',
          component: () => import('@/views/merchant/Products.vue'),
        },
        {
          path: 'cart',
          name: 'MerchantCart',
          component: () => import('@/views/merchant/Cart.vue'),
        },
        {
          path: 'orders',
          name: 'MerchantOrders',
          component: () => import('@/views/merchant/Orders.vue'),
        },
        {
          path: 'profile',
          name: 'MerchantProfile',
          component: () => import('@/views/merchant/Profile.vue'),
        },
      ],
    },
    // 管理端路由 (Element Plus)
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true, role: 'admin' },
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
          path: 'merchants',
          name: 'AdminMerchants',
          component: () => import('@/views/admin/Merchants.vue'),
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
    {
      path: '/',
      redirect: '/login',
    },
  ],
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
    return
  }

  if (to.path === '/login' && userStore.isLoggedIn) {
    if (userStore.isAdmin) {
      next('/admin/dashboard')
    } else if (userStore.isMerchant) {
      next(isMobile() ? '/m/shop' : '/merchant/products')
    } else {
      next('/')
    }
    return
  }

  if (to.meta.role && userStore.userRole !== to.meta.role) {
    if (userStore.isAdmin) {
      next('/admin/dashboard')
    } else if (userStore.isMerchant) {
      next(isMobile() ? '/m/shop' : '/merchant/products')
    } else {
      next('/login')
    }
    return
  }

  if ((to.path === '/merchant' || to.path.startsWith('/merchant/')) && isMobile() && userStore.isMerchant) {
    // 路由映射: PC路径 → 移动路径
    const routeMap = {
      '/merchant': '/m/shop',
      '/merchant/products': '/m/shop',
      '/merchant/cart': '/m/cart',
      '/merchant/orders': '/m/orders',
      '/merchant/profile': '/m/profile',
    }
    const mobilePath = routeMap[to.path] || '/m/shop'
    next(mobilePath)
    return
  }

  if ((to.path === '/m' || to.path.startsWith('/m/')) && !isMobile() && userStore.isMerchant) {
    // 路由映射: 移动路径 → PC路径
    const routeMap = {
      '/m': '/merchant/products',
      '/m/shop': '/merchant/products',
      '/m/cart': '/merchant/cart',
      '/m/orders': '/merchant/orders',
      '/m/profile': '/merchant/profile',
    }
    const pcPath = routeMap[to.path] || '/merchant/products'
    next(pcPath)
    return
  }

  next()
})

export default router
