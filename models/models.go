package models

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Product struct {
	ID    string
	Name  string
	Price float64
}

type CartItem struct {
	ProductID string
	Quantity  int
	Price     float64 // This will be manipulated in vulnerable version
}

type Order struct {
	ID        string
	UserID    string
	Items     []CartItem
	Total     float64
	Status    string
	Timestamp time.Time
}

type Cart struct {
	Items []CartItem
	Total float64
}

// Global stores with mutex for concurrent access
var (
	Products      = make(map[string]Product)
	Orders        = make(map[string]Order)
	Carts         = make(map[string]Cart)   // session_id -> cart
	Sessions      = make(map[string]string) // session_id -> user_id
	ProductsMutex = sync.RWMutex{}
	OrdersMutex   = sync.RWMutex{}
	CartsMutex    = sync.RWMutex{}
	SessionsMutex = sync.RWMutex{}
)

func InitStores() {
	ProductsMutex.Lock()
	defer ProductsMutex.Unlock()

	Products["1"] = Product{ID: "1", Name: "Laptop", Price: 999.99}
	Products["2"] = Product{ID: "2", Name: "Mouse", Price: 29.99}
	Products["3"] = Product{ID: "3", Name: "Keyboard", Price: 79.99}
	Products["4"] = Product{ID: "4", Name: "Monitor", Price: 299.99}
}

func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GetProduct(id string) (Product, bool) {
	ProductsMutex.RLock()
	defer ProductsMutex.RUnlock()
	product, exists := Products[id]
	return product, exists
}

func GetOrder(id string) (Order, bool) {
	OrdersMutex.RLock()
	defer OrdersMutex.RUnlock()
	order, exists := Orders[id]
	return order, exists
}

func SetOrder(order Order) {
	OrdersMutex.Lock()
	defer OrdersMutex.Unlock()
	Orders[order.ID] = order
}

func GetCart(sessionID string) Cart {
	CartsMutex.RLock()
	defer CartsMutex.RUnlock()
	cart, exists := Carts[sessionID]
	if !exists {
		return Cart{Items: []CartItem{}, Total: 0}
	}
	return cart
}

func SetCart(sessionID string, cart Cart) {
	CartsMutex.Lock()
	defer CartsMutex.Unlock()
	Carts[sessionID] = cart
}

func ClearCart(sessionID string) {
	CartsMutex.Lock()
	defer CartsMutex.Unlock()
	delete(Carts, sessionID)
}

func GetSession(sessionID string) (string, bool) {
	SessionsMutex.RLock()
	defer SessionsMutex.RUnlock()
	userID, exists := Sessions[sessionID]
	return userID, exists
}

func SetSession(sessionID, userID string) {
	SessionsMutex.Lock()
	defer SessionsMutex.Unlock()
	Sessions[sessionID] = userID
}
