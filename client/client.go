package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const ENDPOINT_SERVER_COTACACAO = "http://localhost:8080/cotacao"

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	// Utilizando o package "context", o client.go terá um timeout máximo de 300ms para receber o resultado do server.go
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ENDPOINT_SERVER_COTACACAO, nil)
	if err != nil {
		log.Printf("Falha ao formatar requisição: %v\n", err)
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Retornar erro nos logs caso o tempo de execução seja insuficiente.
			log.Println("Falha ao consultar o server de cotação pois a chamada ultrapassou os 300ms permitidos.")
		}
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Falha ao ler o body da resposta: %v\n", err)
		panic(err)
	}

	var c Cotacao
	err = json.Unmarshal([]byte(body), &c)
	if err != nil {
		log.Printf("Falha ao codificar a resposta: %v\n", err)
		panic(err)
	}

	// O client.go terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Printf("Falha ao criar o arquivo: %v\n", err)
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", c.Bid))
	if err != nil {
		fmt.Printf("Falha ao gravar as informações de cotação no arquivo: %v\n", err)
		panic(err)
	}
}
