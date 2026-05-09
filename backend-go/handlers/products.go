package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wholesale/database"
	"wholesale/middleware"
	"wholesale/models"
	"wholesale/services"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ─────────────────── 请求/响应结构 ───────────────────

// parseDateFlexible 兼容前端发送的无时区日期字符串（如 "2026-05-08T00:00:00" 或 "2026-05-08"）
func parseDateFlexible(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	loc := time.FixedZone("Asia/Phnom_Penh", 7*60*60)
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, *s, loc); err == nil {
			return &t
		}
	}
	return nil
}

type ProductCreateRequest struct {
	Name               string   `json:"name" binding:"required"`
	NameKh             *string  `json:"name_kh"`
	NameEn             *string  `json:"name_en"`
	Brand              *string  `json:"brand"`
	CountryOfOrigin    *string  `json:"country_of_origin"`
	Unit               string   `json:"unit"`
	Specs              *string  `json:"specs"`
	Barcode            *string  `json:"barcode"`
	PriceUSD           float64  `json:"price_usd" binding:"required,gt=0"`
	RetailPriceUSD     *float64 `json:"retail_price_usd"`
	PricePerPieceUSD   *float64 `json:"price_per_piece_usd"`
	PricePerPackageUSD *float64 `json:"price_per_package_usd"`
	PiecesPerPackage   *int     `json:"pieces_per_package"`
	UnitName           *string  `json:"unit_name"`
	PackName           *string  `json:"pack_name"`
	Stock              int      `json:"stock"`
	StockWarning       int      `json:"stock_warning"`
	Description        *string  `json:"description"`
	ImageURL           *string  `json:"image_url"`
	Img1               *string  `json:"img1"`
	Img2               *string  `json:"img2"`
	Img3               *string  `json:"img3"`
	Img4               *string  `json:"img4"`
	Img5               *string  `json:"img5"`
	Category           *string  `json:"category"`
	SortOrder          int      `json:"sort_order"`
	IsFeatured         bool     `json:"is_featured"`
	IsActive           bool     `json:"is_active"`
	ProductionDate     *string  `json:"production_date"`
	ExpiryDate         *string  `json:"expiry_date"`
	// 供应商
	SupplierName     *string `json:"supplier_name"`
	PrincipleCompany *string `json:"principle_company"`
	// 基础扩展
	UnitWeightValue *float64 `json:"unit_weight_value"`
	UnitWeightUnit  *string  `json:"unit_weight_unit"`
	PackingFormat   *string  `json:"packing_format"`
	PackSize        *float64 `json:"pack_size"`
	GpPercent       *float64 `json:"gp_percent"`
	ShelfLifeDays   *int     `json:"shelf_life_days"`
	// 包装规格层级
	InnerPackPerCase *int `json:"inner_pack_per_case"`
	UnitPerInnerPack *int `json:"unit_per_inner_pack"`
	UnitPerCase      *int `json:"unit_per_case"`
	// 成本与价格
	CostPerCase    *float64 `json:"cost_per_case"`
	DcPercent      *float64 `json:"dc_percent"`
	NetCostPerCase *float64 `json:"net_cost_per_case"`
	NetCostPerUnit *float64 `json:"net_cost_per_unit"`
	PriceInclVat   *float64 `json:"price_incl_vat"`
	PriceExclVat   *float64 `json:"price_excl_vat"`
	// 尺寸 最小包
	UnitWidthCm  *float64 `json:"unit_width_cm"`
	UnitLengthCm *float64 `json:"unit_length_cm"`
	UnitHeightCm *float64 `json:"unit_height_cm"`
	UnitWeightKg *float64 `json:"unit_weight_kg"`
	// 尺寸 中包
	PackWidthCm  *float64 `json:"pack_width_cm"`
	PackLengthCm *float64 `json:"pack_length_cm"`
	PackHeightCm *float64 `json:"pack_height_cm"`
	PackWeightKg *float64 `json:"pack_weight_kg"`
	// 尺寸 外箱
	CaseWidthCm  *float64 `json:"case_width_cm"`
	CaseLengthCm *float64 `json:"case_length_cm"`
	CaseHeightCm *float64 `json:"case_height_cm"`
	CaseWeightKg *float64 `json:"case_weight_kg"`
}

type ProductUpdateRequest struct {
	Name               *string  `json:"name"`
	NameKh             *string  `json:"name_kh"`
	NameEn             *string  `json:"name_en"`
	Brand              *string  `json:"brand"`
	CountryOfOrigin    *string  `json:"country_of_origin"`
	Unit               *string  `json:"unit"`
	Specs              *string  `json:"specs"`
	Barcode            *string  `json:"barcode"`
	PriceUSD           *float64 `json:"price_usd"`
	RetailPriceUSD     *float64 `json:"retail_price_usd"`
	PricePerPieceUSD   *float64 `json:"price_per_piece_usd"`
	PricePerPackageUSD *float64 `json:"price_per_package_usd"`
	PiecesPerPackage   *int     `json:"pieces_per_package"`
	UnitName           *string  `json:"unit_name"`
	PackName           *string  `json:"pack_name"`
	Stock              *int     `json:"stock"`
	StockWarning       *int     `json:"stock_warning"`
	Description        *string  `json:"description"`
	ImageURL           *string  `json:"image_url"`
	Img1               *string  `json:"img1"`
	Img2               *string  `json:"img2"`
	Img3               *string  `json:"img3"`
	Img4               *string  `json:"img4"`
	Img5               *string  `json:"img5"`
	Category           *string  `json:"category"`
	SortOrder          *int     `json:"sort_order"`
	IsFeatured         *bool    `json:"is_featured"`
	IsActive           *bool    `json:"is_active"`
	ProductionDate     *string  `json:"production_date"`
	ExpiryDate         *string  `json:"expiry_date"`
	// 供应商
	SupplierName     *string `json:"supplier_name"`
	PrincipleCompany *string `json:"principle_company"`
	// 基础扩展
	UnitWeightValue *float64 `json:"unit_weight_value"`
	UnitWeightUnit  *string  `json:"unit_weight_unit"`
	PackingFormat   *string  `json:"packing_format"`
	PackSize        *float64 `json:"pack_size"`
	GpPercent       *float64 `json:"gp_percent"`
	ShelfLifeDays   *int     `json:"shelf_life_days"`
	// 包装规格层级
	InnerPackPerCase *int `json:"inner_pack_per_case"`
	UnitPerInnerPack *int `json:"unit_per_inner_pack"`
	UnitPerCase      *int `json:"unit_per_case"`
	// 成本与价格
	CostPerCase    *float64 `json:"cost_per_case"`
	DcPercent      *float64 `json:"dc_percent"`
	NetCostPerCase *float64 `json:"net_cost_per_case"`
	NetCostPerUnit *float64 `json:"net_cost_per_unit"`
	PriceInclVat   *float64 `json:"price_incl_vat"`
	PriceExclVat   *float64 `json:"price_excl_vat"`
	// 尺寸 最小包
	UnitWidthCm  *float64 `json:"unit_width_cm"`
	UnitLengthCm *float64 `json:"unit_length_cm"`
	UnitHeightCm *float64 `json:"unit_height_cm"`
	UnitWeightKg *float64 `json:"unit_weight_kg"`
	// 尺寸 中包
	PackWidthCm  *float64 `json:"pack_width_cm"`
	PackLengthCm *float64 `json:"pack_length_cm"`
	PackHeightCm *float64 `json:"pack_height_cm"`
	PackWeightKg *float64 `json:"pack_weight_kg"`
	// 尺寸 外箱
	CaseWidthCm  *float64 `json:"case_width_cm"`
	CaseLengthCm *float64 `json:"case_length_cm"`
	CaseHeightCm *float64 `json:"case_height_cm"`
	CaseWeightKg *float64 `json:"case_weight_kg"`
}

type ProductResponse struct {
	ID                 int64      `json:"id"`
	Name               string     `json:"name"`
	NameKh             *string    `json:"name_kh"`
	NameEn             *string    `json:"name_en"`
	Brand              *string    `json:"brand"`
	CountryOfOrigin    *string    `json:"country_of_origin"`
	Unit               string     `json:"unit"`
	Specs              *string    `json:"specs"`
	Barcode            *string    `json:"barcode"`
	PriceUSD           float64    `json:"price_usd"`
	RetailPriceUSD     *float64   `json:"retail_price_usd"`
	PricePerPieceUSD   *float64   `json:"price_per_piece_usd"`
	PricePerPackageUSD *float64   `json:"price_per_package_usd"`
	PiecesPerPackage   *int       `json:"pieces_per_package"`
	UnitName           *string    `json:"unit_name"`
	PackName           *string    `json:"pack_name"`
	Stock              int        `json:"stock"`
	StockWarning       int        `json:"stock_warning"`
	IsLowStock         bool       `json:"is_low_stock"`
	Description        *string    `json:"description"`
	ImageURL           *string    `json:"image_url"`
	Img1               *string    `json:"img1"`
	Img2               *string    `json:"img2"`
	Img3               *string    `json:"img3"`
	Img4               *string    `json:"img4"`
	Img5               *string    `json:"img5"`
	Category           *string    `json:"category"`
	SortOrder          int        `json:"sort_order"`
	IsActive           bool       `json:"is_active"`
	IsFeatured         bool       `json:"is_featured"`
	ProductionDate     *time.Time `json:"production_date"`
	ExpiryDate         *time.Time `json:"expiry_date"`
	UpdatedAt          time.Time  `json:"updated_at"`
	// 供应商
	SupplierName     *string `json:"supplier_name"`
	PrincipleCompany *string `json:"principle_company"`
	// 基础扩展
	UnitWeightValue *float64 `json:"unit_weight_value"`
	UnitWeightUnit  *string  `json:"unit_weight_unit"`
	PackingFormat   *string  `json:"packing_format"`
	PackSize        *float64 `json:"pack_size"`
	GpPercent       *float64 `json:"gp_percent"`
	ShelfLifeDays   *int     `json:"shelf_life_days"`
	// 包装规格层级
	InnerPackPerCase *int `json:"inner_pack_per_case"`
	UnitPerInnerPack *int `json:"unit_per_inner_pack"`
	UnitPerCase      *int `json:"unit_per_case"`
	// 成本与价格
	CostPerCase    *float64 `json:"cost_per_case"`
	DcPercent      *float64 `json:"dc_percent"`
	NetCostPerCase *float64 `json:"net_cost_per_case"`
	NetCostPerUnit *float64 `json:"net_cost_per_unit"`
	PriceInclVat   *float64 `json:"price_incl_vat"`
	PriceExclVat   *float64 `json:"price_excl_vat"`
	// 尺寸 最小包
	UnitWidthCm  *float64 `json:"unit_width_cm"`
	UnitLengthCm *float64 `json:"unit_length_cm"`
	UnitHeightCm *float64 `json:"unit_height_cm"`
	UnitWeightKg *float64 `json:"unit_weight_kg"`
	// 尺寸 中包
	PackWidthCm  *float64 `json:"pack_width_cm"`
	PackLengthCm *float64 `json:"pack_length_cm"`
	PackHeightCm *float64 `json:"pack_height_cm"`
	PackWeightKg *float64 `json:"pack_weight_kg"`
	// 尺寸 外箱
	CaseWidthCm  *float64 `json:"case_width_cm"`
	CaseLengthCm *float64 `json:"case_length_cm"`
	CaseHeightCm *float64 `json:"case_height_cm"`
	CaseWeightKg *float64 `json:"case_weight_kg"`
}

func buildProductResponse(p *models.Product) ProductResponse {
	return ProductResponse{
		ID: p.ID, Name: p.Name, NameKh: p.NameKh, NameEn: p.NameEn,
		Brand: p.Brand, CountryOfOrigin: p.CountryOfOrigin,
		Unit: p.Unit, Specs: p.Specs, Barcode: p.Barcode,
		PriceUSD: p.PriceUSD, RetailPriceUSD: p.RetailPriceUSD,
		PricePerPieceUSD: p.PricePerPieceUSD, PricePerPackageUSD: p.PricePerPackageUSD,
		PiecesPerPackage: p.PiecesPerPackage, UnitName: p.UnitName, PackName: p.PackName,
		Stock: p.Stock, StockWarning: p.StockWarning, IsLowStock: p.IsLowStock(),
		Description: p.Description, ImageURL: p.ImageURL,
		Img1: p.Img1, Img2: p.Img2, Img3: p.Img3, Img4: p.Img4, Img5: p.Img5,
		Category: p.Category, SortOrder: p.SortOrder, IsActive: p.IsActive,
		IsFeatured: p.IsFeatured, ProductionDate: p.ProductionDate,
		ExpiryDate: p.ExpiryDate, UpdatedAt: p.UpdatedAt,
		// 新字段
		SupplierName: p.SupplierName, PrincipleCompany: p.PrincipleCompany,
		UnitWeightValue: p.UnitWeightValue, UnitWeightUnit: p.UnitWeightUnit,
		PackingFormat: p.PackingFormat, PackSize: p.PackSize,
		GpPercent: p.GpPercent, ShelfLifeDays: p.ShelfLifeDays,
		InnerPackPerCase: p.InnerPackPerCase, UnitPerInnerPack: p.UnitPerInnerPack, UnitPerCase: p.UnitPerCase,
		CostPerCase: p.CostPerCase, DcPercent: p.DcPercent,
		NetCostPerCase: p.NetCostPerCase, NetCostPerUnit: p.NetCostPerUnit,
		PriceInclVat: p.PriceInclVat, PriceExclVat: p.PriceExclVat,
		UnitWidthCm: p.UnitWidthCm, UnitLengthCm: p.UnitLengthCm, UnitHeightCm: p.UnitHeightCm, UnitWeightKg: p.UnitWeightKg,
		PackWidthCm: p.PackWidthCm, PackLengthCm: p.PackLengthCm, PackHeightCm: p.PackHeightCm, PackWeightKg: p.PackWeightKg,
		CaseWidthCm: p.CaseWidthCm, CaseLengthCm: p.CaseLengthCm, CaseHeightCm: p.CaseHeightCm, CaseWeightKg: p.CaseWeightKg,
	}
}

// ─────────────────── CRUD ───────────────────

// GET /api/products
func ListProducts(c *gin.Context) {
	q := database.DB.Where("is_deleted = ?", false)

	if v := c.Query("is_active"); v != "" {
		b, _ := strconv.ParseBool(v)
		q = q.Where("is_active = ?", b)
	} else {
		q = q.Where("is_active = ?", true)
	}
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	if c.Query("low_stock_only") == "true" {
		q = q.Where("stock <= stock_warning")
	}

	var products []models.Product
	q.Order("sort_order ASC, id ASC").Find(&products)

	result := make([]ProductResponse, len(products))
	for i, p := range products {
		p := p
		result[i] = buildProductResponse(&p)
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/products/:id
func GetProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var product models.Product
	if err := database.DB.Where("id = ? AND is_deleted = ?", id, false).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "商品不存在"})
		return
	}
	// 浏览量：只统计商户/配货员点开商品详情，管理员不计
	if u := middleware.CurrentUser(c); u != nil && u.Role != models.RoleAdmin {
		services.TrackPageView()
	}
	c.JSON(http.StatusOK, buildProductResponse(&product))
}

// POST /api/products
func CreateProduct(c *gin.Context) {
	_ = middleware.CurrentUser(c) // require_admin already checked

	var req ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	unit := req.Unit
	if unit == "" {
		unit = "件"
	}

	now := models.NowCambodia()
	product := models.Product{
		Name: req.Name, NameKh: req.NameKh, NameEn: req.NameEn,
		Brand: req.Brand, CountryOfOrigin: req.CountryOfOrigin,
		Unit: unit, Specs: req.Specs, Barcode: req.Barcode,
		PriceUSD: req.PriceUSD, RetailPriceUSD: req.RetailPriceUSD,
		PricePerPieceUSD: req.PricePerPieceUSD, PricePerPackageUSD: req.PricePerPackageUSD,
		PiecesPerPackage: req.PiecesPerPackage, UnitName: req.UnitName, PackName: req.PackName,
		Stock: req.Stock, StockWarning: req.StockWarning,
		Description: req.Description, ImageURL: req.ImageURL,
		Img1: req.Img1, Img2: req.Img2, Img3: req.Img3, Img4: req.Img4, Img5: req.Img5,
		Category: req.Category, SortOrder: req.SortOrder,
		IsFeatured: req.IsFeatured, IsActive: req.IsActive,
		ProductionDate: parseDateFlexible(req.ProductionDate),
		ExpiryDate:     parseDateFlexible(req.ExpiryDate),
		// 新字段
		SupplierName: req.SupplierName, PrincipleCompany: req.PrincipleCompany,
		UnitWeightValue: req.UnitWeightValue, UnitWeightUnit: req.UnitWeightUnit,
		PackingFormat: req.PackingFormat, PackSize: req.PackSize,
		GpPercent: req.GpPercent, ShelfLifeDays: req.ShelfLifeDays,
		InnerPackPerCase: req.InnerPackPerCase, UnitPerInnerPack: req.UnitPerInnerPack, UnitPerCase: req.UnitPerCase,
		CostPerCase: req.CostPerCase, DcPercent: req.DcPercent,
		NetCostPerCase: req.NetCostPerCase, NetCostPerUnit: req.NetCostPerUnit,
		PriceInclVat: req.PriceInclVat, PriceExclVat: req.PriceExclVat,
		UnitWidthCm: req.UnitWidthCm, UnitLengthCm: req.UnitLengthCm, UnitHeightCm: req.UnitHeightCm, UnitWeightKg: req.UnitWeightKg,
		PackWidthCm: req.PackWidthCm, PackLengthCm: req.PackLengthCm, PackHeightCm: req.PackHeightCm, PackWeightKg: req.PackWeightKg,
		CaseWidthCm: req.CaseWidthCm, CaseLengthCm: req.CaseLengthCm, CaseHeightCm: req.CaseHeightCm, CaseWeightKg: req.CaseWeightKg,
		CreatedAt: now, UpdatedAt: now,
	}

	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "创建商品失败"})
		return
	}
	c.JSON(http.StatusCreated, buildProductResponse(&product))
}

// PATCH /api/products/:id
func UpdateProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var product models.Product
	if err := database.DB.Where("id = ? AND is_deleted = ?", id, false).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "商品不存在"})
		return
	}

	var req ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updates := map[string]interface{}{"updated_at": models.NowCambodia()}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.NameKh != nil {
		updates["name_kh"] = *req.NameKh
	}
	if req.NameEn != nil {
		updates["name_en"] = *req.NameEn
	}
	if req.Brand != nil {
		updates["brand"] = *req.Brand
	}
	if req.CountryOfOrigin != nil {
		updates["country_of_origin"] = *req.CountryOfOrigin
	}
	if req.Unit != nil {
		updates["unit"] = *req.Unit
	}
	if req.Specs != nil {
		updates["specs"] = *req.Specs
	}
	if req.Barcode != nil {
		updates["barcode"] = *req.Barcode
	}
	if req.PriceUSD != nil {
		updates["price_usd"] = *req.PriceUSD
	}
	if req.RetailPriceUSD != nil {
		updates["retail_price_usd"] = *req.RetailPriceUSD
	}
	if req.PricePerPieceUSD != nil {
		updates["price_per_piece_usd"] = *req.PricePerPieceUSD
	}
	if req.PricePerPackageUSD != nil {
		updates["price_per_package_usd"] = *req.PricePerPackageUSD
	}
	if req.PiecesPerPackage != nil {
		updates["pieces_per_package"] = *req.PiecesPerPackage
	}
	if req.UnitName != nil {
		updates["unit_name"] = *req.UnitName
	}
	if req.PackName != nil {
		updates["pack_name"] = *req.PackName
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.StockWarning != nil {
		updates["stock_warning"] = *req.StockWarning
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ImageURL != nil {
		updates["image_url"] = *req.ImageURL
	}
	if req.Img1 != nil {
		updates["img1"] = *req.Img1
	}
	if req.Img2 != nil {
		updates["img2"] = *req.Img2
	}
	if req.Img3 != nil {
		updates["img3"] = *req.Img3
	}
	if req.Img4 != nil {
		updates["img4"] = *req.Img4
	}
	if req.Img5 != nil {
		updates["img5"] = *req.Img5
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ProductionDate != nil {
		updates["production_date"] = parseDateFlexible(req.ProductionDate)
	}
	if req.ExpiryDate != nil {
		updates["expiry_date"] = parseDateFlexible(req.ExpiryDate)
	}
	// 供应商
	if req.SupplierName != nil {
		updates["supplier_name"] = *req.SupplierName
	}
	if req.PrincipleCompany != nil {
		updates["principle_company"] = *req.PrincipleCompany
	}
	// 基础扩展
	if req.UnitWeightValue != nil {
		updates["unit_weight_value"] = *req.UnitWeightValue
	}
	if req.UnitWeightUnit != nil {
		updates["unit_weight_unit"] = *req.UnitWeightUnit
	}
	if req.PackingFormat != nil {
		updates["packing_format"] = *req.PackingFormat
	}
	if req.PackSize != nil {
		updates["pack_size"] = *req.PackSize
	}
	if req.GpPercent != nil {
		updates["gp_percent"] = *req.GpPercent
	}
	if req.ShelfLifeDays != nil {
		updates["shelf_life_days"] = *req.ShelfLifeDays
	}
	// 包装规格层级
	if req.InnerPackPerCase != nil {
		updates["inner_pack_per_case"] = *req.InnerPackPerCase
	}
	if req.UnitPerInnerPack != nil {
		updates["unit_per_inner_pack"] = *req.UnitPerInnerPack
	}
	if req.UnitPerCase != nil {
		updates["unit_per_case"] = *req.UnitPerCase
	}
	// 成本与价格
	if req.CostPerCase != nil {
		updates["cost_per_case"] = *req.CostPerCase
	}
	if req.DcPercent != nil {
		updates["dc_percent"] = *req.DcPercent
	}
	if req.NetCostPerCase != nil {
		updates["net_cost_per_case"] = *req.NetCostPerCase
	}
	if req.NetCostPerUnit != nil {
		updates["net_cost_per_unit"] = *req.NetCostPerUnit
	}
	if req.PriceInclVat != nil {
		updates["price_incl_vat"] = *req.PriceInclVat
	}
	if req.PriceExclVat != nil {
		updates["price_excl_vat"] = *req.PriceExclVat
	}
	// 尺寸 最小包
	if req.UnitWidthCm != nil {
		updates["unit_width_cm"] = *req.UnitWidthCm
	}
	if req.UnitLengthCm != nil {
		updates["unit_length_cm"] = *req.UnitLengthCm
	}
	if req.UnitHeightCm != nil {
		updates["unit_height_cm"] = *req.UnitHeightCm
	}
	if req.UnitWeightKg != nil {
		updates["unit_weight_kg"] = *req.UnitWeightKg
	}
	// 尺寸 中包
	if req.PackWidthCm != nil {
		updates["pack_width_cm"] = *req.PackWidthCm
	}
	if req.PackLengthCm != nil {
		updates["pack_length_cm"] = *req.PackLengthCm
	}
	if req.PackHeightCm != nil {
		updates["pack_height_cm"] = *req.PackHeightCm
	}
	if req.PackWeightKg != nil {
		updates["pack_weight_kg"] = *req.PackWeightKg
	}
	// 尺寸 外箱
	if req.CaseWidthCm != nil {
		updates["case_width_cm"] = *req.CaseWidthCm
	}
	if req.CaseLengthCm != nil {
		updates["case_length_cm"] = *req.CaseLengthCm
	}
	if req.CaseHeightCm != nil {
		updates["case_height_cm"] = *req.CaseHeightCm
	}
	if req.CaseWeightKg != nil {
		updates["case_weight_kg"] = *req.CaseWeightKg
	}

	database.DB.Model(&product).Updates(updates)
	database.DB.First(&product, id)
	c.JSON(http.StatusOK, buildProductResponse(&product))
}

// DELETE /api/products/:id (软删除)
func DeleteProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var product models.Product
	if err := database.DB.Where("id = ? AND is_deleted = ?", id, false).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "商品不存在"})
		return
	}
	database.DB.Model(&product).Updates(map[string]interface{}{
		"is_deleted": true,
		"updated_at": models.NowCambodia(),
	})
	c.JSON(http.StatusOK, gin.H{"message": "商品已删除"})
}

// ─────────────────── 条码查询 ───────────────────

// GET /api/products/barcode/:barcode
func GetProductByBarcode(c *gin.Context) {
	barcode := c.Param("barcode")
	var product models.Product
	if err := database.DB.Where("barcode = ? AND is_deleted = ?", barcode, false).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "未找到该条码的商品"})
		return
	}
	c.JSON(http.StatusOK, buildProductResponse(&product))
}

// ─────────────────── 批量导入（存根） ───────────────────

// GET /api/products/import/template
func GetImportTemplate(c *gin.Context) {
	header := "name,name_kh,name_en,brand,unit,unit_name,pack_name,price_usd,price_per_piece_usd,price_per_package_usd,pieces_per_package,stock,stock_warning,barcode,category,description\n"
	c.Header("Content-Disposition", "attachment; filename=import_template.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(header))
}

// POST /api/products/import  — 批量导入（支持 CSV 和供应商 Excel .xlsx）
func ImportProducts(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "请上传文件"})
		return
	}
	defer file.Close()

	// ── 公共辅助函数 ──
	strPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	parseF := func(s string) *float64 {
		if s == "" {
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil
		}
		return &v
	}
	parseInt := func(s string) *int {
		if s == "" {
			return nil
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			if fv, ferr := strconv.ParseFloat(s, 64); ferr == nil {
				iv := int(fv)
				return &iv
			}
			return nil
		}
		return &v
	}

	now := models.NowCambodia()
	created, updated, skipped := 0, 0, 0
	errs := []string{}

	// ─────────────────────── XLSX 供应商格式 ───────────────────────
	if strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".xlsx") {
		xlsxFile, err := excelize.OpenReader(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Excel 解析失败: " + err.Error()})
			return
		}
		sheets := xlsxFile.GetSheetList()
		if len(sheets) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Excel 没有工作表"})
			return
		}
		rows, err := xlsxFile.GetRows(sheets[0])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "读取 Excel 数据失败: " + err.Error()})
			return
		}
		if len(rows) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Excel 数据不足（至少需要 4 行）"})
			return
		}

		// 第 1 行（index 0）：供应商名称，例如 "Supplier: 金磨坊"
		supplierName := ""
		if len(rows[0]) > 0 {
			cellA1 := strings.TrimSpace(rows[0][0])
			if idx := strings.Index(cellA1, ":"); idx >= 0 {
				supplierName = strings.TrimSpace(cellA1[idx+1:])
			} else if idx := strings.Index(cellA1, "："); idx >= 0 {
				supplierName = strings.TrimSpace(cellA1[idx+len("："):])
			} else {
				supplierName = cellA1
			}
		}

		// 第 2-3 行（index 1-2）：双层表头，跳过
		// 第 4 行起（index 3+）：数据行
		getCell := func(row []string, idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		parseFCell := func(row []string, idx int) *float64 {
			s := getCell(row, idx)
			if s == "" || s == "-" || s == "N/A" {
				return nil
			}
			s = strings.TrimSuffix(s, "%")
			return parseF(strings.TrimSpace(s))
		}
		parseIntCell := func(row []string, idx int) *int {
			s := getCell(row, idx)
			if s == "" || s == "-" || s == "N/A" {
				return nil
			}
			return parseInt(s)
		}
		// 百分比字段：Excel 中可能是 0.28 或 28，统一转换为 28.0
		parsePercent := func(row []string, idx int) *float64 {
			v := parseFCell(row, idx)
			if v == nil {
				return nil
			}
			if *v > 0 && *v < 1 {
				pct := *v * 100
				return &pct
			}
			return v
		}
		// 解析单位重量字符串，例如 "20g" → (20.0, "G")
		parseUnitWeight := func(s string) (*float64, *string) {
			s = strings.TrimSpace(s)
			if s == "" || s == "-" {
				return nil, nil
			}
			i := 0
			for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
				i++
			}
			if i == 0 {
				return nil, nil
			}
			v, err := strconv.ParseFloat(s[:i], 64)
			if err != nil {
				return nil, nil
			}
			if i >= len(s) {
				return &v, nil
			}
			u := strings.ToUpper(strings.TrimSpace(s[i:]))
			return &v, &u
		}

		for rowIdx, row := range rows[3:] {
			lineNo := rowIdx + 4
			nameZH := getCell(row, 7)
			nameEN := getCell(row, 5)
			nameKH := getCell(row, 6)
			name := nameZH
			if name == "" {
				name = nameEN
			}
			if name == "" {
				skipped++
				continue
			}

			// 条码：优先最小单位（col 3），退而中包（col 2），再退外箱（col 1）
			barcode := getCell(row, 3)
			if barcode == "" {
				barcode = getCell(row, 2)
			}
			if barcode == "" {
				barcode = getCell(row, 1)
			}

			// 价格：优先 net_cost_per_unit（col 17），退而 price_excl_vat（col 19）
			priceUSD := 0.0
			if pv := parseFCell(row, 17); pv != nil {
				priceUSD = *pv
			} else if pv := parseFCell(row, 19); pv != nil {
				priceUSD = *pv
			}
			if priceUSD <= 0 {
				errs = append(errs, fmt.Sprintf("第 %d 行 (%s): 价格无效，已跳过", lineNo, name))
				skipped++
				continue
			}

			uwVal, uwUnit := parseUnitWeight(getCell(row, 9))
			sortOrder := 0
			if sv := parseIntCell(row, 0); sv != nil {
				sortOrder = *sv
			}

			var barcodePtr *string
			if barcode != "" {
				barcodePtr = &barcode
			}

			var existing models.Product
			found := false
			if barcodePtr != nil {
				found = database.DB.Where("barcode = ? AND is_deleted = ?", *barcodePtr, false).First(&existing).Error == nil
			}

			if found {
				upd := map[string]interface{}{
					"name":       name,
					"price_usd":  priceUSD,
					"sort_order": sortOrder,
					"updated_at": now,
				}
				if nameEN != "" {
					upd["name_en"] = nameEN
				}
				if nameKH != "" {
					upd["name_kh"] = nameKH
				}
				if v := getCell(row, 4); v != "" {
					upd["brand"] = v
				}
				if supplierName != "" {
					upd["supplier_name"] = supplierName
				}
				if v := strPtr(getCell(row, 22)); v != nil {
					upd["principle_company"] = v
				}
				if v := strPtr(getCell(row, 23)); v != nil {
					upd["country_of_origin"] = v
				}
				if v := strPtr(getCell(row, 10)); v != nil {
					upd["packing_format"] = v
				}
				if uwVal != nil {
					upd["unit_weight_value"] = uwVal
				}
				if uwUnit != nil {
					upd["unit_weight_unit"] = uwUnit
				}
				if v := parseIntCell(row, 11); v != nil {
					upd["inner_pack_per_case"] = v
				}
				if v := parseIntCell(row, 12); v != nil {
					upd["unit_per_inner_pack"] = v
				}
				if v := parseIntCell(row, 13); v != nil {
					upd["unit_per_case"] = v
				}
				if v := parseFCell(row, 14); v != nil {
					upd["cost_per_case"] = v
				}
				if v := parsePercent(row, 15); v != nil {
					upd["dc_percent"] = v
				}
				if v := parseFCell(row, 16); v != nil {
					upd["net_cost_per_case"] = v
				}
				if v := parseFCell(row, 17); v != nil {
					upd["net_cost_per_unit"] = v
				}
				if v := parseFCell(row, 18); v != nil {
					upd["price_incl_vat"] = v
				}
				if v := parseFCell(row, 19); v != nil {
					upd["price_excl_vat"] = v
				}
				if v := parsePercent(row, 20); v != nil {
					upd["gp_percent"] = v
				}
				if v := parseIntCell(row, 21); v != nil {
					upd["shelf_life_days"] = v
				}
				if v := parseFCell(row, 24); v != nil {
					upd["unit_width_cm"] = v
				}
				if v := parseFCell(row, 25); v != nil {
					upd["unit_length_cm"] = v
				}
				if v := parseFCell(row, 26); v != nil {
					upd["unit_height_cm"] = v
				}
				if v := parseFCell(row, 27); v != nil {
					upd["unit_weight_kg"] = v
				}
				if v := parseFCell(row, 28); v != nil {
					upd["pack_width_cm"] = v
				}
				if v := parseFCell(row, 29); v != nil {
					upd["pack_length_cm"] = v
				}
				if v := parseFCell(row, 30); v != nil {
					upd["pack_height_cm"] = v
				}
				if v := parseFCell(row, 31); v != nil {
					upd["pack_weight_kg"] = v
				}
				if v := parseFCell(row, 32); v != nil {
					upd["case_width_cm"] = v
				}
				if v := parseFCell(row, 33); v != nil {
					upd["case_length_cm"] = v
				}
				if v := parseFCell(row, 34); v != nil {
					upd["case_height_cm"] = v
				}
				if v := parseFCell(row, 35); v != nil {
					upd["case_weight_kg"] = v
				}
				if dbErr := database.DB.Model(&existing).Updates(upd).Error; dbErr != nil {
					errs = append(errs, fmt.Sprintf("第 %d 行 (%s): 更新失败 %v", lineNo, name, dbErr))
					continue
				}
				updated++
			} else {
				product := models.Product{
					Name:             name,
					NameEn:           strPtr(nameEN),
					NameKh:           strPtr(nameKH),
					Brand:            strPtr(getCell(row, 4)),
					Barcode:          barcodePtr,
					PriceUSD:         priceUSD,
					SortOrder:        sortOrder,
					Unit:             "件",
					IsActive:         true,
					CreatedAt:        now,
					UpdatedAt:        now,
					SupplierName:     strPtr(supplierName),
					PrincipleCompany: strPtr(getCell(row, 22)),
					CountryOfOrigin:  strPtr(getCell(row, 23)),
					PackingFormat:    strPtr(getCell(row, 10)),
					UnitWeightValue:  uwVal,
					UnitWeightUnit:   uwUnit,
					InnerPackPerCase: parseIntCell(row, 11),
					UnitPerInnerPack: parseIntCell(row, 12),
					UnitPerCase:      parseIntCell(row, 13),
					CostPerCase:      parseFCell(row, 14),
					DcPercent:        parsePercent(row, 15),
					NetCostPerCase:   parseFCell(row, 16),
					NetCostPerUnit:   parseFCell(row, 17),
					PriceInclVat:     parseFCell(row, 18),
					PriceExclVat:     parseFCell(row, 19),
					GpPercent:        parsePercent(row, 20),
					ShelfLifeDays:    parseIntCell(row, 21),
					UnitWidthCm:      parseFCell(row, 24),
					UnitLengthCm:     parseFCell(row, 25),
					UnitHeightCm:     parseFCell(row, 26),
					UnitWeightKg:     parseFCell(row, 27),
					PackWidthCm:      parseFCell(row, 28),
					PackLengthCm:     parseFCell(row, 29),
					PackHeightCm:     parseFCell(row, 30),
					PackWeightKg:     parseFCell(row, 31),
					CaseWidthCm:      parseFCell(row, 32),
					CaseLengthCm:     parseFCell(row, 33),
					CaseHeightCm:     parseFCell(row, 34),
					CaseWeightKg:     parseFCell(row, 35),
				}
				if dbErr := database.DB.Create(&product).Error; dbErr != nil {
					errs = append(errs, fmt.Sprintf("第 %d 行 (%s): 创建失败 %v", lineNo, name, dbErr))
					continue
				}
				created++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"created": created,
			"updated": updated,
			"skipped": skipped,
			"errors":  errs,
		})
		return
	}

	// ─────────────────────── CSV 格式（原有逻辑）───────────────────────
	reader := csv.NewReader(file)
	csvRows, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "CSV 解析失败: " + err.Error()})
		return
	}
	if len(csvRows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "CSV 文件没有数据行"})
		return
	}

	// 第一行是 header，建立列名→索引映射
	csvHeader := csvRows[0]
	col := make(map[string]int)
	for i, h := range csvHeader {
		col[strings.TrimSpace(strings.ToLower(h))] = i
	}
	get := func(row []string, name string) string {
		if i, ok := col[name]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}

	for i, row := range csvRows[1:] {
		name := get(row, "name")
		if name == "" {
			errs = append(errs, fmt.Sprintf("第 %d 行: name 不能为空", i+2))
			continue
		}

		priceStr := get(row, "price_usd")
		priceUSD, pErr := strconv.ParseFloat(priceStr, 64)
		if pErr != nil || priceUSD <= 0 {
			errs = append(errs, fmt.Sprintf("第 %d 行: price_usd 无效", i+2))
			continue
		}

		barcode := strPtr(get(row, "barcode"))
		unit := get(row, "unit")
		if unit == "" {
			unit = "件"
		}
		stockVal := 0
		if sv := parseInt(get(row, "stock")); sv != nil {
			stockVal = *sv
		}
		warnVal := 0
		if wv := parseInt(get(row, "stock_warning")); wv != nil {
			warnVal = *wv
		}

		var existing models.Product
		found := false
		if barcode != nil && *barcode != "" {
			found = database.DB.Where("barcode = ? AND is_deleted = ?", *barcode, false).First(&existing).Error == nil
		}

		if found {
			upd := map[string]interface{}{
				"name":          name,
				"price_usd":     priceUSD,
				"stock":         stockVal,
				"stock_warning": warnVal,
				"unit":          unit,
				"updated_at":    now,
			}
			if v := strPtr(get(row, "name_kh")); v != nil {
				upd["name_kh"] = v
			}
			if v := strPtr(get(row, "name_en")); v != nil {
				upd["name_en"] = v
			}
			if v := strPtr(get(row, "brand")); v != nil {
				upd["brand"] = v
			}
			if v := strPtr(get(row, "category")); v != nil {
				upd["category"] = v
			}
			if v := strPtr(get(row, "description")); v != nil {
				upd["description"] = v
			}
			if v := parseF(get(row, "price_per_piece_usd")); v != nil {
				upd["price_per_piece_usd"] = v
			}
			if v := parseF(get(row, "price_per_package_usd")); v != nil {
				upd["price_per_package_usd"] = v
			}
			if v := parseInt(get(row, "pieces_per_package")); v != nil {
				upd["pieces_per_package"] = v
			}
			if v := strPtr(get(row, "unit_name")); v != nil {
				upd["unit_name"] = v
			}
			if v := strPtr(get(row, "pack_name")); v != nil {
				upd["pack_name"] = v
			}
			database.DB.Model(&existing).Updates(upd)
			updated++
		} else {
			product := models.Product{
				Name:               name,
				NameKh:             strPtr(get(row, "name_kh")),
				NameEn:             strPtr(get(row, "name_en")),
				Brand:              strPtr(get(row, "brand")),
				Unit:               unit,
				UnitName:           strPtr(get(row, "unit_name")),
				PackName:           strPtr(get(row, "pack_name")),
				Barcode:            barcode,
				PriceUSD:           priceUSD,
				PricePerPieceUSD:   parseF(get(row, "price_per_piece_usd")),
				PricePerPackageUSD: parseF(get(row, "price_per_package_usd")),
				PiecesPerPackage:   parseInt(get(row, "pieces_per_package")),
				Stock:              stockVal,
				StockWarning:       warnVal,
				Category:           strPtr(get(row, "category")),
				Description:        strPtr(get(row, "description")),
				IsActive:           true,
				CreatedAt:          now,
				UpdatedAt:          now,
			}
			if dbErr := database.DB.Create(&product).Error; dbErr != nil {
				errs = append(errs, fmt.Sprintf("第 %d 行: 创建失败 %v", i+2, dbErr))
				continue
			}
			created++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"created": created,
		"updated": updated,
		"skipped": skipped,
		"errors":  errs,
	})
}

// GET /api/products/expiring?days=30  — 管理员查询临近过期商品
func ListExpiringProducts(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	threshold := time.Now().AddDate(0, 0, days)
	var products []models.Product
	database.DB.Where("is_deleted = ? AND expiry_date IS NOT NULL AND expiry_date <= ?", false, threshold).
		Order("expiry_date ASC").Find(&products)

	result := make([]ProductResponse, len(products))
	for i, p := range products {
		p := p
		result[i] = buildProductResponse(&p)
	}
	c.JSON(http.StatusOK, gin.H{"products": result, "days": days, "threshold": threshold.Format("2006-01-02")})
}
