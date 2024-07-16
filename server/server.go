package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const ENDPOINT_COTACAO = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

var DB *gorm.DB

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

type Cotacao struct {
	USDBRL struct {
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

type CotacaoTable struct {
	Code       string `gorm:"primaryKey"`
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)

	var err error
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Printf("Falha ao fazer a conexão com o banco de dados: %v\n", err)
		panic(err)
	}
	DB.AutoMigrate(&CotacaoTable{})

	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cotacao, err := buscarCotacao(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = gravarCotacao(ctx, cotacao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CotacaoResponse{Bid: cotacao.USDBRL.Bid})
}

func buscarCotacao(ctx context.Context) (*Cotacao, error) {
	// O timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ENDPOINT_COTACAO, nil)
	if err != nil {
		log.Printf("Falha ao formatar requisição: %v\n", err)
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
		log.Printf("Falha ao ler o body da resposta: %v\n", err)
		return nil, err
	}

	var c Cotacao
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Printf("Falha ao codificar a resposta: %v\n", err)
		return nil, err
	}

	return &c, nil
}

func gravarCotacao(ctx context.Context, cotacao *Cotacao) error {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	err := DB.WithContext(ctx).Save(&CotacaoTable{
		Code:       cotacao.USDBRL.Code,
		Codein:     cotacao.USDBRL.Codein,
		Name:       cotacao.USDBRL.Name,
		High:       cotacao.USDBRL.High,
		Low:        cotacao.USDBRL.Low,
		VarBid:     cotacao.USDBRL.VarBid,
		PctChange:  cotacao.USDBRL.PctChange,
		Bid:        cotacao.USDBRL.Bid,
		Ask:        cotacao.USDBRL.Ask,
		Timestamp:  cotacao.USDBRL.Timestamp,
		CreateDate: cotacao.USDBRL.CreateDate,
	}).Error
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Retornar erro nos logs caso o tempo de execução seja insuficiente.
			log.Println("Falha ao inserir as informações de cotação pois a operação ultrapassou os 10ms permitidos.")
		}
		return err
	}

	return nil
}
