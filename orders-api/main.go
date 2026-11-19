package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	db  *pgxpool.Pool
	rdb *redis.Client
	ctx = context.Background()
)

type OrderRequest struct {
	CustomerName string `json:"customer_name"`
	Items []struct {
		MenuItemID int64 `json:"menu_item_id"`
		Quantity   int   `json:"quantity"`
	} `json:"items"`
}

func jsonResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func health(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(r.Context()); err != nil {
		jsonResponse(w, 503, map[string]string{"status": "database unavailable"})
		return
	}
	if err := rdb.Ping(r.Context()).Err(); err != nil {
		jsonResponse(w, 503, map[string]string{"status": "redis unavailable"})
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "healthy"})
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var request OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if request.CustomerName == "" || len(request.Items) == 0 {
		jsonResponse(w, 400, map[string]string{"error": "customer_name and items are required"})
		return
	}

	tx, err := db.Begin(r.Context())
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": "could not start transaction"})
		return
	}
	defer tx.Rollback(r.Context())

	var orderID int64
	err = tx.QueryRow(r.Context(),
		`INSERT INTO orders (customer_name, status, total_amount)
		 VALUES ($1, 'NEW', 0) RETURNING id`,
		request.CustomerName).Scan(&orderID)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": "could not create order"})
		return
	}

	var total float64
	for _, item := range request.Items {
		if item.MenuItemID <= 0 || item.Quantity <= 0 {
			jsonResponse(w, 400, map[string]string{"error": "invalid item or quantity"})
			return
		}
		var price float64
		err = tx.QueryRow(r.Context(),
			`SELECT price FROM menu_items WHERE id=$1 AND available=TRUE`,
			item.MenuItemID).Scan(&price)
		if err != nil {
			jsonResponse(w, 400, map[string]string{"error": fmt.Sprintf("menu item %d unavailable", item.MenuItemID)})
			return
		}
		lineTotal := price * float64(item.Quantity)
		_, err = tx.Exec(r.Context(),
			`INSERT INTO order_items (order_id, menu_item_id, quantity, unit_price, line_total)
			 VALUES ($1,$2,$3,$4,$5)`,
			orderID, item.MenuItemID, item.Quantity, price, lineTotal)
		if err != nil {
			jsonResponse(w, 500, map[string]string{"error": "could not save order items"})
			return
		}
		total += lineTotal
	}

	_, err = tx.Exec(r.Context(),
		`UPDATE orders SET total_amount=$1 WHERE id=$2`, total, orderID)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": "could not update total"})
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		jsonResponse(w, 500, map[string]string{"error": "could not commit order"})
		return
	}

	if err := rdb.LPush(r.Context(), "kitchen:orders", orderID).Err(); err != nil {
		log.Printf("Redis enqueue failed for order %d: %v", orderID, err)
	}
	jsonResponse(w, 201, map[string]any{
		"id": orderID, "status": "NEW", "total_amount": total,
	})
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid order ID"})
		return
	}
	var order struct {
		ID int64 `json:"id"`
		Customer string `json:"customer_name"`
		Status string `json:"status"`
		Total float64 `json:"total_amount"`
	}
	err = db.QueryRow(r.Context(),
		`SELECT id, customer_name, status, total_amount FROM orders WHERE id=$1`, id).
		Scan(&order.ID, &order.Customer, &order.Status, &order.Total)
	if err == pgx.ErrNoRows {
		jsonResponse(w, 404, map[string]string{"error": "order not found"})
		return
	}
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": "database error"})
		return
	}
	jsonResponse(w, 200, order)
}

func main() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), env("DB_PORT", "5432"), env("DB_NAME", "dhaba"))
	var err error
	for i := 0; i < 30; i++ {
		db, err = pgxpool.New(ctx, dsn)
		if err == nil {
			err = db.Ping(ctx)
		}
		if err == nil { break }
		if db != nil { db.Close() }
		log.Printf("Waiting for PostgreSQL: %v", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil { log.Fatal(err) }

	rdb = redis.NewClient(&redis.Options{
		Addr: env("REDIS_HOST", "redis") + ":" + env("REDIS_PORT", "6379"),
	})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /orders", createOrder)
	mux.HandleFunc("GET /orders/{id}", getOrder)
	log.Println("Orders API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" { return value }
	return fallback
}
