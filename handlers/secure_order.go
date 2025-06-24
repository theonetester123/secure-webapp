package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"secure-webapp/models"
	"strconv"
	"time"
)

func SecureOrderHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Secure Order Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Secure Order Processing Shop</h1>
        <p class="success">This shop implements proper security measures!</p>
        
        <div class="products">
            <h2>Products</h2>
            {{range $id, $product := .Products}}
            <div class="product">
                <h3>{{$product.Name}}</h3>
                <p>Price: ${{printf "%.2f" $product.Price}}</p>
                <form method="POST" action="/secure-order/add-to-cart">
                    <input type="hidden" name="product_id" value="{{$product.ID}}">
                    <input type="number" name="quantity" value="1" min="1" max="10">
                    <button type="submit">Add to Cart</button>
                </form>
            </div>
            {{end}}
        </div>

        <div class="cart">
            <h2>Cart</h2>
            {{if .Cart.Items}}
                {{range .Cart.Items}}
                <div class="cart-item">
                    <p>Product ID: {{.ProductID}} - Quantity: {{.Quantity}} - Price: ${{printf "%.2f" .Price}}</p>
                </div>
                {{end}}
                <p><strong>Total: ${{printf "%.2f" .Cart.Total}}</strong></p>
                <form method="POST" action="/secure-order/checkout">
                    <button type="submit">Checkout</button>
                </form>
            {{else}}
                <p>Cart is empty</p>
            {{end}}
        </div>
        
        <a href="/">Back to Home</a>
    </div>
</body>
</html>`

	data := struct {
		Products map[string]models.Product
		Cart     models.Cart
	}{
		Products: models.Products,
		Cart:     cart,
	}

	t, _ := template.New("secure-order").Parse(tmpl)
	t.Execute(w, data)
}

func SecureAddToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/secure-order", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	productID := r.FormValue("product_id")
	quantity, err := strconv.Atoi(r.FormValue("quantity"))

	// Validate quantity
	if err != nil || quantity < 1 || quantity > 10 {
		http.Redirect(w, r, "/secure-order", http.StatusSeeOther)
		return
	}

	// SECURITY: Always look up product and price from server database
	product, exists := models.GetProduct(productID)
	if !exists {
		http.Redirect(w, r, "/secure-order", http.StatusSeeOther)
		return
	}

	cart := models.GetCart(sessionID)
	cart.Items = append(cart.Items, models.CartItem{
		ProductID: productID,
		Quantity:  quantity,
		Price:     product.Price, // Server-side price only
	})

	// Recalculate total server-side with fresh price lookups
	cart.Total = 0
	for i, item := range cart.Items {
		serverProduct, exists := models.GetProduct(item.ProductID)
		if exists {
			// Ensure cart item has correct server price
			cart.Items[i].Price = serverProduct.Price
			cart.Total += serverProduct.Price * float64(item.Quantity)
		}
	}

	models.SetCart(sessionID, cart)
	http.Redirect(w, r, "/secure-order", http.StatusSeeOther)
}
func SecureCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/secure-order", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	if len(cart.Items) == 0 {
		http.Redirect(w, r, "/secure-order", http.StatusSeeOther)
		return
	}

	// Validate cart total server-side
	serverTotal := 0.0
	for _, item := range cart.Items {
		product, exists := models.GetProduct(item.ProductID)
		if exists {
			serverTotal += product.Price * float64(item.Quantity)
		}
	}

	orderID := models.GenerateID()
	userID, _ := models.GetSession(sessionID)

	order := models.Order{
		ID:        orderID,
		UserID:    userID,
		Items:     cart.Items,
		Total:     serverTotal, // Use server-calculated total
		Status:    "pending",
		Timestamp: time.Now(),
	}

	models.SetOrder(order)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Payment - Secure Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Secure Payment Page</h1>
        <p>Order ID: {{.OrderID}}</p>
        <p>Total: ${{printf "%.2f" .Total}}</p>
        
        <form method="POST" action="/secure-order/confirm">
            <input type="hidden" name="order_id" value="{{.OrderID}}">
            <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
            <h3>Payment Details</h3>
            <input type="text" placeholder="Card Number" required>
            <input type="text" placeholder="CVV" required>
            <button type="submit">Pay Now</button>
        </form>
    </div>
</body>
</html>`

	data := struct {
		OrderID   string
		Total     float64
		CSRFToken string
	}{
		OrderID:   orderID,
		Total:     serverTotal,
		CSRFToken: models.GenerateID(), // Simple CSRF token
	}

	t, _ := template.New("secure-payment").Parse(tmpl)
	t.Execute(w, data)
}

func SecureConfirmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/secure-order", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	orderID := r.FormValue("order_id")

	// Validate order ownership
	order, exists := models.GetOrder(orderID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	userID, _ := models.GetSession(sessionID)
	if order.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Simulate payment verification
	if order.Status != "pending" {
		http.Error(w, "Order already processed", http.StatusBadRequest)
		return
	}

	// Update order status
	order.Status = "completed"
	models.SetOrder(order)

	// Clear cart
	models.ClearCart(sessionID)

	// Redirect to result
	http.Redirect(w, r, fmt.Sprintf("/secure-order/result?order_id=%s", orderID), http.StatusSeeOther)
}

func SecureOrderResultHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getOrCreateSession(w, r)
	orderID := r.URL.Query().Get("order_id")

	order, exists := models.GetOrder(orderID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Validate order ownership
	userID, _ := models.GetSession(sessionID)
	if order.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Order Result - Secure Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Order Complete</h1>
        <p>Order ID: {{.Order.ID}}</p>
        <p>Status: {{.Order.Status}}</p>
        <p>Total: ${{printf "%.2f" .Order.Total}}</p>
        <p>Date: {{.Order.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        
        <h3>Items:</h3>
        {{range .Order.Items}}
        <div class="order-item">
            <p>Product ID: {{.ProductID}} - Quantity: {{.Quantity}} - Price: ${{printf "%.2f" .Price}}</p>
        </div>
        {{end}}
        
        <a href="/secure-order">Back to Shop</a>
        <a href="/">Home</a>
    </div>
</body>
</html>`

	data := struct {
		Order models.Order
	}{
		Order: order,
	}

	t, _ := template.New("secure-result").Parse(tmpl)
	t.Execute(w, data)
}
