package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

const ENDPOINT_COTACAO = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"-"`
		Codein     string `json:"-"`
		Name       string `json:"-"`
		High       string `json:"-"`
		Low        string `json:"-"`
		VarBid     string `json:"-"`
		PctChange  string `json:"-"`
		Bid        string `json:"bid"`
		Ask        string `json:"-"`
		Timestamp  string `json:"-"`
		CreateDate string `json:"-"`
	} `json:"USDBRL"`
}

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	cotacao, err := buscarCotacao()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CotacaoResponse{Bid: cotacao.USDBRL.Bid})
}

func buscarCotacao() (*Cotacao, error) {
	// O timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ENDPOINT_COTACAO, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Retornar erro nos logs caso o tempo de execução seja insuficiente.
			log.Println("Falha ao consultar a cotação pois a chamada ultrapassou os 200ms permitidos.")
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c Cotacao
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
