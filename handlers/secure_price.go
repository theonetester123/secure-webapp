package handlers

import (
	"html/template"
	"net/http"
	"secure-webapp/models"
	"strconv"
)

func SecurePriceHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Secure Price Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Secure Price Manipulation Shop</h1>
        <p class="success">This shop validates all prices server-side!</p>
        
        <div class="products">
            <h2>Products</h2>
            {{range $id, $product := .Products}}
            <div class="product">
                <h3>{{$product.Name}}</h3>
                <p>Price: ${{printf "%.2f" $product.Price}}</p>
                <form method="POST" action="/secure-price/add-to-cart">
                    <input type="hidden" name="product_id" value="{{$product.ID}}">
                    <!-- NO PRICE FIELD - Server will look up the price -->
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
                <form method="POST" action="/secure-price/checkout">
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

	t, _ := template.New("secure-price").Parse(tmpl)
	t.Execute(w, data)
}

func SecurePriceAddToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/secure-price", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	productID := r.FormValue("product_id")
	quantity, err := strconv.Atoi(r.FormValue("quantity"))

	// Validate quantity
	if err != nil || quantity < 1 || quantity > 10 {
		http.Redirect(w, r, "/secure-price", http.StatusSeeOther)
		return
	}

	// SECURITY: Only use server-side price lookup - no client price input at all
	product, exists := models.GetProduct(productID)
	if !exists {
		http.Redirect(w, r, "/secure-price", http.StatusSeeOther)
		return
	}

	cart := models.GetCart(sessionID)
	cart.Items = append(cart.Items, models.CartItem{
		ProductID: productID,
		Quantity:  quantity,
		Price:     product.Price, // Always use server-side price lookup
	})

	// Recalculate total using server-side prices only
	cart.Total = 0
	for _, item := range cart.Items {
		serverProduct, exists := models.GetProduct(item.ProductID)
		if exists {
			// Update the item price to ensure consistency
			cart.Total += serverProduct.Price * float64(item.Quantity)
		}
	}

	models.SetCart(sessionID, cart)
	http.Redirect(w, r, "/secure-price", http.StatusSeeOther)
}

func SecurePriceCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/secure-price", http.StatusSeeOther)
		return
	}

	sessionID := getOrCreateSession(w, r)
	cart := models.GetCart(sessionID)

	if len(cart.Items) == 0 {
		http.Redirect(w, r, "/secure-price", http.StatusSeeOther)
		return
	}

	// Double-check all prices server-side before processing
	serverTotal := 0.0
	validatedItems := []models.CartItem{}

	for _, item := range cart.Items {
		product, exists := models.GetProduct(item.ProductID)
		if exists {
			validatedItem := models.CartItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price, // Always enforce server price
			}
			validatedItems = append(validatedItems, validatedItem)
			serverTotal += product.Price * float64(item.Quantity)
		}
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Checkout - Secure Price Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Checkout Complete</h1>
        <p class="success">Order processed with server-validated prices!</p>
        
        <h3>Order Summary:</h3>
        {{range .Items}}
        <div class="order-item">
            <p>Product ID: {{.ProductID}} - Quantity: {{.Quantity}} - Price: ${{printf "%.2f" .Price}}</p>
        </div>
        {{end}}
        
        <p><strong>Total Paid: ${{printf "%.2f" .Total}}</strong></p>
        <p><small>All prices were looked up server-side - no client input trusted!</small></p>
        
        <a href="/secure-price">Back to Shop</a>
        <a href="/">Home</a>
    </div>
</body>
</html>`

	// Clear cart after checkout
	models.ClearCart(sessionID)

	data := struct {
		Items []models.CartItem
		Total float64
	}{
		Items: validatedItems,
		Total: serverTotal,
	}

	t, _ := template.New("secure-checkout").Parse(tmpl)
	t.Execute(w, data)
}
