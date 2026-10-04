package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Image       string `json:"image"`
}
type Item struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}
type Order struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Method  string `json:"method"`
	Address string `json:"address"`
	Notes   string `json:"notes"`
	Items   []Item `json:"items"`
	Kind    string `json:"kind"`
	Flavor  string `json:"flavor"`
	Size    string `json:"size"`
	Date    string `json:"date"`
}

func openDB(path string) (*sql.DB, error) {
	db, e := sql.Open("sqlite", path)
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(1)
	_, e = db.Exec(`PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000;
 CREATE TABLE IF NOT EXISTS products(id INTEGER PRIMARY KEY,name TEXT NOT NULL,description TEXT NOT NULL,price INTEGER NOT NULL,image TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS orders(id TEXT PRIMARY KEY,created_at TEXT NOT NULL,kind TEXT NOT NULL,customer TEXT NOT NULL,phone TEXT NOT NULL,method TEXT NOT NULL,address TEXT NOT NULL,notes TEXT NOT NULL,total INTEGER NOT NULL,details TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS order_items(order_id TEXT NOT NULL REFERENCES orders(id),product_id INTEGER NOT NULL REFERENCES products(id),quantity INTEGER NOT NULL,unit_price INTEGER NOT NULL);
 INSERT OR IGNORE INTO products VALUES(1,'Fatia sabor Kinder','Fatia de bolo de chocolate com recheio cremoso sabor Kinder.',1500,'kinder'),(2,'Fatia Ninho','Fatia de bolo com creme de leite Ninho e morangos.',1500,'ninho'),(3,'Bombom gourmet','Caixa de bombons gourmet para presentear ou saborear.',3500,'bombom'),(4,'Morango cravejado','Morango cravejado com cobertura cremosa.',1500,'morango');
 CREATE TABLE IF NOT EXISTS migrations(id TEXT PRIMARY KEY);
 UPDATE products SET name='Fatia sabor Kinder',description='Fatia de bolo de chocolate com recheio cremoso sabor Kinder.',price=1500,image='kinder' WHERE id=1 AND NOT EXISTS(SELECT 1 FROM migrations WHERE id='real-catalog-v1');
 UPDATE products SET name='Fatia Ninho',description='Fatia de bolo com creme de leite Ninho e morangos.',price=1500,image='ninho' WHERE id=2 AND NOT EXISTS(SELECT 1 FROM migrations WHERE id='real-catalog-v1');
 UPDATE products SET name='Bombom gourmet',description='Caixa de bombons gourmet para presentear ou saborear.',price=3500,image='bombom' WHERE id=3 AND NOT EXISTS(SELECT 1 FROM migrations WHERE id='real-catalog-v1');
 UPDATE products SET name='Morango cravejado',description='Morango cravejado com cobertura cremosa.',price=1500,image='morango' WHERE id=4 AND NOT EXISTS(SELECT 1 FROM migrations WHERE id='real-catalog-v1');
 INSERT OR IGNORE INTO migrations VALUES('real-catalog-v1');`)
	if e != nil {
		db.Close()
		return nil, e
	}
	return db, nil
}
func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, msg string) {
	reply(w, status, map[string]string{"error": msg})
}
func validate(o Order) error {
	if len(strings.TrimSpace(o.Name)) < 2 || len(o.Name) > 120 {
		return errors.New("Informe seu nome completo.")
	}
	digits := 0
	for _, r := range o.Phone {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	if digits < 10 || digits > 13 || len(o.Phone) > 30 {
		return errors.New("Informe um telefone válido com DDD.")
	}
	if o.Method != "pickup" && o.Method != "delivery" {
		return errors.New("Escolha entrega ou retirada.")
	}
	if o.Method == "delivery" && len(strings.TrimSpace(o.Address)) < 10 {
		return errors.New("Informe o endereço completo.")
	}
	if len(o.Address) > 500 || len(o.Notes) > 1000 {
		return errors.New("Texto acima do limite permitido.")
	}
	if o.Kind == "custom" {
		d, e := time.Parse("2006-01-02", o.Date)
		today := time.Now().Format("2006-01-02")
		if e != nil || d.Format("2006-01-02") < today {
			return errors.New("Escolha uma data futura.")
		}
		if o.Flavor != "Chocolate" && o.Flavor != "Morango" && o.Flavor != "Baunilha" {
			return errors.New("Escolha um sabor válido.")
		}
		if o.Size != "10 pessoas" && o.Size != "20 pessoas" && o.Size != "30 pessoas" {
			return errors.New("Escolha um tamanho válido.")
		}
		if len(o.Items) > 0 {
			return errors.New("Solicitação de bolo não deve conter produtos.")
		}
		return nil
	}
	if o.Kind != "ready" || len(o.Items) == 0 || len(o.Items) > 30 {
		return errors.New("Adicione produtos à sacola.")
	}
	seen := map[int]bool{}
	for _, i := range o.Items {
		if i.Quantity < 1 || i.Quantity > 50 || seen[i.ProductID] {
			return errors.New("Quantidade de produtos inválida.")
		}
		seen[i.ProductID] = true
	}
	return nil
}
func handler(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if db.PingContext(r.Context()) != nil {
			fail(w, 503, "Banco indisponível.")
			return
		}
		reply(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/products", func(w http.ResponseWriter, r *http.Request) {
		rows, e := db.QueryContext(r.Context(), "SELECT id,name,description,price,image FROM products ORDER BY id")
		if e != nil {
			fail(w, 500, "Não foi possível carregar o cardápio.")
			return
		}
		defer rows.Close()
		products := []Product{}
		for rows.Next() {
			var p Product
			if rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image) != nil {
				fail(w, 500, "Erro no catálogo.")
				return
			}
			products = append(products, p)
		}
		if rows.Err() != nil {
			fail(w, 500, "Erro no catálogo.")
			return
		}
		reply(w, 200, products)
	})
	mux.HandleFunc("POST /api/orders", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 16384)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var o Order
		if dec.Decode(&o) != nil {
			fail(w, 400, "Pedido inválido.")
			return
		}
		var extra any
		if dec.Decode(&extra) != io.EOF {
			fail(w, 400, "Pedido inválido.")
			return
		}
		if e := validate(o); e != nil {
			fail(w, 400, e.Error())
			return
		}
		tx, e := db.BeginTx(r.Context(), nil)
		if e != nil {
			fail(w, 500, "Não foi possível salvar o pedido.")
			return
		}
		defer tx.Rollback()
		total := 0
		prices := []int{}
		lines := []string{"*PEDIDO — CHEF OTHON*"}
		for _, i := range o.Items {
			var price int
			var name string
			if tx.QueryRowContext(r.Context(), "SELECT price,name FROM products WHERE id=?", i.ProductID).Scan(&price, &name) != nil {
				fail(w, 400, "Produto indisponível.")
				return
			}
			prices = append(prices, price)
			total += price * i.Quantity
			lines = append(lines, fmt.Sprintf("• %d × %s — %s cada | %s", i.Quantity, name, brl(price), brl(price*i.Quantity)))
		}
		bytes := make([]byte, 8)
		if _, e = rand.Read(bytes); e != nil {
			fail(w, 500, "Não foi possível criar o pedido.")
			return
		}
		id := strings.ToUpper(hex.EncodeToString(bytes))
		details, _ := json.Marshal(map[string]string{"flavor": o.Flavor, "size": o.Size, "date": o.Date})
		_, e = tx.ExecContext(r.Context(), "INSERT INTO orders VALUES(?,?,?,?,?,?,?,?,?,?)", id, time.Now().UTC().Format(time.RFC3339), o.Kind, strings.TrimSpace(o.Name), o.Phone, o.Method, o.Address, o.Notes, total, string(details))
		if e != nil {
			fail(w, 500, "Não foi possível salvar o pedido.")
			return
		}
		for n, i := range o.Items {
			if _, e = tx.ExecContext(r.Context(), "INSERT INTO order_items VALUES(?,?,?,?)", id, i.ProductID, i.Quantity, prices[n]); e != nil {
				fail(w, 500, "Não foi possível salvar os itens.")
				return
			}
		}
		if tx.Commit() != nil {
			fail(w, 500, "Não foi possível confirmar o pedido.")
			return
		}
		if o.Kind == "custom" {
			lines = []string{"*ORÇAMENTO DE BOLO — CHEF OTHON*", "Sabor: " + o.Flavor, "Tamanho: " + o.Size, "Data: " + o.Date, "Valor: a confirmar"}
		} else {
			lines = append(lines, "", "*Subtotal: "+brl(total)+"*", "Entrega e pagamento: a combinar")
		}
		lines = append(lines, "", "*Pedido: "+id+"*", "Cliente: "+strings.TrimSpace(o.Name), "Telefone: "+o.Phone)
		if o.Method == "delivery" {
			lines = append(lines, "Recebimento: Entrega", "Endereço: "+o.Address)
		} else {
			lines = append(lines, "Recebimento: Retirada")
		}
		if strings.TrimSpace(o.Notes) != "" {
			lines = append(lines, "Observações: "+o.Notes)
		}
		whatsappURL := "https://wa.me/5517991540161?text=" + url.QueryEscape(strings.Join(lines, "\n"))
		reply(w, 201, map[string]any{"id": id, "total": total, "status": "received", "kind": o.Kind, "whatsappUrl": whatsappURL})
	})
	return mux
}
func brl(cents int) string { return fmt.Sprintf("R$ %d,%02d", cents/100, cents%100) }
func main() {
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "chef-othon.db"
	}
	db, e := openDB(path)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	server := http.Server{Addr: addr, Handler: handler(db), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("Chef Othon API: http://%s", addr)
	log.Fatal(server.ListenAndServe())
}
