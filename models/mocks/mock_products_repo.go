package mock

import(
	"github.com/mytheresa/go-hiring-challenge/models"
)


type MockProductsRepository struct {
	CalledLimit int
	CalledOffset int
	MockProductCount func() (int64, error)
	MockListProducts func(limit int, offset int) ([]models.Product, error)
}


func (m *MockProductsRepository) ListProducts(limit int, offset int) ([]models.Product, error) {
	m.CalledLimit = limit
	m.CalledOffset = offset
	return m.MockListProducts(limit, offset)
}

func (m *MockProductsRepository) CountProducts() (int64, error) {
	return m.MockProductCount() 
}