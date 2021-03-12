package handlers

import (
	"context"
	"gomicroservice/data"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Products is a http.Handler
type Products struct {
	l *log.Logger
}

// NewProducts creates a products handler with the given logger
func NewProducts(l *log.Logger) *Products {
	return &Products{l}
}

// GetProducts returns the products from the data store
func (p *Products) GetProducts(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle GET Products")

	// fetch the products from the datastore
	lp := data.GetProducts()

	// serialize the list to JSON
	err := lp.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to marshal JSON", http.StatusInternalServerError)
	}
}

// AddProduct adds a product to the datastor
func (p *Products) AddProduct(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle POST Product")

	// instantiate a Product data
	prod := &data.Product{}

	// deserialize json
	err := prod.FromJSON(r.Body)
	if err != nil {
		http.Error(rw, "Unable to unmarshal JSON", http.StatusBadRequest)
	}

	// add to the datastore
	data.AddProduct(prod)
}

// UpdateProducts update a product in datastore based on id
func (p *Products) UpdateProducts(rw http.ResponseWriter, r *http.Request) {
	// extract vars from route variables
	vars := mux.Vars(r)

	// change the id into int and assign it
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "Fail to receive id", http.StatusBadRequest)
		return
	}

	p.l.Println("Handle PUT Product", id)

	// instantiate a Product data
	prod := &data.Product{}

	// deserialize json
	err = prod.FromJSON(r.Body)
	if err != nil {
		http.Error(rw, "Unable to unmarshal JSON", http.StatusBadRequest)
	}

	// update product in datastore
	err = data.UpdateProduct(id, prod)

	// handle wrong id
	if err == data.ErrProductNotFound {
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	// handle update failure
	if err != nil {
		http.Error(rw, "Fail to update product", http.StatusInternalServerError)
		return
	}
}

type KeyProduct struct {
}

func (p Products) MiddlewareValidateProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prod := &data.Product{}
		err := prod.FromJSON(r.Body)

		if err != nil {
			http.Error(rw, "Unable to unmarshal JSON", http.StatusBadRequest)
			return
		}

		// add the product to the context
		ctx := context.WithValue(r.Context(), KeyProduct{}, prod)
		req := r.WithContext(ctx)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, req)
	})
}
