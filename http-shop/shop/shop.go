package shop

import (
	"errors"
	"slices"
	"time"

	"github.com/bojanz/currency"
)

type Product struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Price     currency.Amount `json:"price"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type ProductCreateRequest struct {
	Name  string
	Price currency.Amount
}

type ProductUpdateRequest struct {
	ID    int
	Name  string
	Price currency.Amount
}

type Shop struct {
	Products []Product
}

type ShopRepository interface {
	List() ([]Product, error)
	Get(id int) (Product, error)
	Create(pcr ProductCreateRequest) (Product, error)
	Update(pur ProductUpdateRequest) (Product, error)
	Delete(id int) error
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
		[]Product{
			{ID: 1, Name: "Comb", Price: combPrice, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{ID: 2, Name: "Toothbrush", Price: toothbrushPrice, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{ID: 3, Name: "Shampoo", Price: shampooPrice, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		},
	}
}

func (s *Shop) List() ([]Product, error) {
	return s.Products, nil
}

func (s *Shop) Get(id int) (Product, error) {
	for _, product := range s.Products {
		if product.ID == id {
			return product, nil
		}
	}
	return Product{}, errors.New("product not found")
}

func (s *Shop) Create(pcr ProductCreateRequest) (Product, error) {
	if pcr.Name == "" || pcr.Price.IsZero() || pcr.Price.IsNegative() {
		return Product{}, errors.New("invalid product data")
	}

	newProduct := Product{
		// TODO: Use a less error-prone way to generate IDs
		// TODO: make the shop thread-safe
		ID:        len(s.Products) + 1,
		Name:      pcr.Name,
		Price:     pcr.Price,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.Products = append(s.Products, newProduct)
	return newProduct, nil
}

func (s *Shop) Update(pur ProductUpdateRequest) (Product, error) {
	if pur.Name == "" && !pur.Price.IsPositive() {
		return Product{}, errors.New("invalid product data")
	}

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

func (s *Shop) Delete(id int) error {
	indexToDelete := slices.IndexFunc(s.Products, func(p Product) bool {
		return p.ID == id
	})
	if indexToDelete == -1 {
		return errors.New("product not found")
	}
	s.Products = append(s.Products[:indexToDelete], s.Products[indexToDelete+1:]...)
	return nil
}
