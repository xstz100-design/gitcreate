package handlers

import (
	"net/http"
	"strconv"
	"time"
	"wholesale/database"
	"wholesale/middleware"
	"wholesale/models"
	"wholesale/services"

	"github.com/gin-gonic/gin"
)

// ─────────────────── 请求/响应结构 ───────────────────

type ProductCreateRequest struct {
	Name               string     `json:"name" binding:"required"`
	NameKh             *string    `json:"name_kh"`
	NameEn             *string    `json:"name_en"`
	Brand              *string    `json:"brand"`
	CountryOfOrigin    *string    `json:"country_of_origin"`
	Unit               string     `json:"unit"`
	Specs              *string    `json:"specs"`
	Barcode            *string    `json:"barcode"`
	PriceUSD           float64    `json:"price_usd" binding:"required,gt=0"`
	RetailPriceUSD     *float64   `json:"retail_price_usd"`
	PricePerPieceUSD   *float64   `json:"price_per_piece_usd"`
	PricePerPackageUSD *float64   `json:"price_per_package_usd"`
	PiecesPerPackage   *int       `json:"pieces_per_package"`
	Stock              int        `json:"stock"`
	StockWarning       int        `json:"stock_warning"`
	Description        *string    `json:"description"`
	ImageURL           *string    `json:"image_url"`
	Img1               *string    `json:"img1"`
	Img2               *string    `json:"img2"`
	Img3               *string    `json:"img3"`
	Img4               *string    `json:"img4"`
	Img5               *string    `json:"img5"`
	Category           *string    `json:"category"`
	SortOrder          int        `json:"sort_order"`
	IsFeatured         bool       `json:"is_featured"`
	IsActive           bool       `json:"is_active"`
	ProductionDate     *time.Time `json:"production_date"`
	ExpiryDate         *time.Time `json:"expiry_date"`
}

type ProductUpdateRequest struct {
	Name               *string    `json:"name"`
	NameKh             *string    `json:"name_kh"`
	NameEn             *string    `json:"name_en"`
	Brand              *string    `json:"brand"`
	CountryOfOrigin    *string    `json:"country_of_origin"`
	Unit               *string    `json:"unit"`
	Specs              *string    `json:"specs"`
	Barcode            *string    `json:"barcode"`
	PriceUSD           *float64   `json:"price_usd"`
	RetailPriceUSD     *float64   `json:"retail_price_usd"`
	PricePerPieceUSD   *float64   `json:"price_per_piece_usd"`
	PricePerPackageUSD *float64   `json:"price_per_package_usd"`
	PiecesPerPackage   *int       `json:"pieces_per_package"`
	Stock              *int       `json:"stock"`
	StockWarning       *int       `json:"stock_warning"`
	Description        *string    `json:"description"`
	ImageURL           *string    `json:"image_url"`
	Img1               *string    `json:"img1"`
	Img2               *string    `json:"img2"`
	Img3               *string    `json:"img3"`
	Img4               *string    `json:"img4"`
	Img5               *string    `json:"img5"`
	Category           *string    `json:"category"`
	SortOrder          *int       `json:"sort_order"`
	IsFeatured         *bool      `json:"is_featured"`
	IsActive           *bool      `json:"is_active"`
	ProductionDate     *time.Time `json:"production_date"`
	ExpiryDate         *time.Time `json:"expiry_date"`
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
}

func buildProductResponse(p *models.Product) ProductResponse {
	return ProductResponse{
		ID: p.ID, Name: p.Name, NameKh: p.NameKh, NameEn: p.NameEn,
		Brand: p.Brand, CountryOfOrigin: p.CountryOfOrigin,
		Unit: p.Unit, Specs: p.Specs, Barcode: p.Barcode,
		PriceUSD: p.PriceUSD, RetailPriceUSD: p.RetailPriceUSD,
		PricePerPieceUSD: p.PricePerPieceUSD, PricePerPackageUSD: p.PricePerPackageUSD,
		PiecesPerPackage: p.PiecesPerPackage,
		Stock:            p.Stock, StockWarning: p.StockWarning, IsLowStock: p.IsLowStock(),
		Description: p.Description, ImageURL: p.ImageURL,
		Img1: p.Img1, Img2: p.Img2, Img3: p.Img3, Img4: p.Img4, Img5: p.Img5,
		Category: p.Category, SortOrder: p.SortOrder, IsActive: p.IsActive,
		IsFeatured: p.IsFeatured, ProductionDate: p.ProductionDate,
		ExpiryDate: p.ExpiryDate, UpdatedAt: p.UpdatedAt,
	}
}

// ─────────────────── CRUD ───────────────────

// GET /api/products
func ListProducts(c *gin.Context) {
	services.TrackPageView()

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
		PiecesPerPackage: req.PiecesPerPackage,
		Stock:            req.Stock, StockWarning: req.StockWarning,
		Description: req.Description, ImageURL: req.ImageURL,
		Img1: req.Img1, Img2: req.Img2, Img3: req.Img3, Img4: req.Img4, Img5: req.Img5,
		Category: req.Category, SortOrder: req.SortOrder,
		IsFeatured: req.IsFeatured, IsActive: req.IsActive,
		ProductionDate: req.ProductionDate, ExpiryDate: req.ExpiryDate,
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
		updates["production_date"] = *req.ProductionDate
	}
	if req.ExpiryDate != nil {
		updates["expiry_date"] = *req.ExpiryDate
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
	header := "name,name_kh,name_en,brand,unit,price_usd,stock,barcode,category\n"
	c.Header("Content-Disposition", "attachment; filename=import_template.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(header))
}

// POST /api/products/import
func ImportProducts(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"detail": "批量导入功能暂未开放"})
}
