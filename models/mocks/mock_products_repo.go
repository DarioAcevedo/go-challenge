package mock

import(
	"github.com/mytheresa/go-hiring-challenge/models"
)


type MockProductsRepository struct {
	CalledLimit int
	CalledOffset int
	CalledNumTimes int
	MockListProducts func(limit int, offset int) ([]models.Product, error)
}


func (m *MockProductsRepository) ListProducts(limit int, offset int) ([]models.Product, error) {
	m.CalledLimit = limit
	m.CalledOffset = offset
	m.CalledNumTimes++
	return m.MockListProducts(limit, offset)
}