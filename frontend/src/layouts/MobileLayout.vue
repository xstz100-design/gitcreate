<template>
  <div class="mobile-layout">
    <div class="mobile-header">
      <img src="/images/logo2.svg" alt="Logo" class="mobile-logo" />
      <button class="lang-switch" @click="toggleLang">
        {{ currentLang === 'zh' ? 'EN' : '中' }}
      </button>
    </div>
    
    <router-view v-slot="{ Component }">
      <keep-alive>
        <component :is="Component" />
      </keep-alive>
    </router-view>
    
    <!-- 底部导航栏 -->
    <van-tabbar v-model="activeTab" @change="handleTabChange" :placeholder="true">
      <van-tabbar-item icon="shop-o" to="/m/shop">{{ $t('nav.shop') }}</van-tabbar-item>
      <van-tabbar-item icon="shopping-cart-o" :badge="cartStore.totalCount || ''">
        {{ $t('nav.cart') }}
      </van-tabbar-item>
      <van-tabbar-item icon="orders-o" to="/m/orders">{{ $t('nav.orders') }}</van-tabbar-item>
      <van-tabbar-item icon="user-o" to="/m/profile">{{ $t('nav.profile') }}</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { hapticFeedback } from '@/utils/device'
import { setLanguage, getCurrentLanguage } from '@/i18n'

const router = useRouter()
const route = useRoute()
const cartStore = useCartStore()

const activeTab = ref(0)
const currentLang = ref(getCurrentLanguage())

const toggleLang = () => {
  const newLang = currentLang.value === 'zh' ? 'en' : 'zh'
  setLanguage(newLang)
  currentLang.value = newLang
  hapticFeedback('light')
}

// 根据路由更新选中状态
watch(() => route.path, (newPath) => {
  if (newPath.includes('/m/shop')) activeTab.value = 0
  else if (newPath.includes('/m/cart')) activeTab.value = 1
  else if (newPath.includes('/m/orders')) activeTab.value = 2
  else if (newPath.includes('/m/profile')) activeTab.value = 3
}, { immediate: true })

const handleTabChange = (index) => {
  hapticFeedback('light')
  const routes = ['/m/shop', '/m/cart', '/m/orders', '/m/profile']
  router.push(routes[index])
}
</script>

<style scoped>
.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 46px;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: 1px solid #eee;
  z-index: 999;
}

.mobile-logo {
  height: 28px;
  width: auto;
  object-fit: contain;
}

.lang-switch {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 24px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: #f5f5f5;
  color: #333;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.lang-switch:active {
  background: #e0e0e0;
}

.mobile-layout {
  min-height: 100vh;
  padding-top: 46px;
  background: #f5f5f5;
  padding-bottom: 50px; /* 为底部导航留空间 */
}
</style>
