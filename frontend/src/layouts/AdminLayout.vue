<template>
  <el-container class="admin-layout" :class="{ 'is-mobile': mobile }">
    <!-- 桌面端 header -->
    <el-header v-if="!mobile">
      <div class="header-content">
        <div class="header-left">
          <img src="/images/logo2.svg" alt="Logo" class="header-logo" />
        </div>
        <div class="header-right">
          <button class="lang-btn" @click="toggleLang">{{ currentLang === 'zh' ? 'EN' : '中文' }}</button>
          <el-dropdown @command="handleCommand">
            <span class="user-dropdown">
              {{ userStore.userInfo?.full_name }} ({{ $t('admin.admin') }})
              <el-icon><arrow-down /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">{{ $t('admin.myProfile') }}</el-dropdown-item>
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
      <!-- 桌面端侧边栏 -->
      <el-aside v-if="!mobile" width="200px">
        <el-menu
          :default-active="$route.path"
          router
          class="admin-menu"
        >
          <el-menu-item index="/admin/dashboard">
            <el-icon><data-analysis /></el-icon>
            <span>{{ $t('admin.dashboard') }}</span>
          </el-menu-item>
          <el-menu-item index="/admin/products">
            <el-icon><goods /></el-icon>
            <span>{{ $t('admin.products') }}</span>
          </el-menu-item>
          <el-menu-item index="/admin/orders">
            <el-icon><list /></el-icon>
            <span>{{ $t('admin.orders') }}</span>
          </el-menu-item>
          <el-menu-item index="/admin/merchants">
            <el-icon><user /></el-icon>
            <span>{{ $t('admin.merchants') }}</span>
          </el-menu-item>
          <el-menu-item index="/admin/categories">
            <el-icon><menu /></el-icon>
            <span>{{ $t('admin.categories') }}</span>
          </el-menu-item>
          <el-menu-item index="/admin/announcements">
            <el-icon><bell /></el-icon>
            <span>{{ $t('admin.announcements') }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-main :style="mobile ? { 'padding-bottom': '60px' } : {}">
        <router-view />
      </el-main>
    </el-container>

    <!-- 移动端底部 TabBar -->
    <div v-if="mobile" class="mobile-tabbar">
      <div class="tab-item" :class="{ active: $route.path === '/admin/dashboard' }" @click="$router.push('/admin/dashboard')">
        <el-icon :size="20"><DataAnalysis /></el-icon>
        <span>{{ $t('admin.dashboard') }}</span>
      </div>
      <div class="tab-item" :class="{ active: $route.path === '/admin/products' }" @click="$router.push('/admin/products')">
        <el-icon :size="20"><Goods /></el-icon>
        <span>{{ $t('admin.products') }}</span>
      </div>
      <div class="tab-item" :class="{ active: $route.path === '/admin/orders' }" @click="$router.push('/admin/orders')">
        <el-icon :size="20"><List /></el-icon>
        <span>{{ $t('admin.orders') }}</span>
      </div>
      <div class="tab-item" :class="{ active: $route.path === '/admin/merchants' }" @click="$router.push('/admin/merchants')">
        <el-icon :size="20"><User /></el-icon>
        <span>{{ $t('admin.merchants') }}</span>
      </div>
      <div class="tab-item" :class="{ active: isMoreActive }" @click="showMore = !showMore">
        <el-icon :size="20"><MoreFilled /></el-icon>
        <span>更多</span>
      </div>

      <!-- 更多菜单弹出层 -->
      <transition name="fade">
        <div v-if="showMore" class="more-menu-overlay" @click="showMore = false">
          <div class="more-menu" @click.stop>
            <div class="more-item" @click="goTo('/admin/categories')">
              <el-icon :size="22"><Menu /></el-icon>
              <span>{{ $t('admin.categories') }}</span>
            </div>
            <div class="more-item" @click="goTo('/admin/announcements')">
              <el-icon :size="22"><Bell /></el-icon>
              <span>{{ $t('admin.announcements') }}</span>
            </div>
            <div class="more-item" @click="goTo('/admin/profile')">
              <el-icon :size="22"><User /></el-icon>
              <span>{{ $t('admin.myProfile') }}</span>
            </div>
            <div class="more-item logout" @click="handleLogout">
              <el-icon :size="22"><SwitchButton /></el-icon>
              <span>{{ $t('profile.logout') }}</span>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { ArrowDown, DataAnalysis, Goods, List, User, Menu, Bell, MoreFilled, SwitchButton } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import { setLanguage, getCurrentLanguage } from '@/i18n'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const { t } = useI18n()
const currentLang = ref(getCurrentLanguage())
const mobile = ref(window.innerWidth < 768)
const showMore = ref(false)

const isMoreActive = computed(() => {
  return ['/admin/categories', '/admin/announcements'].includes(route.path)
})

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

const goTo = (path) => {
  showMore.value = false
  router.push(path)
}

const handleLogout = () => {
  showMore.value = false
  ElMessageBox.confirm(t('admin.logoutConfirm'), t('admin.hint'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
  }).then(() => {
    userStore.logout()
    router.push('/login')
  })
}

const handleCommand = (command) => {
  if (command === 'logout') {
    handleLogout()
  } else if (command === 'profile') {
    router.push('/admin/profile')
  }
}
</script>

<style scoped>
.admin-layout {
  height: 100vh;
}

.el-header {
  background: #545c64;
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
  background: #f5f7fa;
  border-right: 1px solid #e4e7ed;
}

.admin-menu {
  border: none;
}

.el-main {
  background: #fff;
}

/* ========== 移动端 header ========== */
.mobile-header {
  height: 46px;
  background: #545c64;
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

/* ========== 移动端底部 TabBar ========== */
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
  color: #545c64;
}

/* ========== 更多菜单 ========== */
.more-menu-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 200;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.more-menu {
  background: #fff;
  border-radius: 16px 16px 0 0;
  width: 100%;
  padding: 20px 16px;
  padding-bottom: calc(20px + env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.more-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  border-radius: 10px;
  font-size: 15px;
  color: #333;
  cursor: pointer;
  transition: background 0.15s;
}

.more-item:active {
  background: #f5f5f5;
}

.more-item.logout {
  color: #f56c6c;
  margin-top: 8px;
  border-top: 1px solid #f0f0f0;
  padding-top: 18px;
  border-radius: 0;
}

/* ========== 移动端适配 ========== */
.is-mobile .el-main {
  padding-top: 46px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
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
