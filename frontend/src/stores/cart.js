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

    function resolvePriceByMode(product, purchaseMode = 'default') {
      if (purchaseMode === 'piece' && product.price_per_piece_usd) {
        return Number(product.price_per_piece_usd)
      }
      if (purchaseMode === 'package') {
        if (product.price_per_package_usd) {
          return Number(product.price_per_package_usd)
        }
        if (product.pieces_per_package) {
          return Number(product.price_usd) * Number(product.pieces_per_package)
        }
      }
      return Number(product.price_usd)
    }

    // 添加到购物车
    function addItem(product, quantity = 1, purchaseMode = 'default') {
      const existingItem = items.value.find((item) => item.id === product.id && item.purchase_mode === purchaseMode)
      const unitPrice = resolvePriceByMode(product, purchaseMode)

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
          price_usd: unitPrice,
          purchase_mode: purchaseMode,
          display_unit: purchaseMode === 'piece'
            ? (product.unit_name || '件')
            : purchaseMode === 'package'
              ? (product.pack_name || '箱')
              : product.unit,
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
