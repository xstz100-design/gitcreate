import { defineStore } from 'pinia'
import { ref, computed } from 'vue'


export const useCartStore = defineStore(
  'cart',
  () => {
    const items = ref([])

    // 购物车商品数量
    const totalCount = computed(() => {
      return items.value.reduce((sum, item) => sum + item.quantity, 0)
    })

    // 购物车总价 (USD)
    const totalPrice = computed(() => {
      return items.value.reduce((sum, item) => sum + item.price_usd * item.quantity, 0)
    })

    // 添加到购物车
    function addItem(product, quantity = 1) {
      const existingItem = items.value.find((item) => item.id === product.id)

      if (existingItem) {
        // 检查库存
        if (existingItem.quantity + quantity > product.stock) {
          return false
        }
        existingItem.quantity += quantity
      } else {
        if (quantity > product.stock) {
          return false
        }
        items.value.push({
          id: product.id,
          name: product.name,
          name_kh: product.name_kh,
          unit: product.unit,
          price_usd: product.price_usd,
          stock: product.stock,
          image_url: product.image_url,
          quantity,
        })
      }

      return true
    }

    // 更新购物车商品数量
    function updateQuantity(productId, quantity) {
      const item = items.value.find((item) => item.id === productId)
      if (item) {
        if (quantity <= 0) {
          removeItem(productId)
          return
        }
        if (quantity > item.stock) {
          return false
        }
        item.quantity = quantity
        return true
      }
      return false
    }

    // 从购物车移除
    function removeItem(productId) {
      const index = items.value.findIndex((item) => item.id === productId)
      if (index > -1) {
        items.value.splice(index, 1)
      }
    }

    // 清空购物车
    function clear() {
      items.value = []
    }

    return {
      items,
      totalCount,
      totalPrice,
      addItem,
      updateQuantity,
      removeItem,
      clear,
    }
  },
  {
    persist: {
      key: 'cart',
      storage: localStorage,
      pick: ['items'],
    },
  }
)
