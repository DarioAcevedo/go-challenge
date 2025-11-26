package models

import (
	"gorm.io/gorm"
)

type ProductsRepository interface {
	ListProducts(limit int, offset int) ([]Product, error)
}

type productsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) ProductsRepository {
	return &productsRepository{
		db: db,
	}
}

func (r *productsRepository) ListProducts(limit, offset int) ([]Product, error) {
	var products []Product
	err := r.db.Preload("Variants").Preload("Category").Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}
