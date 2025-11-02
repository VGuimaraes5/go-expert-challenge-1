package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	http.HandleFunc("/cotacao", cotacaoHandle)
	http.ListenAndServe(":8080", nil)
}

func cotacaoHandle(w http.ResponseWriter, r *http.Request) {
	quotation := getRealTimeQuotation()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quotation)
}

type QuotationResponse struct {
	UsdBrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Quotation struct {
	Bid string `json:"bid"`
}

func getRealTimeQuotation() Quotation {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var QuotationResponse QuotationResponse
	err = json.NewDecoder(res.Body).Decode(&QuotationResponse)
	if err != nil {
		panic(err)
	}

	persistQuotation(QuotationResponse)
	quotation := Quotation{Bid: QuotationResponse.UsdBrl.Bid}
	return quotation
}

func persistQuotation(quotation QuotationResponse) {
	// connect to database sqlite
	db, err := sql.Open("sqlite", "./server.db")
	if err != nil {
		fmt.Println("Erro ao abrir o banco de dados:", err)
		return
	}
	defer db.Close() // Fechar a conexão quando a função terminar

	// Exemplo: criar uma tabela se ela não existir
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS quotations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT,
			codein TEXT,
			name TEXT,
			bid TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)

	if err != nil {
		fmt.Println("Erro ao criar tabela:", err)
		return
	}

	// timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms.
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	insertQuery := `INSERT INTO quotations (code, codein, name, bid) VALUES (?, ?, ?, ?)`
	_, err = db.ExecContext(ctx, insertQuery, quotation.UsdBrl.Code, quotation.UsdBrl.Codein, quotation.UsdBrl.Name, quotation.UsdBrl.Bid)
	if err != nil {
		fmt.Println("Erro ao inserir cotação:", err)
		return
	}
}
