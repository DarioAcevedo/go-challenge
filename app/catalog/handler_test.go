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

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/mytheresa/go-hiring-challenge/models/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCatalogHandler_HandleGet(t *testing.T) {
	var mockProducts []models.Product
	content, _:= os.ReadFile("testdata/products.json")
	_ = json.Unmarshal(content, &mockProducts)
	var tests = []struct {
		name string
		limit string
		offset string
		expectedLimit int
		expectedOffset int
		mockProducts []models.Product
		mockError error
		expectedError error
		expectedStatus int
	}{
		{
			"everything works fine",
			"10",
			"2",
			10,
			2,
			mockProducts,
			nil,
			nil,
			200,
		},
		{
			"no limit, no offset",
			"",
			"",
			10,
			0,
			mockProducts,
			nil,
			nil,
			200,
		},
		{
			"lower limit than allowed",
			"0",
			"",
			1,
			0,
			mockProducts,
			nil,
			nil,
			200,
		},
				{
			"greater limit than allowed",
			"1000",
			"",
			100,
			0,
			mockProducts,
			nil,
			nil,
			200,
		},
		{
			"bad limit and offset",
			"abc",
			"abcd",
			0,
			0,
			mockProducts,
			nil,
			fmt.Errorf("Limit must be an integer\n"),
			400,
		},
		{
			"repo list products return error",
			"",
			"",
			0,
			0,
			[]models.Product{},
			fmt.Errorf("database error"),
			fmt.Errorf("database error\n"),
			500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mock.MockProductsRepository{
				MockListProducts: func(limit int, offset int) ([]models.Product, error) {
					return tt.mockProducts, tt.mockError
				},
			}
			handler := NewCatalogHandler(mockRepo)
			req := createRequest("GET", "/catalog?limit="+tt.limit+"&offset="+tt.offset, ``)
			w := httptest.NewRecorder()
			handler.HandleGet(w, req)
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedError != nil {
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, tt.expectedError.Error(), string(body))
				return
			}
			
			var responseBody Response
			err := json.NewDecoder(resp.Body).Decode(&responseBody)
			if err != nil {
				t.Fatalf("could not decode response: %v", err)
			}
			assert.Equal(t, tt.expectedLimit, mockRepo.CalledLimit)

		})
	}
}

func createRequest(method string, path string, body string) *http.Request { 
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	return req
}