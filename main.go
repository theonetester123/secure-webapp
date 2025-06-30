package main

import (
	"log"
	"net/http"
	"secure-webapp/handlers"
	"secure-webapp/models"
)

func main() {
	// Initialize data stores
	models.InitStores()

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Routes
	http.HandleFunc("/", handlers.HomeHandler)

	// Vulnerable Order Processing Shop
	http.HandleFunc("/vulnerable-order", handlers.VulnerableOrderHandler)
	http.HandleFunc("/vulnerable-order/add-to-cart", handlers.VulnerableAddToCartHandler)
	http.HandleFunc("/vulnerable-order/checkout", handlers.VulnerableCheckoutHandler)
	http.HandleFunc("/vulnerable-order/pay", handlers.VulnerablePayHandler)
	http.HandleFunc("/vulnerable-order/confirm", handlers.VulnerableConfirmHandler)
	http.HandleFunc("/vulnerable-order/result", handlers.VulnerableOrderResultHandler)

	// Secure Order Processing Shop
	http.HandleFunc("/secure-order", handlers.SecureOrderHandler)
	http.HandleFunc("/secure-order/add-to-cart", handlers.SecureAddToCartHandler)
	http.HandleFunc("/secure-order/checkout", handlers.SecureCheckoutHandler)
	http.HandleFunc("/secure-order/pay", handlers.SecurePayHandler)
	http.HandleFunc("/secure-order/result", handlers.SecureOrderResultHandler)

	// Vulnerable Price Manipulation Shop
	http.HandleFunc("/vulnerable-price", handlers.VulnerablePriceHandler)
	http.HandleFunc("/vulnerable-price/add-to-cart", handlers.VulnerablePriceAddToCartHandler)
	http.HandleFunc("/vulnerable-price/checkout", handlers.VulnerablePriceCheckoutHandler)

	// Secure Price Manipulation Shop
	http.HandleFunc("/secure-price", handlers.SecurePriceHandler)
	http.HandleFunc("/secure-price/add-to-cart", handlers.SecurePriceAddToCartHandler)
	http.HandleFunc("/secure-price/checkout", handlers.SecurePriceCheckoutHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
