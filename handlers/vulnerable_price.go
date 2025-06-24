package handlers

import (
	"html/template"
	"net/http"
	"secure-webapp/models"
	"strconv"
)

func VulnerablePriceHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Vulnerable Price Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Vulnerable Price Manipulation Shop</h1>
        <p class="warning">Warning: Prices can be manipulated using browser inspector!</p>
        
        <div class="products">
            <h2>Products</h2>
            {{range $id, $product := .Products}}
            <div class="product">
                <h3>{{$product.Name}}</h3>
                <p>Price: ${{printf "%.2f" $product.Price}}</p>
                <form method="POST" action="/vulnerable-price/add-to-cart">
                    <input type="hidden" name="product_id" value="{{$product.ID}}">
                    <input type="hidden" name="price" value="{{$product.Price}}" id="price_{{$product.ID}}">
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
                <form method="POST" action="/vulnerable-price/checkout">
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

	t, _ := template.New("vulnerable-price").Parse(tmpl)
	t.Execute(w, data)
}

func VulnerablePriceAddToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/vulnerable-price", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	productID := r.FormValue("product_id")
	quantity, _ := strconv.Atoi(r.FormValue("quantity"))

	// VULNERABILITY: Trust client-side price input
	clientPrice, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		clientPrice = 0.0 // Default to 0 if invalid
	}

	_, exists := models.GetProduct(productID)
	if !exists {
		http.Redirect(w, r, "/vulnerable-price", http.StatusSeeOther)
		return
	}

	cart := models.GetCart(sessionID)
	cart.Items = append(cart.Items, models.CartItem{
		ProductID: productID,
		Quantity:  quantity,
		Price:     clientPrice, // Using client-provided price!
	})

	// Recalculate total using manipulated prices
	cart.Total = 0
	for _, item := range cart.Items {
		cart.Total += item.Price * float64(item.Quantity)
	}

	models.SetCart(sessionID, cart)
	http.Redirect(w, r, "/vulnerable-price", http.StatusSeeOther)
}

func VulnerablePriceCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/vulnerable-price", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	if len(cart.Items) == 0 {
		http.Redirect(w, r, "/vulnerable-price", http.StatusSeeOther)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Checkout - Vulnerable Price Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Checkout Complete</h1>
        <p class="warning">Order processed with manipulated prices!</p>
        
        <h3>Order Summary:</h3>
        {{range .Cart.Items}}
        <div class="order-item">
            <p>Product ID: {{.ProductID}} - Quantity: {{.Quantity}} - Price: ${{printf "%.2f" .Price}}</p>
        </div>
        {{end}}
        
        <p><strong>Total Paid: ${{printf "%.2f" .Cart.Total}}</strong></p>
        <p><small>You successfully manipulated the prices!</small></p>
        
        <a href="/vulnerable-price">Back to Shop</a>
        <a href="/">Home</a>
    </div>
</body>
</html>`

	// Clear cart after checkout
	models.ClearCart(sessionID)

	data := struct {
		Cart models.Cart
	}{
		Cart: cart,
	}

	t, _ := template.New("vulnerable-checkout").Parse(tmpl)
	t.Execute(w, data)
}
