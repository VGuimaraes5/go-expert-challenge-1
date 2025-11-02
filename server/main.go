package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
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
	data, err := json.MarshalIndent(quotation, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Println("Salvando cotação no banco de dados")
	log.Println(string(data))
	// TODO: implementar a persistência no banco de dados
}
