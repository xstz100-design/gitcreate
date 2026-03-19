<template>
  <el-container class="merchant-layout" :class="{ 'is-mobile': mobile }">
    <el-header v-if="!mobile">
      <div class="header-content">
        <div class="header-left">
          <img src="/images/logo2.svg" alt="Logo" class="header-logo" />
        </div>
        <div class="header-right">
          <button class="lang-btn" @click="toggleLang">{{ currentLang === 'zh' ? 'EN' : '中文' }}</button>
          <el-badge :value="cartStore.totalCount" :hidden="cartStore.totalCount === 0">
            <el-button @click="$router.push('/merchant/cart')" :icon="ShoppingCart">
              {{ $t('nav.cart') }}
            </el-button>
          </el-badge>
          <el-dropdown @command="handleCommand">
            <span class="user-dropdown">
              {{ userStore.userInfo?.full_name }}
              <el-icon><arrow-down /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">{{ $t('profile.logout') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>

    <!-- 移动端 header -->
    <div v-if="mobile" class="mobile-header">
      <img src="/images/logo2.svg" alt="Logo" class="header-logo" />
      <button class="lang-btn" @click="toggleLang">{{ currentLang === 'zh' ? 'EN' : '中文' }}</button>
    </div>
    
    <el-container>
      <el-aside v-if="!mobile" width="200px">
        <el-menu
          :default-active="$route.path"
          router
          class="merchant-menu"
        >
          <el-menu-item index="/merchant/products">
            <el-icon><goods /></el-icon>
            <span>{{ $t('nav.products') }}</span>
          </el-menu-item>
          <el-menu-item index="/merchant/orders">
            <el-icon><list /></el-icon>
            <span>{{ $t('nav.myOrders') }}</span>
          </el-menu-item>
          <el-menu-item index="/merchant/profile">
            <el-icon><user /></el-icon>
            <span>{{ $t('nav.profile') }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-main :style="mobile ? { 'padding-bottom': '60px' } : {}">
        <router-view />
      </el-main>
    </el-container>

    <!-- 移动端底部TabBar -->
    <div v-if="mobile" class="mobile-tabbar">
      <div class="tab-item" :class="{ active: $route.path === '/merchant/products' }" @click="$router.push('/merchant/products')">
        <el-icon :size="20"><Goods /></el-icon>
        <span>{{ $t('nav.shop') }}</span>
      </div>
      <div class="tab-item" :class="{ active: $route.path === '/merchant/cart' }" @click="$router.push('/merchant/cart')">
        <div class="tab-icon-wrap">
          <el-icon :size="20"><ShoppingCart /></el-icon>
          <span v-if="cartStore.totalCount > 0" class="tab-badge">{{ cartStore.totalCount }}</span>
        </div>
        <span>{{ $t('nav.cart') }}</span>
      </div>
      <div class="tab-item" :class="{ active: $route.path === '/merchant/orders' }" @click="$router.push('/merchant/orders')">
        <el-icon :size="20"><List /></el-icon>
        <span>{{ $t('nav.orders') }}</span>
      </div>
      <div class="tab-item" :class="{ active: $route.path === '/merchant/profile' }" @click="$router.push('/merchant/profile')">
        <el-icon :size="20"><User /></el-icon>
        <span>{{ $t('nav.profile') }}</span>
      </div>
    </div>
  </el-container>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { ShoppingCart, ArrowDown, Goods, List, User } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { useI18n } from 'vue-i18n'
import { setLanguage, getCurrentLanguage } from '@/i18n'

const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()
const { t } = useI18n()
const currentLang = ref(getCurrentLanguage())
const mobile = ref(window.innerWidth < 768)

const onResize = () => {
  mobile.value = window.innerWidth < 768
}

onMounted(() => {
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
})

const toggleLang = () => {
  const newLang = currentLang.value === 'zh' ? 'en' : 'zh'
  setLanguage(newLang)
  currentLang.value = newLang
}

const handleCommand = (command) => {
  if (command === 'logout') {
    ElMessageBox.confirm(t('merchant.logoutConfirm'), t('merchant.hint'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }).then(() => {
      userStore.logout()
      cartStore.clear()
      router.push('/login')
    })
  }
}
</script>

<style scoped>
.merchant-layout {
  height: 100vh;
}

.el-header {
  background: #409eff;
  color: white;
  display: flex;
  align-items: center;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-logo {
  height: 32px;
  width: auto;
  object-fit: contain;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.lang-btn {
  padding: 4px 12px;
  border: 1px solid rgba(255,255,255,0.5);
  border-radius: 4px;
  background: transparent;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.lang-btn:hover {
  background: rgba(255,255,255,0.15);
  border-color: #fff;
}

.user-dropdown {
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
}

.el-aside {
  background: #fff;
  border-right: 1px solid #f0f0f0;
}

.merchant-menu {
  border: none;
  background: #fff;
}

.el-main {
  background: #f7f7f7;
  padding: 0;
  overflow-y: auto;
}

/* === 移动端 header === */
.mobile-header {
  height: 46px;
  background: #409eff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
}

.mobile-header .header-logo {
  height: 28px;
}

.mobile-header .lang-btn {
  font-size: 12px;
  padding: 2px 8px;
}

/* === 移动端底部 TabBar === */
.mobile-tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 50px;
  background: #fff;
  border-top: 1px solid #eee;
  display: flex;
  z-index: 100;
}

.mobile-tabbar .tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  font-size: 10px;
  color: #999;
  cursor: pointer;
  transition: color 0.2s;
}

.mobile-tabbar .tab-item.active {
  color: #409eff;
}

.tab-icon-wrap {
  position: relative;
}

.tab-badge {
  position: absolute;
  top: -6px;
  right: -10px;
  min-width: 16px;
  height: 16px;
  line-height: 16px;
  font-size: 10px;
  background: #f56c6c;
  color: #fff;
  border-radius: 8px;
  text-align: center;
  padding: 0 4px;
}

/* === 移动端适配 === */
.is-mobile .el-main {
  padding-top: 46px;
}

@media (max-width: 767px) {
  .el-header {
    display: none !important;
  }

  .el-aside {
    display: none !important;
  }

  .el-main {
    padding-top: 46px !important;
  }
}
</style>
