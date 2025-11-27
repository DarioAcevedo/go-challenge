package categories

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type ListCategoriesResponse struct {
	Categories []Category `json:"categories"`
}

type CreateCategoryResponse struct {
	ID       uint    `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
}

type Category struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
}

type CategoriesHandler struct {
	repo models.CategoriesRepository
}

func NewCategoriesHandler(r models.CategoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{
		repo: r,
	}
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.ListCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	var response ListCategoriesResponse
	for _, c := range categories {
		response.Categories = append(response.Categories, Category{
			Name: c.Name,
			Code: c.Code,
		})
	}
	api.OKResponse(w, response)
}


func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req Category
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "name and code are required")
        return
    }
	category, err := h.repo.CreateCategory(req.Code, req.Name)
    if err != nil {
        switch err {
        case models.ErrCategoryCodeExists, models.ErrCategoryNameExists:
			api.ErrorResponse(w, http.StatusConflict, err.Error())
        default:
			api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
        }
        return
    }

    res := CreateCategoryResponse{
        ID:   category.ID,
        Code: category.Code,
        Name: category.Name,
    }
	api.OKResponse(w, res)	
}
