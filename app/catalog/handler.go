package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/models"
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
	params:= r.URL.Query()
	limit, offset, err := getPaginationParams(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := h.repo.ListProducts(limit, offset)
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
			Category: strconv.Itoa(int(p.ProductCategoryID)),
		}
	}
	productCount, err := h.repo.CountProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var nextOffset int
	if offset+limit < int(productCount) {
		nextOffset = offset + limit + 1
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