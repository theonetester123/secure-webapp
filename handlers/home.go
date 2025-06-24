package handlers

import (
	"html/template"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Security Demo Shop</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Security Demo Shopping Platform</h1>
        <p>Choose a shop to explore different security scenarios:</p>
        
		<div class="shop-category">
				<h2>Price Manipulation</h2>
				<div class="shop-pair">
					<a href="/vulnerable-price" class="shop-btn vulnerable">
						<h3>Vulnerable Version</h3>
						<p>Client-side price manipulation</p>
					</a>
					
					<a href="/secure-price" class="shop-btn secure">
						<h3>Secure Version</h3>
						<p>Server-side price validation</p>
					</a>
				</div>
			</div>
			
			<div class="shop-category">
				<h2>Order Processing</h2>
				<div class="shop-pair">
					<a href="/vulnerable-order" class="shop-btn vulnerable">
						<h3>Vulnerable Version</h3>
						<p>Order manipulation vulnerabilities</p>
					</a>
					
					<a href="/secure-order" class="shop-btn secure">
						<h3>Secure Version</h3>
						<p>Proper validation & authorization</p>
					</a>
				</div>
			</div>
			</div>
		</body>
</html>`

	t, _ := template.New("home").Parse(tmpl)
	t.Execute(w, nil)
}
