package models

import (
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func (r *ProductsRepository) List() ([]Product, error) {
	var products []Product
	err := r.db.Preload("Variants").Find(&products).Error
	return products, err
}
