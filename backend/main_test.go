package main

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestOrderPersistence(t *testing.T) {
	db, e := openDB(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	h := handler(db)
	body := `{"kind":"ready","name":"Cliente Teste","phone":"11999999999","method":"pickup","items":[{"productId":1,"quantity":2}]}`
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/orders", strings.NewReader(body)))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result struct {
		ID          string
		Total       int
		WhatsAppURL string `json:"whatsappUrl"`
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.Total != 3000 {
		t.Fatal(result)
	}
	link, err := url.Parse(result.WhatsAppURL)
	if err != nil {
		t.Fatal(err)
	}
	if link.Host != "wa.me" || link.Path != "/5517991540161" {
		t.Fatal("incorrect recipient", link)
	}
	message := link.Query().Get("text")
	for _, expected := range []string{"2 × Fatia sabor Kinder", "R$ 30,00", "Cliente: Cliente Teste", "Recebimento: Retirada", result.ID} {
		if !strings.Contains(message, expected) {
			t.Fatalf("message missing %s: %s", expected, message)
		}
	}
	var count int
	db.QueryRow("SELECT COUNT(*) FROM order_items WHERE order_id=?", result.ID).Scan(&count)
	if count != 1 {
		t.Fatal("items not persisted")
	}
}
func TestRejectInvalidOrders(t *testing.T) {
	db, _ := openDB(":memory:")
	defer db.Close()
	h := handler(db)
	for _, items := range []string{`[]`, `[{"productId":1,"quantity":-1}]`, `[{"productId":999,"quantity":1}]`, `[{"productId":1,"quantity":1},{"productId":1,"quantity":1}]`} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/api/orders", strings.NewReader(`{"kind":"ready","name":"Teste","phone":"11999999999","method":"pickup","items":`+items+`}`)))
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	var count int
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if count != 0 {
		t.Fatal("invalid orders saved")
	}
}
