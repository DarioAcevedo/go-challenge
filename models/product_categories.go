package models

type ProductCategory struct {
	ID       uint     `gorm:"primaryKey"`
	Name     string   `gorm:"not null"`
	Code string `gorm:"not null"`
	Products []Product `gorm:"foreignKey;ProductCategoryID"` 
}

func (v *ProductCategory) TableName() string {
	return "product_categories"
}