package categories

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/mytheresa/go-hiring-challenge/models/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCategoriesHandler_HandleGet(t *testing.T) {
	mockCategories := []models.ProductCategory{
		{
			ID:   1,
			Name: "Clothing",
			Code: "CLOTH",
		},
		{
			ID:   2,
			Name: "Footwear",
			Code: "FOOT",
		},
	}

	var tests = []struct {
		name           string
		mockResponse []models.ProductCategory
		expectedError  error
		expectedStatus int
		expectedBody   ListCategoriesResponse
	}{
		{
			"everything works fine",
			mockCategories,
			nil,
			200,
			ListCategoriesResponse{
				Categories: []Category{
					{Name: "Clothing", Code: "CLOTH"},
					{Name: "Footwear", Code: "FOOT"},
				},
			},
		},
		{
			"repository returns error",
			[]models.ProductCategory{},
			fmt.Errorf("database error"),
			500,
			ListCategoriesResponse{},
		},
		{
			"no categories available",
			[]models.ProductCategory{},
			nil,
			200,
			ListCategoriesResponse{
				Categories: []Category{},
			},
		},
		}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()
			
			mockCategoriesRepo := mock.NewMockCategoriesRepository(mockCtl)
			if tt.expectedStatus == 200 {
				mockCategoriesRepo.EXPECT().
				ListCategories().
				Return(tt.mockResponse, tt.expectedError).
				Times(1)
			}
			if tt.expectedStatus == 500 {
				mockCategoriesRepo.EXPECT().
				ListCategories().
				Return([]models.ProductCategory{}, tt.expectedError).
				Times(1)
			}
			handler := NewCategoriesHandler(mockCategoriesRepo)
			req := httptest.NewRequest("GET", "/categories", nil)
			w := httptest.NewRecorder()
			handler.HandleGet(w, req)
			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.expectedStatus == 200 {
				var responseBody ListCategoriesResponse
				if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				for idx, cat := range responseBody.Categories {
					assert.Equal(t, tt.expectedBody.Categories[idx], cat)
				}
			}
		})
	}
}


func TestCategoriesHandler_HandleCreate(t *testing.T) {
    var tests = []struct {
        name            string
        body            string
        mockCategory    *models.ProductCategory
        mockError       error
        expectedStatus  int
        expectedMessage string
    }{
        {
            name:           "successful creation",
            body:           `{"name": "Clothing", "code": "CLOTH"}`,
            mockCategory:   &models.ProductCategory{ID: 1, Name: "Clothing", Code: "CLOTH"},
            mockError:      nil,
            expectedStatus: http.StatusOK,
        },
        {
            name:            "invalid json",
            body:            `{invalid json`,
            mockCategory:    nil,
            mockError:       nil,
            expectedStatus:  http.StatusBadRequest,
            expectedMessage: "invalid request body",
        },
        {
            name:            "missing fields",
            body:            `{"name": ""}`,
            mockCategory:    nil,
            mockError:       nil,
            expectedStatus:  http.StatusBadRequest,
            expectedMessage: "name and code are required",
        },
        {
            name:           "code already exists",
            body:           `{"name": "Clothing", "code": "CLOTH"}`,
            mockCategory:   nil,
            mockError:      models.ErrCategoryCodeExists,
            expectedStatus: http.StatusConflict,
            expectedMessage: models.ErrCategoryCodeExists.Error(),
        },
        {
            name:           "name already exists",
            body:           `{"name": "Clothing", "code": "CLOTH"}`,
            mockCategory:   nil,
            mockError:      models.ErrCategoryNameExists,
            expectedStatus: http.StatusConflict,
            expectedMessage: models.ErrCategoryNameExists.Error(),
        },
        {
            name:           "repository error",
            body:           `{"name": "Clothing", "code": "CLOTH"}`,
            mockCategory:   nil,
            mockError:      fmt.Errorf("database error"),
            expectedStatus: http.StatusInternalServerError,
            expectedMessage: "database error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockCtl := gomock.NewController(t)
            defer mockCtl.Finish()

            mockCategoriesRepo := mock.NewMockCategoriesRepository(mockCtl)

            if tt.expectedStatus == http.StatusOK ||
                tt.expectedStatus == http.StatusConflict ||
                tt.expectedStatus == http.StatusInternalServerError {

                mockCategoriesRepo.EXPECT().
                    CreateCategory("CLOTH", "Clothing").
                    Return(tt.mockCategory, tt.mockError).
                    Times(1)
            }

            handler := NewCategoriesHandler(mockCategoriesRepo)

            req := httptest.NewRequest("POST", "/categories", strings.NewReader(tt.body))
            w := httptest.NewRecorder()

            handler.HandleCreate(w, req)

            resp := w.Result()

            if resp.StatusCode != tt.expectedStatus {
                t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
            }

            if tt.expectedStatus == http.StatusOK {
                var responseBody CreateCategoryResponse
                if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
                    t.Fatalf("could not decode json response: %v", err)
                }
                assert.Equal(t, tt.mockCategory.ID, responseBody.ID)
                assert.Equal(t, tt.mockCategory.Name, responseBody.Name)
                assert.Equal(t, tt.mockCategory.Code, responseBody.Code)
                return
            }

            if tt.expectedMessage != "" {
                body, _ := io.ReadAll(resp.Body)
                assert.Contains(t, string(body), tt.expectedMessage)
            }
        })
    }
}
