<!-- 骨架屏组件 - 商品列表加载占位 -->
<template>
  <div class="skeleton-wrapper">
    <div
      v-for="i in count"
      :key="i"
      class="skeleton-card"
    >
      <div class="skeleton-image skeleton"></div>
      <div class="skeleton-content">
        <div class="skeleton-title skeleton"></div>
        <div class="skeleton-subtitle skeleton"></div>
        <div class="skeleton-row">
          <div class="skeleton-price skeleton"></div>
          <div class="skeleton-badge skeleton"></div>
        </div>
        <div class="skeleton-actions">
          <div class="skeleton-stepper skeleton"></div>
          <div class="skeleton-button skeleton"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  count: {
    type: Number,
    default: 5,
  },
})
</script>

<style scoped lang="scss">
.skeleton-wrapper {
  padding: 8px;
}

.skeleton-card {
  display: flex;
  gap: 12px;
  padding: 12px;
  background-color: #FFFFFF;
  border: 1px solid #E8E8E8;
  border-radius: 2px;
  margin-bottom: 8px;
}

.skeleton {
  background: linear-gradient(
    90deg,
    #F0F0F0 0%,
    #F8F8F8 50%,
    #F0F0F0 100%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
  border-radius: 2px;
}

@keyframes shimmer {
  0% {
    background-position: -200% 0;
  }
  100% {
    background-position: 200% 0;
  }
}

.skeleton-image {
  width: 80px;
  height: 80px;
  flex-shrink: 0;
}

.skeleton-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-title {
  height: 18px;
  width: 60%;
}

.skeleton-subtitle {
  height: 14px;
  width: 40%;
}

.skeleton-row {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
}

.skeleton-price {
  height: 20px;
  width: 80px;
}

.skeleton-badge {
  height: 20px;
  width: 60px;
}

.skeleton-actions {
  display: flex;
  gap: 8px;
  margin-top: auto;
}

.skeleton-stepper {
  height: 32px;
  flex: 1;
}

.skeleton-button {
  height: 32px;
  width: 80px;
}
</style>
