package categories

import (
	"encoding/json"
	"net/http"

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	var response ListCategoriesResponse
	for _, c := range categories {
		response.Categories = append(response.Categories, Category{
			Name: c.Name,
			Code: c.Code,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}


func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req Category
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Code == "" {
        http.Error(w, "name and code are required", http.StatusBadRequest)
        return
    }
	category, err := h.repo.CreateCategory(req.Code, req.Name)
    if err != nil {
        switch err {
        case models.ErrCategoryCodeExists, models.ErrCategoryNameExists:
            http.Error(w, err.Error(), http.StatusConflict)
        default:
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    res := CreateCategoryResponse{
        ID:   category.ID,
        Code: category.Code,
        Name: category.Name,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
