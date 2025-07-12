package shop

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/bojanz/currency"
	"github.com/google/uuid"
)

type Product struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Price     currency.Amount `json:"price"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

var ErrProductNotFound = errors.New("product not found")

type ProductCreateRequest struct {
	Name  string
	Price currency.Amount
}

type ProductUpdateRequest struct {
	ID    string
	Name  string
	Price currency.Amount
}

type Shop struct {
	mu       sync.RWMutex
	Products []Product
}

type ShopRepository interface {
	List() ([]Product, error)
	Get(id string) (Product, error)
	Create(pcr ProductCreateRequest) (Product, error)
	Update(pur ProductUpdateRequest) (Product, error)
	Delete(id string) error
}

var _ ShopRepository = (*Shop)(nil)

func CreatePrice(value string, currencyCode string) (currency.Amount, error) {
	amount, err := currency.NewAmount(value, currencyCode)
	if err != nil {
		return currency.Amount{}, err
	}
	return amount, nil
}

func NewShop() *Shop {
	// Initialize a new shop with some products for development purposes; ignore errors because we know the values are valid.
	combPrice, _ := currency.NewAmount("3", "USD")
	toothbrushPrice, _ := currency.NewAmount("2", "USD")
	shampooPrice, _ := currency.NewAmount("5", "USD")

	return &Shop{
		Products: []Product{
			{ID: uuid.NewString(), Name: "Comb", Price: combPrice, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{ID: uuid.NewString(), Name: "Toothbrush", Price: toothbrushPrice, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{ID: uuid.NewString(), Name: "Shampoo", Price: shampooPrice, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		},
	}
}

func (s *Shop) List() ([]Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	productsCopy := make([]Product, len(s.Products))
	copy(productsCopy, s.Products)
	return productsCopy, nil
}

func (s *Shop) Get(id string) (Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, product := range s.Products {
		if product.ID == id {
			return product, nil
		}
	}
	return Product{}, ErrProductNotFound
}

func (s *Shop) Create(pcr ProductCreateRequest) (Product, error) {
	if pcr.Name == "" || pcr.Price.IsZero() || pcr.Price.IsNegative() {
		return Product{}, errors.New("invalid product data")
	}

	newProduct := Product{
		ID:        uuid.NewString(),
		Name:      pcr.Name,
		Price:     pcr.Price,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.Products = append(s.Products, newProduct)
	return newProduct, nil
}

func (s *Shop) Update(pur ProductUpdateRequest) (Product, error) {
	if pur.Name == "" && !pur.Price.IsPositive() {
		return Product{}, errors.New("invalid product data")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.Products {
		if s.Products[i].ID == pur.ID {
			if pur.Name != "" {
				s.Products[i].Name = pur.Name
			}
			if pur.Price.IsPositive() {
				s.Products[i].Price = pur.Price
			}
			s.Products[i].UpdatedAt = time.Now().UTC()
			return s.Products[i], nil
		}
	}

	return Product{}, errors.New("product not found")
}

func (s *Shop) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	indexToDelete := slices.IndexFunc(s.Products, func(p Product) bool {
		return p.ID == id
	})
	if indexToDelete == -1 {
		return errors.New("product not found")
	}
	s.Products = append(s.Products[:indexToDelete], s.Products[indexToDelete+1:]...)
	return nil
}
