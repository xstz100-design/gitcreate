package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"wholesale/config"
	"wholesale/database"
	"wholesale/middleware"
	"wholesale/models"
	"wholesale/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─────────────────── 请求/响应结构 ───────────────────

type OrderItemRequest struct {
	ProductID    int64  `json:"product_id" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required,gt=0"`
	PurchaseMode string `json:"purchase_mode"` // default | piece | package
}

type OrderCreateRequest struct {
	Items           []OrderItemRequest   `json:"items" binding:"required,min=1"`
	PaymentStatus   models.PaymentStatus `json:"payment_status"`
	DeliveryAddress *string              `json:"delivery_address"`
	DeliveryPhone   *string              `json:"delivery_phone"`
	Note            *string              `json:"note"`
	DistanceKM      float64              `json:"distance_km"`
	ScheduledAt     *time.Time           `json:"scheduled_at"`
	ClientRequestID *string              `json:"client_request_id"`
}

type OrderUpdateRequest struct {
	PaymentStatus   *models.PaymentStatus  `json:"payment_status"`
	DeliveryStatus  *models.DeliveryStatus `json:"delivery_status"`
	DeliveryAddress *string                `json:"delivery_address"`
	DeliveryPhone   *string                `json:"delivery_phone"`
	Note            *string                `json:"note"`
	ScheduledAt     *time.Time             `json:"scheduled_at"`
}

type OrderItemResponse struct {
	ID           int64   `json:"id"`
	ProductID    int64   `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	UnitPriceUSD float64 `json:"unit_price_usd"`
	SubtotalUSD  float64 `json:"subtotal_usd"`
	PurchaseMode string  `json:"purchase_mode"`
}

type OrderDetailResponse struct {
	ID              int64                 `json:"id"`
	OrderNo         string                `json:"order_no"`
	MerchantID      int64                 `json:"merchant_id"`
	MerchantName    string                `json:"merchant_name"`
	TotalUSD        float64               `json:"total_usd"`
	GoodsTotalUSD   float64               `json:"goods_total_usd"`
	DeliveryFeeUSD  float64               `json:"delivery_fee_usd"`
	DistanceKM      *float64              `json:"distance_km"`
	TotalKHR        float64               `json:"total_khr"`
	PaymentStatus   models.PaymentStatus  `json:"payment_status"`
	DeliveryStatus  models.DeliveryStatus `json:"delivery_status"`
	DeliveryAddress *string               `json:"delivery_address"`
	DeliveryPhone   *string               `json:"delivery_phone"`
	Note            *string               `json:"note"`
	Items           []OrderItemResponse   `json:"items"`
	CreatedAt       string                `json:"created_at"`
	ScheduledAt     *string               `json:"scheduled_at"`
	PickedAt        *string               `json:"picked_at"`
	UnpaidDays      *int                  `json:"unpaid_days"`
	DaysToBilling   *int                  `json:"days_to_billing"`
}

func buildOrderResponse(order *models.Order, merchantName string) OrderDetailResponse {
	items := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		productName := "已删除"
		if item.Product != nil {
			productName = item.Product.Name
		}
		items[i] = OrderItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  productName,
			Quantity:     item.Quantity,
			UnitPriceUSD: item.UnitPriceUSD,
			SubtotalUSD:  item.SubtotalUSD,
			PurchaseMode: item.PurchaseMode,
		}
	}

	var scheduledAt, pickedAt *string
	if order.ScheduledAt != nil {
		s := order.ScheduledAt.Format(time.RFC3339)
		scheduledAt = &s
	}
	if order.PickedAt != nil {
		s := order.PickedAt.Format(time.RFC3339)
		pickedAt = &s
	}

	unpaidDays := calcUnpaidDays(order)
	var daysToBilling *int
	if order.Merchant != nil && order.PaymentStatus == models.PaymentMonthly && order.Merchant.BillingCycleDays != nil {
		d := *order.Merchant.BillingCycleDays
		daysToBilling = &d
	}

	return OrderDetailResponse{
		ID:              order.ID,
		OrderNo:         order.OrderNo,
		MerchantID:      order.MerchantID,
		MerchantName:    merchantName,
		TotalUSD:        order.TotalUSD,
		GoodsTotalUSD:   order.GoodsTotalUSD,
		DeliveryFeeUSD:  order.DeliveryFeeUSD,
		DistanceKM:      order.DistanceKM,
		TotalKHR:        order.TotalKHR(config.C.USDToKHRRate),
		PaymentStatus:   order.PaymentStatus,
		DeliveryStatus:  order.DeliveryStatus,
		DeliveryAddress: order.DeliveryAddress,
		DeliveryPhone:   order.DeliveryPhone,
		Note:            order.Note,
		Items:           items,
		CreatedAt:       order.CreatedAt.Format(time.RFC3339),
		ScheduledAt:     scheduledAt,
		PickedAt:        pickedAt,
		UnpaidDays:      unpaidDays,
		DaysToBilling:   daysToBilling,
	}
}

func calcUnpaidDays(order *models.Order) *int {
	if order.PaymentStatus == models.PaymentCash {
		return nil
	}
	if order.DeliveryStatus == models.DeliveryCancelled {
		return nil
	}
	loc := time.FixedZone("Asia/Phnom_Penh", 7*60*60)
	now := time.Now().In(loc)
	created := order.CreatedAt.In(loc)
	days := int(now.Sub(created).Hours() / 24)
	return &days
}

func generateOrderNo() string {
	loc := time.FixedZone("Asia/Phnom_Penh", 7*60*60)
	today := time.Now().In(loc).Format("20060102")
	prefix := "ORD" + today

	var count int64
	database.DB.Model(&models.Order{}).Where("order_no LIKE ?", prefix+"%").Count(&count)
	return fmt.Sprintf("%s%06d", prefix, count+1)
}

func loadOrder(id int64) (*models.Order, error) {
	var order models.Order
	err := database.DB.
		Preload("Merchant").
		Preload("Items.Product").
		Where("id = ? AND is_deleted = ?", id, false).
		First(&order).Error
	return &order, err
}

// ─────────────────── GET /api/orders ───────────────────

func ListOrders(c *gin.Context) {
	user := middleware.CurrentUser(c)

	q := database.DB.
		Preload("Merchant").
		Preload("Items.Product").
		Where("orders.is_deleted = ?", false)

	if user.Role == models.RoleMerchant {
		q = q.Where("merchant_id = ?", user.ID)
	} else {
		if mid := c.Query("merchant_id"); mid != "" {
			q = q.Where("merchant_id = ?", mid)
		}
	}
	if ps := c.Query("payment_status"); ps != "" {
		q = q.Where("payment_status = ?", ps)
	}
	if ds := c.Query("delivery_status"); ds != "" {
		q = q.Where("delivery_status = ?", ds)
	}

	var orders []models.Order
	q.Order("created_at DESC").Find(&orders)

	result := make([]OrderDetailResponse, len(orders))
	for i, o := range orders {
		o := o
		merchantName := ""
		if o.Merchant != nil {
			merchantName = o.Merchant.FullName
		}
		result[i] = buildOrderResponse(&o, merchantName)
	}
	c.JSON(http.StatusOK, result)
}

// ─────────────────── GET /api/orders/:id ───────────────────

func GetOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	order, err := loadOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "订单不存在"})
		return
	}

	user := middleware.CurrentUser(c)
	if user.Role == models.RoleMerchant && order.MerchantID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"detail": "无权访问此订单"})
		return
	}

	merchantName := ""
	if order.Merchant != nil {
		merchantName = order.Merchant.FullName
	}
	c.JSON(http.StatusOK, buildOrderResponse(order, merchantName))
}

// ─────────────────── POST /api/orders ───────────────────

func CreateOrder(c *gin.Context) {
	user := middleware.CurrentUser(c)

	if user.Role == models.RoleMerchant {
		if user.ApprovalStatus != models.ApprovalApproved {
			c.JSON(http.StatusForbidden, gin.H{"detail": "您的账号尚未通过审核，无法下单"})
			return
		}
		if user.FullName == "" || user.Phone == nil || *user.Phone == "" || user.Address == nil || *user.Address == "" {
			c.JSON(http.StatusForbidden, gin.H{"detail": "请先完善个人资料（姓名、电话、地址）后再下单"})
			return
		}
	}

	var req OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// 幂等校验
	if req.ClientRequestID != nil && *req.ClientRequestID != "" {
		var existing models.Order
		if database.DB.Where("merchant_id = ? AND client_request_id = ? AND is_deleted = ?",
			user.ID, *req.ClientRequestID, false).First(&existing).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"detail": "请勿重复提交订单"})
			return
		}
	}

	// 月结权限
	if req.PaymentStatus == models.PaymentMonthly && !user.AllowCredit {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "您没有赊账权限，请联系管理员开通"})
		return
	}

	// 在事务中验证库存并扣减
	var orderItemsData []struct {
		ProductID    int64
		ProductName  string
		Barcode      *string
		Unit         string
		Quantity     int
		UnitPriceUSD float64
		SubtotalUSD  float64
		PurchaseMode string
	}
	var goodsTotal float64

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			var product models.Product
			if err := tx.Where("id = ? AND is_deleted = ?", item.ProductID, false).First(&product).Error; err != nil {
				return fmt.Errorf("商品 ID %d 不存在", item.ProductID)
			}
			if !product.IsActive {
				return fmt.Errorf("商品 %s 已下架", product.Name)
			}
			if product.Stock < item.Quantity {
				return fmt.Errorf("商品 %s 库存不足，当前库存: %d", product.Name, product.Stock)
			}

			// 原子扣减库存
			result := tx.Model(&models.Product{}).
				Where("id = ? AND stock >= ?", product.ID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))
			if result.Error != nil || result.RowsAffected == 0 {
				return fmt.Errorf("商品 %s 库存不足，请重试", product.Name)
			}

			// 确定单价
			purchaseMode := item.PurchaseMode
			if purchaseMode == "" {
				purchaseMode = "default"
			}
			var unitPrice float64
			switch purchaseMode {
			case "piece":
				if product.PricePerPieceUSD == nil {
					return fmt.Errorf("商品 %s 暂未配置按件价格", product.Name)
				}
				unitPrice = *product.PricePerPieceUSD
			case "package":
				if product.PricePerPackageUSD == nil {
					return fmt.Errorf("商品 %s 暂未配置按包价格", product.Name)
				}
				unitPrice = *product.PricePerPackageUSD
			default:
				unitPrice = product.PriceUSD
			}

			subtotal := unitPrice * float64(item.Quantity)
			goodsTotal += subtotal
			orderItemsData = append(orderItemsData, struct {
				ProductID    int64
				ProductName  string
				Barcode      *string
				Unit         string
				Quantity     int
				UnitPriceUSD float64
				SubtotalUSD  float64
				PurchaseMode string
			}{
				ProductID:    product.ID,
				ProductName:  product.Name,
				Barcode:      product.Barcode,
				Unit:         product.Unit,
				Quantity:     item.Quantity,
				UnitPriceUSD: unitPrice,
				SubtotalUSD:  subtotal,
				PurchaseMode: purchaseMode,
			})
		}

		// 配送费
		freeKM, feePerKM := services.GetDeliveryFeeSettings()
		distKM := req.DistanceKM
		if distKM < 0 {
			distKM = 0
		}
		deliveryFee := services.CalculateDeliveryFee(distKM, freeKM, feePerKM)
		totalUSD := goodsTotal + deliveryFee

		// 配送地址默认用商户资料
		deliveryAddr := req.DeliveryAddress
		if deliveryAddr == nil || *deliveryAddr == "" {
			deliveryAddr = user.Address
		}
		deliveryPhone := req.DeliveryPhone
		if deliveryPhone == nil || *deliveryPhone == "" {
			deliveryPhone = user.Phone
		}

		now := models.NowCambodia()
		order := models.Order{
			OrderNo:         generateOrderNo(),
			MerchantID:      user.ID,
			TotalUSD:        totalUSD,
			GoodsTotalUSD:   goodsTotal,
			DeliveryFeeUSD:  deliveryFee,
			DistanceKM:      &distKM,
			PaymentStatus:   req.PaymentStatus,
			DeliveryAddress: deliveryAddr,
			DeliveryPhone:   deliveryPhone,
			Note:            req.Note,
			ClientRequestID: req.ClientRequestID,
			ScheduledAt:     req.ScheduledAt,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		for _, d := range orderItemsData {
			item := models.OrderItem{
				OrderID:      order.ID,
				ProductID:    d.ProductID,
				Quantity:     d.Quantity,
				UnitPriceUSD: d.UnitPriceUSD,
				SubtotalUSD:  d.SubtotalUSD,
				PurchaseMode: d.PurchaseMode,
				CreatedAt:    now,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		// 月结累加
		if req.PaymentStatus == models.PaymentMonthly {
			tx.Model(user).UpdateColumn("credit_limit", gorm.Expr("credit_limit + ?", totalUSD))
		}

		// 重新加载订单用于返回
		tx.Preload("Merchant").Preload("Items.Product").First(&order, order.ID)

		// Telegram 通知（非阻塞）
		go func(o models.Order, items []map[string]interface{}) {
			services.NotifyAdminsNewOrder(o.OrderNo, user.FullName, o.TotalUSD, items)
			notifyItems := make([]map[string]interface{}, len(orderItemsData))
			for i, d := range orderItemsData {
				notifyItems[i] = map[string]interface{}{
					"name":     d.ProductName,
					"barcode":  d.Barcode,
					"unit":     d.Unit,
					"quantity": d.Quantity,
				}
			}
			services.NotifyPickerNewOrder(o.OrderNo, o.ID, notifyItems)
			// 库存预警检查
			for _, d := range orderItemsData {
				var p models.Product
				if database.DB.First(&p, d.ProductID).Error == nil && p.IsLowStock() {
					services.NotifyLowStock(p.Name, p.Stock)
				}
			}
		}(order, nil)

		merchantName := ""
		if order.Merchant != nil {
			merchantName = order.Merchant.FullName
		}
		c.JSON(http.StatusCreated, buildOrderResponse(&order, merchantName))
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
	}
}

// ─────────────────── PATCH /api/orders/:id ───────────────────

func UpdateOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	order, err := loadOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "订单不存在"})
		return
	}

	user := middleware.CurrentUser(c)
	// 商户只能修改自己的订单，且只能修改备注/预约时间
	if user.Role == models.RoleMerchant && order.MerchantID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"detail": "无权修改此订单"})
		return
	}

	var req OrderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updates := map[string]interface{}{"updated_at": models.NowCambodia()}

	if user.Role == models.RoleAdmin {
		if req.PaymentStatus != nil {
			updates["payment_status"] = *req.PaymentStatus
		}
		if req.DeliveryStatus != nil {
			updates["delivery_status"] = *req.DeliveryStatus
			// 配送中 → 通知配送员
			if *req.DeliveryStatus == models.DeliveryDelivering {
				merchantName := ""
				addr := ""
				phone := ""
				if order.Merchant != nil {
					merchantName = order.Merchant.FullName
				}
				if order.DeliveryAddress != nil {
					addr = *order.DeliveryAddress
				}
				if order.DeliveryPhone != nil {
					phone = *order.DeliveryPhone
				}
				go services.NotifyDeliveryOrder(order.OrderNo, order.ID, merchantName, addr, phone)
			}
			// 已送达
			if *req.DeliveryStatus == models.DeliveryDelivered {
				now := models.NowCambodia()
				updates["delivered_at"] = now
			}
		}
		if req.DeliveryAddress != nil {
			updates["delivery_address"] = *req.DeliveryAddress
		}
		if req.DeliveryPhone != nil {
			updates["delivery_phone"] = *req.DeliveryPhone
		}
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}
	if req.ScheduledAt != nil {
		updates["scheduled_at"] = *req.ScheduledAt
	}

	database.DB.Model(order).Updates(updates)
	updated, _ := loadOrder(id)
	merchantName := ""
	if updated.Merchant != nil {
		merchantName = updated.Merchant.FullName
	}
	c.JSON(http.StatusOK, buildOrderResponse(updated, merchantName))
}

// ─────────────────── DELETE /api/orders/:id ───────────────────

func DeleteOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var order models.Order
	if err := database.DB.Where("id = ? AND is_deleted = ?", id, false).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "订单不存在"})
		return
	}

	user := middleware.CurrentUser(c)
	if user.Role == models.RoleMerchant && order.MerchantID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"detail": "无权删除此订单"})
		return
	}

	// 恢复库存
	var items []models.OrderItem
	database.DB.Where("order_id = ?", id).Find(&items)
	for _, item := range items {
		database.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity))
	}

	database.DB.Model(&order).Updates(map[string]interface{}{
		"is_deleted":      true,
		"updated_at":      models.NowCambodia(),
		"delivery_status": models.DeliveryCancelled,
	})
	c.JSON(http.StatusOK, gin.H{"message": "订单已取消"})
}

// ─────────────────── POST /api/orders/:id/cancel ───────────────────

// POST /api/orders/:id/cancel — 商户取消自己的订单
func CancelOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	user := middleware.CurrentUser(c)

	var order models.Order
	if err := database.DB.Where("id = ? AND is_deleted = ?", id, false).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "订单不存在"})
		return
	}

	// 商户只能取消自己的订单；管理员可取消任意订单
	if user.Role == models.RoleMerchant && order.MerchantID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"detail": "无权取消此订单"})
		return
	}

	// 只有 pending 状态可取消
	if order.DeliveryStatus != models.DeliveryPending {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "只有待处理的订单可以取消"})
		return
	}

	// 恢复库存
	var items []models.OrderItem
	database.DB.Where("order_id = ?", id).Find(&items)
	for _, item := range items {
		database.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity))
	}

	database.DB.Model(&order).Updates(map[string]interface{}{
		"delivery_status": models.DeliveryCancelled,
		"updated_at":      models.NowCambodia(),
	})
	database.DB.Preload("Items.Product").Preload("Merchant").First(&order, id)
	merchantName := ""
	if order.Merchant != nil {
		merchantName = order.Merchant.FullName
	}
	c.JSON(http.StatusOK, buildOrderResponse(&order, merchantName))
}

// ─────────────────── 配货员视图 ───────────────────

// GET /api/orders/picker/items/:orderId
func GetPickerItems(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("orderId"), 10, 64)
	var order models.Order
	if err := database.DB.Preload("Items.Product").Where("id = ? AND is_deleted = ?", id, false).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "订单不存在"})
		return
	}

	items := make([]gin.H, len(order.Items))
	for i, item := range order.Items {
		productName := "已删除"
		imageURL := ""
		if item.Product != nil {
			productName = item.Product.Name
			if item.Product.ImageURL != nil {
				imageURL = *item.Product.ImageURL
			}
		}
		items[i] = gin.H{
			"id":            item.ID,
			"product_id":    item.ProductID,
			"product_name":  productName,
			"image_url":     imageURL,
			"quantity":      item.Quantity,
			"purchase_mode": item.PurchaseMode,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"order_id": order.ID,
		"order_no": order.OrderNo,
		"items":    items,
	})
}

// POST /api/orders/:id/pick — 标记已配货
func MarkOrderPicked(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var order models.Order
	if err := database.DB.Where("id = ? AND is_deleted = ?", id, false).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "订单不存在"})
		return
	}

	user := middleware.CurrentUser(c)
	now := models.NowCambodia()
	database.DB.Model(&order).Updates(map[string]interface{}{
		"picked_at":    now,
		"picked_by_id": user.ID,
		"updated_at":   now,
	})
	c.JSON(http.StatusOK, gin.H{"message": "已标记配货完成"})
}
