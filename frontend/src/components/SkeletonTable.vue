<!-- 骨架屏组件 - 数据表格加载占位 -->
<template>
  <div class="skeleton-table">
    <div class="skeleton-table-header">
      <div
        v-for="i in columns"
        :key="i"
        class="skeleton-th skeleton"
      ></div>
    </div>
    <div
      v-for="i in rows"
      :key="i"
      class="skeleton-table-row"
    >
      <div
        v-for="j in columns"
        :key="j"
        class="skeleton-td skeleton"
      ></div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  rows: {
    type: Number,
    default: 5,
  },
  columns: {
    type: Number,
    default: 4,
  },
})
</script>

<style scoped lang="scss">
.skeleton-table {
  background-color: #FFFFFF;
  border: 1px solid #E8E8E8;
  border-radius: 4px;
  overflow: hidden;
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
}

@keyframes shimmer {
  0% {
    background-position: -200% 0;
  }
  100% {
    background-position: 200% 0;
  }
}

.skeleton-table-header {
  display: flex;
  gap: 1px;
  background-color: #FAFAFA;
  padding: 12px 16px;
  border-bottom: 1px solid #E8E8E8;
}

.skeleton-th {
  flex: 1;
  height: 16px;
  border-radius: 2px;
}

.skeleton-table-row {
  display: flex;
  gap: 1px;
  padding: 16px;
  border-bottom: 1px solid #F0F0F0;
  
  &:last-child {
    border-bottom: none;
  }
}

.skeleton-td {
  flex: 1;
  height: 14px;
  border-radius: 2px;
}
</style>
