package models

import(
	"errors"
	"gorm.io/gorm"
	"strings"
)

var ErrCategoryCodeExists = errors.New("category code already exists")
var ErrCategoryNameExists = errors.New("category name already exists")

type CategoriesRepository interface {
	ListCategories() ([]ProductCategory, error)
	CreateCategory(string, string) (*ProductCategory, error)
}


type categoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) CategoriesRepository {
	return &categoriesRepository{
		db: db,
	}
}

func(c *categoriesRepository) ListCategories() ([]ProductCategory, error) {
	var categories []ProductCategory
	err := c.db.Model(&ProductCategory{}).Find(&categories).Error
	return categories, err
}

func (r *categoriesRepository) CreateCategory(code string, name string) (*ProductCategory, error) {
    var existingByCode ProductCategory
    if err := r.db.Where("code = ?", code).First(&existingByCode).Error; err == nil {
        return nil, ErrCategoryCodeExists
    }

    // Check unique name: validate trimmed and lowercase to make sure there are no coincidences
    var existingByName ProductCategory
	nameClean := strings.ToLower(strings.TrimSpace(name))
    if err := r.db.Where("LOWER(name) = ?", nameClean).First(&existingByName).Error; err == nil {
        return nil, ErrCategoryNameExists
    }
	
	c := &ProductCategory{
        Code: code,
        Name: name,
    }

    if err := r.db.Create(c).Error; err != nil {
        return nil, err
    }

    return c, nil
}