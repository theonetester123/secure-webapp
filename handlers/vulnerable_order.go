package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"secure-webapp/models"
	"strconv"
	"time"
)

func VulnerableOrderHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Vulnerable Order Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Vulnerable Order Processing Shop</h1>
        <p class="warning">Warning: Orders can be completed without paying!</p>
        
        <div class="products">
            <h2>Products</h2>
            {{range $id, $product := .Products}}
            <div class="product">
                <h3>{{$product.Name}}</h3>
                <p>Price: ${{printf "%.2f" $product.Price}}</p>
                <form method="POST" action="/vulnerable-order/add-to-cart">
                    <input type="hidden" name="product_id" value="{{$product.ID}}">
                    <input type="number" name="quantity" value="1" min="1">
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
                <form method="POST" action="/vulnerable-order/checkout">
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

	t, _ := template.New("vulnerable-order").Parse(tmpl)
	t.Execute(w, data)
}

func VulnerableAddToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/vulnerable-order", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	productID := r.FormValue("product_id")
	quantity, _ := strconv.Atoi(r.FormValue("quantity"))

	product, exists := models.GetProduct(productID)
	if !exists {
		http.Redirect(w, r, "/vulnerable-order", http.StatusSeeOther)
		return
	}

	cart := models.GetCart(sessionID)
	cart.Items = append(cart.Items, models.CartItem{
		ProductID: productID,
		Quantity:  quantity,
		Price:     product.Price,
	})

	// Recalculate total
	cart.Total = 0
	for _, item := range cart.Items {
		cart.Total += item.Price * float64(item.Quantity)
	}

	models.SetCart(sessionID, cart)
	http.Redirect(w, r, "/vulnerable-order", http.StatusSeeOther)
}

func VulnerableCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/vulnerable-order", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	if len(cart.Items) == 0 {
		http.Redirect(w, r, "/vulnerable-order", http.StatusSeeOther)
		return
	}

	// Create order
	orderID := models.GenerateID()
	userID, _ := models.GetSession(sessionID)

	order := models.Order{
		ID:        orderID,
		UserID:    userID,
		Items:     cart.Items,
		Total:     cart.Total,
		Status:    "pending",
		Timestamp: time.Now(),
	}

	models.SetOrder(order)

	// Redirect to payment page
	http.Redirect(w, r, fmt.Sprintf("/vulnerable-order/pay?order_id=%s", orderID), http.StatusSeeOther)
}

// Payment page - shows form and handles POST
func VulnerablePayHandler(w http.ResponseWriter, r *http.Request) {
	var orderID string

	if r.Method == "POST" {
		// For POST requests, get order_id from form data
		orderID = r.FormValue("order_id")
	} else {
		// For GET requests, get order_id from URL query
		orderID = r.URL.Query().Get("order_id")
	}
	if orderID == "" {
		http.Redirect(w, r, "/vulnerable-order", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {

		// VULNERABILITY: No validation of payment details or order ownership
		// Just redirect to confirm page
		http.Redirect(w, r, fmt.Sprintf("/vulnerable-order/confirm?order_id=%s", orderID), http.StatusSeeOther)
		return
	}

	// VULNERABILITY: No validation of order ownership when showing payment form
	order, exists := models.GetOrder(orderID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Show payment form (GET request)
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Payment - Vulnerable Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Payment Page</h1>
        <p>Order ID: {{.OrderID}}</p>
        <p>Total: ${{printf "%.2f" .Total}}</p>
        
        <form method="POST" action="/vulnerable-order/pay">
            <input type="hidden" name="order_id" value="{{.OrderID}}">
            <h3>Payment Details</h3>
            <div>
                <label>Card Number:</label>
                <input type="text" name="card_number" placeholder="1234-5678-9012-3456" required>
            </div>
            <div>
                <label>CVV:</label>
                <input type="text" name="cvv" placeholder="123" required>
            </div>
            <button type="submit">Pay Now</button>
        </form>
    </div>
</body>
</html>`

	data := struct {
		OrderID string
		Total   float64
	}{
		OrderID: orderID,
		Total:   order.Total,
	}

	t, err := template.New("vulnerable-payment").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func VulnerableConfirmHandler(w http.ResponseWriter, r *http.Request) {
	var orderID string

	if r.Method == "POST" {
		orderID = r.FormValue("order_id")

		order, exists := models.GetOrder(orderID)
		if !exists {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		order.Status = "completed"
		models.SetOrder(order)

		sessionID := getOrCreateSession(w, r)
		models.ClearCart(sessionID)

		http.Redirect(w, r, fmt.Sprintf("/vulnerable-order/result?order_id=%s", orderID), http.StatusSeeOther)
		return
	}

	// GET request - show confirmation page
	orderID = r.URL.Query().Get("order_id")
	if orderID == "" {
		http.Redirect(w, r, "/vulnerable-order", http.StatusSeeOther)
		return
	}

	order, exists := models.GetOrder(orderID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Confirm Payment</title>
    <link rel="stylesheet" href="/static/style.css">
    <script>
        let countdown = 3;
        function updateCountdown() {
            document.getElementById('countdown').textContent = countdown;
            if (countdown <= 0) {
                document.getElementById('confirmForm').submit();
            } else {
                countdown--;
                setTimeout(updateCountdown, 1000);
            }
        }
        window.onload = function() {
            updateCountdown();
        };
    </script>
</head>
<body>
    <div class="container">
        <h1>Confirm Your Payment</h1>
        <p>Order ID: {{.OrderID}}</p>
        <p>Total: ${{printf "%.2f" .Total}}</p>
        <p>Payment processed. Order will be confirmed in <span id="countdown">3</span> seconds...</p>
        
        <form id="confirmForm" method="POST" action="/vulnerable-order/confirm">
            <input type="hidden" name="order_id" value="{{.OrderID}}">
            <button type="submit">Confirm Now</button>
        </form>
    </div>
</body>
</html>`

	data := struct {
		OrderID string
		Total   float64
	}{
		OrderID: orderID,
		Total:   order.Total,
	}

	t, _ := template.New("vulnerable-confirm").Parse(tmpl)
	t.Execute(w, data)
}

func VulnerableOrderResultHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")

	// VULNERABILITY: Anyone can view any order by guessing order_id
	order, exists := models.GetOrder(orderID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Order Result - Vulnerable Shop</title>
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
        
        <a href="/vulnerable-order">Back to Shop</a>
        <a href="/">Home</a>
    </div>
</body>
</html>`

	data := struct {
		Order models.Order
	}{
		Order: order,
	}

	t, _ := template.New("vulnerable-result").Parse(tmpl)
	t.Execute(w, data)
}
