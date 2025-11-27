package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

const (
	DefaultLimit  = 10
	MinLimit = 1
	MaxLimit = 100
)

type Response struct {
	Products []Product `json:"products"`
	TotalProductCount int `json:"available_products"`
	Next int `json:"next"` 
}

type Product struct {
	Code  string  `json:"code"`
	Price float64 `json:"price"`
	Category string `json:"category"`
}

type ProductVariant struct {
	SKU  string  `json:"sku"`
	Price decimal.Decimal `json:"price"`
	Name string `json:"name"`
}

type ProductDetails struct {
	Code     string           `json:"code"`
	Price    float64          `json:"price"`
	Category string           `json:"category"`
	Variants []ProductVariant `json:"variants"`
}

type CatalogHandler struct {
	repo models.ProductsRepository
}

func NewCatalogHandler(r models.ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	// parse queryparams
	queryParmas:= r.URL.Query()
	filters, err := getHandlerParams(&queryParmas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := h.repo.ListProducts(&filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
			Category: p.Category.Name,
		}
	}
	productCount, err := h.repo.CountProducts(&filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var nextOffset int
	if filters.Offset+filters.Limit < int(productCount) {
		nextOffset = filters.Offset + filters.Limit + 1
	}

	// Return the products as a JSON response
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Products: products,
		TotalProductCount: int(productCount),
		Next: nextOffset,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *CatalogHandler) HandleDetails(w http.ResponseWriter, r *http.Request) {
	// parse queryparams
	productCode := r.PathValue("code")
	if productCode == "" {
		http.Error(w, "product code is required", http.StatusBadRequest)
		return
	}
	product, err := h.repo.GetProductByCode(productCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if product == nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	
	// Map response
	response := ProductDetails {
		Code:  product.Code,
		Price: product.Price.InexactFloat64(),
		Category: product.Category.Name,
	}
	for _, variant := range product.Variants {
		pv := ProductVariant{
			SKU: variant.SKU,
			Name: variant.Name,
			Price: variant.Price,
		}
		if cmp := variant.Price.Cmp(decimal.Zero); cmp == 0 {
			pv.Price = product.Price
		}
		response.Variants = append(response.Variants, pv)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getHandlerParams(params *url.Values) (models.ProductFilters, error) {
	limit, offset, err := getPaginationParams(params)
	if err != nil {
		return models.ProductFilters{}, err
	}
	var category *string
	if cat := params.Get("category"); cat != "" {
		category = &cat
	}
	var priceLte *decimal.Decimal
	if priceStr := params.Get("price_lt"); priceStr != "" {
		priceDec, err := decimal.NewFromString(priceStr)
		if err != nil {
			return models.ProductFilters{}, fmt.Errorf("price_lt must be a valid decimal number")
		}
		priceLte = &priceDec
	}
	return models.ProductFilters{
		ProductCategory: category,
		PriceLt: priceLte,
		Offset: offset,
		Limit: limit,
	}, nil
}

func getPaginationParams(params *url.Values) (limit, offset int, err error) {
	offsetStr := params.Get("offset")
	limitStr := params.Get("limit")
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return limit, offset, fmt.Errorf("Limit must be an integer")
		}
		limit = sanitizeLimit(limit)
	} else {
		limit = DefaultLimit
	}
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return limit, offset, fmt.Errorf("Offset must be an integer")
		}
	} else {
		offset = 0
	}
	return
}

func sanitizeLimit(inLimit int) (int){
	if inLimit > MaxLimit {
		return MaxLimit

	}
	if inLimit < MinLimit {
		return MinLimit
	}
	return inLimit
}