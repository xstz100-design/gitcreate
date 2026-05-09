<template>
  <!-- 桌面端外层容器：根据环境自适应 -->
  <div class="desktop-wrapper" :class="{ 'tg-frame-mode': isTgContext }">
    <div class="mobile-layout" :class="{ 'has-shell-header': showShellHeader, 'tg-frame-mode': isTgContext }">
      <div v-if="showShellHeader" class="mobile-header">
        <button class="lang-switch" @click="toggleLang">
          {{ langLabel }}
        </button>
      </div>

      <!-- 内容区（可滚动） -->
      <div class="mobile-content">
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>

      <!-- 底部导航栏（始终固定在底部） -->
      <van-tabbar v-model="activeTab" @change="handleTabChange" :fixed="false" :placeholder="false">
        <van-tabbar-item icon="shop-o" to="/m/shop">{{ $t('nav.shop') }}</van-tabbar-item>
        <van-tabbar-item icon="shopping-cart-o" :badge="cartStore.totalCount || ''">
          {{ $t('nav.cart') }}
        </van-tabbar-item>
        <van-tabbar-item icon="orders-o" to="/m/orders">{{ $t('nav.orders') }}</van-tabbar-item>
        <van-tabbar-item icon="user-o" to="/m/profile">{{ $t('nav.profile') }}</van-tabbar-item>
      </van-tabbar>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { hapticFeedback } from '@/utils/device'
import { setLanguage, getCurrentLanguage } from '@/i18n'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const cartStore = useCartStore()

const activeTab = ref(0)
const currentLang = ref(getCurrentLanguage())

// 检测是否在 Telegram 环境中（移动端或桌面端 Telegram）
// Telegram Mini App 会注入 window.Telegram.WebApp，或 UA 含 "Telegram"
const isTgContext = ref(
  typeof window !== 'undefined' && (
    !!(window.Telegram?.WebApp?.initData) ||
    !!(window.TelegramWebviewProxy) ||
    /Telegram/i.test(navigator.userAgent)
  )
)

const routesWithPageNavbar = ['/m/cart', '/m/orders', '/m/profile']
const showShellHeader = computed(() => !routesWithPageNavbar.some(prefix => route.path.startsWith(prefix)))

const toggleLang = () => {
  const order = ['zh', 'en', 'kh']
  const idx = order.indexOf(currentLang.value)
  const newLang = order[(idx + 1) % order.length]
  setLanguage(newLang)
  currentLang.value = newLang
  hapticFeedback('light')
}

const langLabel = computed(() => {
  const next = { zh: 'EN', en: 'ខ្មែរ', kh: '中' }
  return next[currentLang.value] || 'EN'
})

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
/* ── 外层容器 ── */
.desktop-wrapper {
  height: 100vh;
  height: 100dvh;
  background: #f0f2f5;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  overflow: hidden;
}

/* ── 主容器（所有设备） ── */
.mobile-layout {
  width: 100%;
  max-width: 520px;
  height: 100vh;
  height: 100dvh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 0 24px rgba(0, 0, 0, 0.12);
}

/* ── 内容滚动区 ── */
.mobile-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior-y: contain;
  /* 必须有 position + z-index，否则 iOS/Telegram 中 van-tabbar(z-index:1)
     会压盖内容区的所有固定定位子元素（弹窗/结算栏等） */
  position: relative;
  z-index: 2;
}

/* ════════════════════════════════════════
   桌面端 (≥600px) — 分两种模式
   ════════════════════════════════════════ */

/* 模式1：电脑 Telegram → 仿手机框 */
@media (min-width: 600px) {
  .desktop-wrapper.tg-frame-mode {
    background: linear-gradient(160deg, #0d1b2a 0%, #1b2d3e 45%, #0f3460 100%);
  }
  .mobile-layout.tg-frame-mode {
    max-width: 520px;
    height: calc(100vh - 20px);
    height: calc(100dvh - 20px);
    margin-top: 20px;
    border-radius: 32px 32px 0 0;
    overflow: hidden;
    box-shadow:
      0 0 0 6px rgba(255, 255, 255, 0.07),
      0 0 0 9px rgba(255, 255, 255, 0.04),
      0 32px 80px rgba(0, 0, 0, 0.55);
  }
}

/* 模式2：电脑浏览器 → 响应式宽屏 */
@media (min-width: 600px) {
  .desktop-wrapper:not(.tg-frame-mode) {
    background: #eef0f3;
  }
  .mobile-layout:not(.tg-frame-mode) {
    max-width: 860px;
    border-radius: 0;
    box-shadow: 0 0 0 1px rgba(0,0,0,0.06), 0 4px 24px rgba(0,0,0,0.08);
  }
}

@media (min-width: 1100px) {
  .mobile-layout:not(.tg-frame-mode) {
    max-width: 960px;
  }
}

/* ══ 顶部 header（flex 子项，自然置顶） ══ */
.mobile-header {
  flex-shrink: 0;
  height: 46px;
  background: #1d4ed8;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: none;
  box-shadow: 0 2px 8px rgba(29, 78, 216, 0.3);
  z-index: 100;
  position: relative;
}

.mobile-logo {
  height: 28px;
  width: auto;
  object-fit: contain;
  filter: brightness(0) invert(1);
}

.lang-switch {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 24px;
  border: 1px solid rgba(255,255,255,0.4);
  border-radius: 4px;
  background: rgba(255,255,255,0.15);
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.lang-switch:active {
  background: rgba(255,255,255,0.3);
}

/* ── 底部导航栏（始终在底部，不需要 fixed） ── */
:deep(.van-tabbar) {
  flex-shrink: 0;
  width: 100%;
  position: relative;
}
</style>
