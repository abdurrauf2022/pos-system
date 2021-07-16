package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductFormatter formats a product object.
func ProductFormatter(product *Product) interface{} {
	if product == nil {
		return nil
	}

	return gin.H{
		"id":         product.ID,
		"name":       product.Name,
		"price":      product.Price,
		"type":       product.Type,
		"created_at": product.CreatedAt,
	}
}

// ProductHandlers defines the handlers for all of the products.
type ProductHandlers struct {
	DB *gorm.DB
}

// CreateProduct creates a product.
func (ph *ProductHandlers) CreateProduct(c *gin.Context) {
	name := c.PostForm("name")
	priceStr := c.PostForm("price")
	productType := c.PostForm("type")

	apiResponse := ApiResponse{}

	if name == "" || priceStr == "" || productType == "" {
		apiResponse.Success = false
		apiResponse.Error = "Invalid inputs"
		c.JSON(http.StatusOK, apiResponse)
		return
	}

	if productType != "food" && productType != "drink" && productType != "pastry" {
		apiResponse.Success = false
		apiResponse.Error = "Invalid product type"
		c.JSON(http.StatusOK, apiResponse)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)

	if err != nil {
		apiResponse.Success = false
		apiResponse.Error = err.Error()
		c.JSON(http.StatusOK, apiResponse)
		return
	}

	ph.DB.Create(&Product{
		Name:  name,
		Price: price,
		Type:  productType,
	}).Commit()

	apiResponse.Success = true

	c.JSON(http.StatusOK, apiResponse)
}

// ListProducts returns a complete list of products.
func (ph *ProductHandlers) ListProducts(c *gin.Context) {
	var productObjs []*Product
	var products []interface{}

	// The GET param that tells us to get all products or not.
	allProducts := c.Query("all") != ""

	if allProducts == true {
		ph.DB.Order("type asc").Find(&productObjs)
	} else {
		ph.DB.Where("discontinued = 0").
			Order("type asc").
			Find(&productObjs)
	}

	for _, product := range productObjs {
		products = append(products, ProductFormatter(product))
	}

	apiResponse := ApiResponse{
		Data:    products,
		Success: true,
	}

	c.JSON(http.StatusOK, apiResponse)
}
