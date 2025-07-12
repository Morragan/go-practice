package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"morragan.go-practice/http-shop/shop"
)

type ShopPresenter struct {
	shop *shop.Shop
}

func (sp *ShopPresenter) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := sp.shop.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	productsJSON, err := json.Marshal(products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(productsJSON)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (sp *ShopPresenter) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	product, err := sp.shop.Get(id)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	productJSON, err := json.Marshal(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(productJSON)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (sp *ShopPresenter) CreateProduct(w http.ResponseWriter, r *http.Request) {
	productName := r.URL.Query().Get("name")
	priceStr := r.URL.Query().Get("price")
	currencyCode := r.URL.Query().Get("currency")
	if productName == "" || priceStr == "" {
		http.Error(w, "Missing product name or price", http.StatusBadRequest)
		return
	}

	price, err := shop.CreatePrice(priceStr, currencyCode)
	if err != nil {
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}

	newProduct := shop.ProductCreateRequest{
		Name:  productName,
		Price: price,
	}

	createdProduct, err := sp.shop.Create(newProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	createdProductJSON, err := json.Marshal(createdProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(createdProductJSON)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (sp *ShopPresenter) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	updateRequest := shop.ProductUpdateRequest{
		ID: id,
	}

	productName := r.URL.Query().Get("name")
	priceStr := r.URL.Query().Get("price")
	currencyCode := r.URL.Query().Get("currency")
	if productName == "" && (priceStr == "" || currencyCode == "") {
		http.Error(w, "Missing product name or price", http.StatusBadRequest)
		return
	}

	if priceStr != "" && currencyCode != "" {
		price, err := shop.CreatePrice(priceStr, currencyCode)
		if err != nil {
			http.Error(w, "Invalid price format", http.StatusBadRequest)
			return
		}
		updateRequest.Price = price
	}

	if productName != "" {
		updateRequest.Name = productName
	}

	updatedProduct, err := sp.shop.Update(updateRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	updatedProductJSON, err := json.Marshal(updatedProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(updatedProductJSON)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (sp *ShopPresenter) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := sp.shop.Delete(id)
	if err != nil {
		if errors.Is(err, shop.ErrProductNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleProductsRequest(sp *ShopPresenter, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sp.ListProducts(w, r)
	case http.MethodPost:
		sp.CreateProduct(w, r)
	case http.MethodPut:
		sp.UpdateProduct(w, r)
	case http.MethodDelete:
		sp.DeleteProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	shopPresenter := ShopPresenter{shop: shop.NewShop()}

	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		handleProductsRequest(&shopPresenter, w, r)
	})
	http.HandleFunc("/product", shopPresenter.GetProduct)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
