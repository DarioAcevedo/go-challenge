package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"github.com/golang/mock/gomock"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/mytheresa/go-hiring-challenge/models/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/shopspring/decimal"
)

func TestCatalogHandler_HandleGet(t *testing.T) {
	var mockProducts []models.Product
	content, _:= os.ReadFile("testdata/products.json")
	_ = json.Unmarshal(content, &mockProducts)
	var tests = []struct {
		name string
		limit string
		offset string
		price_lt string
		category string
		expectedFilters models.ProductFilters
		expectedError error
		expectedStatus int
	}{
		{
			"everything works fine",
			"10",
			"2",
			"10",
			"Clothing",
			models.ProductFilters{
				Limit: 10,
				Offset: 2,
				ProductCategory: Ptr("Clothing"),
				PriceLt: Ptr(decimal.New(10, 0)),
			},
			nil,
			200,
		},
		{
			"no limit, no offset",
			"",
			"",
			"10",
			"Clothing",
			models.ProductFilters{
				Limit: DefaultLimit,
				Offset: 0,
				ProductCategory: Ptr("Clothing"),
				PriceLt: Ptr(decimal.New(10, 0)),
			},
			nil,
			200,
		},
		{
			"lower limit than allowed",
			"0",
			"",
			"10",
			"Clothing",
			models.ProductFilters{
				Limit: MinLimit,
				Offset: 0,
				ProductCategory: Ptr("Clothing"),
				PriceLt: Ptr(decimal.New(10, 0)),
			},
			nil,
			200,
		},
				{
			"greater limit than allowed",
			"1000",
			"",
			"10",
			"Clothing",
			models.ProductFilters{
				Limit: MaxLimit,
				Offset: 0,
				ProductCategory: Ptr("Clothing"),
				PriceLt: Ptr(decimal.New(10, 0)),
			},
			nil,
			200,
		},
		{
			"bad limit and offset",
			"abc",
			"abcd",
			"10",
			"Clothing",
			models.ProductFilters{},
			fmt.Errorf("Limit must be an integer\n"),
			400,
		},
		{
			"repo list products return error",
			"",
			"",
			"10",
			"Clothing",
			models.ProductFilters{
				Limit: DefaultLimit,
				Offset: 0,
				ProductCategory: Ptr("Clothing"),
				PriceLt: Ptr(decimal.New(10, 0)),
			},
			fmt.Errorf("database error"),
			500,
		},
		{
			"bad price_lt filter",
			"",
			"",
			"abcd",
			"Clothing",
			models.ProductFilters{
				Limit: DefaultLimit,
				Offset: 0,
				ProductCategory: Ptr("Clothing"),
				PriceLt: Ptr(decimal.New(10, 0)),
			},
			fmt.Errorf("price_lt must be a valid decimal number"),
			400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtl:= gomock.NewController(t)
			defer mockCtl.Finish()
			mockProductsRepo := mock.NewMockProductsRepository(mockCtl)
			if tt.expectedStatus == 400 {
				mockProductsRepo.EXPECT().CountProducts(gomock.Any()).Times(0)
				mockProductsRepo.EXPECT().ListProducts(gomock.Any()).Times(0)
			}
			if tt.expectedStatus == 500 {
				mockProductsRepo.EXPECT().CountProducts(&tt.expectedFilters).Return(int64(0), tt.expectedError).Times(0)
				mockProductsRepo.EXPECT().ListProducts(gomock.Any()).Return([]models.Product{}, tt.expectedError).Times(1)
			}
			if tt.expectedStatus == 200 {
				mockProductsRepo.EXPECT().CountProducts(gomock.Any()).DoAndReturn(
					func(f *models.ProductFilters) (int64, error) {
						assert.Equal(t, tt.expectedFilters.Limit, f.Limit)
						assert.Equal(t, tt.expectedFilters.Offset, f.Offset)
						if tt.expectedFilters.ProductCategory != nil {
							assert.Equal(t, *tt.expectedFilters.ProductCategory, *f.ProductCategory)
						}
						if tt.expectedFilters.PriceLt != nil {
							assert.Equal(t, *tt.expectedFilters.PriceLt, *f.PriceLt)
						}
						return int64(len(mockProducts)), nil
					}).Times(1)
				mockProductsRepo.EXPECT().ListProducts(gomock.Any()).DoAndReturn(
					func(f *models.ProductFilters) ([]models.Product, error) {
						assert.Equal(t, tt.expectedFilters.Limit, f.Limit)
						assert.Equal(t, tt.expectedFilters.Offset, f.Offset)
						if tt.expectedFilters.ProductCategory != nil {
							assert.Equal(t, *tt.expectedFilters.ProductCategory, *f.ProductCategory)
						}
						if tt.expectedFilters.PriceLt != nil {
							assert.Equal(t, *tt.expectedFilters.PriceLt, *f.PriceLt)
						}
						return mockProducts, nil
					}).Times(1)
			}
			handler := NewCatalogHandler(mockProductsRepo)
			req := createRequest(
				"GET", 
				"/catalog?limit="+tt.limit+"&offset="+tt.offset+"&category="+tt.category+"&price_lt="+tt.price_lt, 
				``)
			w := httptest.NewRecorder()
			handler.HandleGet(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedError != nil {
				body, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(body), tt.expectedError.Error())
				return
			}
			
			var responseBody Response
			err := json.NewDecoder(resp.Body).Decode(&responseBody)
			if err != nil {
				t.Fatalf("could not decode response: %v", err)
			}
			assert.NotEmpty(t, responseBody.TotalProductCount)
		})
	}
}

func TestCatalogHandler_HandleDetails(t *testing.T) {
	var mockProducts []models.Product
	content, _:= os.ReadFile("testdata/products.json")
	_ = json.Unmarshal(content, &mockProducts)
	var tests = []struct {
		name string
		productCode string
		expectedProduct *models.Product
		expectedError error
		expectedStatus int
	}{
		{
			"product found",
			"PROD001",
			&mockProducts[0],
			nil,
			200,
		},
		{
			"product code missing",
			"",
			nil,
			fmt.Errorf("product code is required"),
			400,
		},
		{
			"product not found",
			"UNKNOWN",
			nil,
			fmt.Errorf("product not found"),
			404,
		},
		{
			"repo get product returns error",
			"PROD001",
			nil,
			fmt.Errorf("database error"),
			500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtl:= gomock.NewController(t)
			defer mockCtl.Finish()
			mockProductsRepo := mock.NewMockProductsRepository(mockCtl)
			if tt.expectedStatus == 500 {
				mockProductsRepo.
				EXPECT().
				GetProductByCode(tt.productCode).
				Return(tt.expectedProduct, tt.expectedError).
				Times(1)
			} else if tt.expectedStatus == 400 {
				mockProductsRepo.
				EXPECT().
				GetProductByCode(gomock.Any()).
				Times(0)
			} else {
				mockProductsRepo.
				EXPECT().
				GetProductByCode(tt.productCode).
				Return(tt.expectedProduct, nil).
				Times(1)
			}
			
			handler := NewCatalogHandler(mockProductsRepo)
			req := createRequest(
				"GET", 
				"/catalog/"+tt.productCode, 
				``)
			req.SetPathValue("code", tt.productCode)
			w := httptest.NewRecorder()
			handler.HandleDetails(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedError != nil {
				body, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(body), tt.expectedError.Error())
				return
			}
			var responseBody ProductDetails
			err := json.NewDecoder(resp.Body).Decode(&responseBody)
			if err != nil {
				t.Fatalf("could not decode response: %v", err)
			}
			assert.Equal(t, tt.expectedProduct.Code, responseBody.Code)
			for _, variant := range responseBody.Variants {
				fmt.Println(variant)
				comp := variant.Price.Cmp(decimal.Zero)
				assert.Greater(t, comp, 0)
			}
		})
	}
}

func createRequest(method string, path string, body string) *http.Request { 
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	return req
}

func Ptr[T any](v T) *T { return &v }