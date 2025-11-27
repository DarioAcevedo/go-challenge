package models

import (
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductsRepository interface {
	ListProducts(*ProductFilters) ([]Product, error)
	CountProducts(*ProductFilters) (productCount int64, err error)
	GetProductByCode(string) (*Product, error)
}

type productsRepository struct {
	db *gorm.DB
}

type ProductFilters struct {
	Limit  int
	Offset int
	ProductCategory *string
	PriceLt *decimal.Decimal
}

func NewProductsRepository(db *gorm.DB) ProductsRepository {
	return &productsRepository{
		db: db,
	}
}

func (r *productsRepository) ListProducts(filters *ProductFilters) ([]Product, error) {
	var products []Product

	query := r.db.Model(&Product{}).Preload("Category").Preload("Variants")
	if filters.ProductCategory != nil {
		query = query.Joins("Category").Where(`"Category"."name" = ?`, filters.ProductCategory)
	}
	if filters.PriceLt != nil {
		query = query.Where("price < ?", filters.PriceLt)
	}
	err := query.Limit(filters.Limit).Offset(filters.Offset).Find(&products).Error
	return products, err
}

func (r *productsRepository) CountProducts(filters *ProductFilters) (productCount int64, err error) {
	query := r.db.Model(&Product{})
	if filters.ProductCategory != nil {
		query = query.Joins("Category").Where(`"Category"."name" = ?`, filters.ProductCategory)
	}
	if filters.PriceLt != nil {
		query = query.Where("price < ?", filters.PriceLt)
	}
	err = query.Count(&productCount).Error
	return 
}

func (r *productsRepository) GetProductByCode(code string) (*Product, error) {
	p := &Product{}

	query := r.db.Model(&Product{}).Preload("Variants").Preload("Category").Where("code = ?", code)
	err := query.First(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}