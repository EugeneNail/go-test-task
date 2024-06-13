package workers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Worker struct {
	plainSymbols  string
	requestsCount int // ничего не защитил мьютексом, хотя должен обращаться к этим параметрам из разных рутин
	prices        map[string]string
}

type price struct {
	Symbol string `json:"symbol"`
	Value  string `json:"price"`
}

func New(symbols []string) *Worker {
	var plainSymbols string
	for _, symbol := range symbols {
		plainSymbols += fmt.Sprintf(`"%s",`, symbol)
	}

	return &Worker{
		"[" + plainSymbols[:len(plainSymbols)-1] + "]",
		0,
		map[string]string{},
	}
}

func (worker *Worker) Run(stopChan <-chan string, messageChan chan<- string) {
	for {
	selectLoop:
		select {
		case <-stopChan:
			return
		default:
			worker.requestsCount++
			var newPrices []price

			response, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbols=" + worker.plainSymbols) // дефолтный клиент под катом не имеет таймаута, можно зависнуть навсегда
			if err != nil {
				fmt.Println(fmt.Errorf("workers.Run: can't fetch data for %s: %w\n", worker.plainSymbols, err))
				response.Body.Close()
				break selectLoop // почему бы не делать просто continue
			}

			if response.StatusCode != http.StatusOK {
				var data map[string]any
				json.NewDecoder(response.Body).Decode(&data) // не обработана ошибка
				fmt.Println(fmt.Errorf("workers.Run: status %d %+v for symbols %s\n", response.StatusCode, data, worker.plainSymbols))
				response.Body.Close()
				break selectLoop
			}

			if err := json.NewDecoder(response.Body).Decode(&newPrices); err != nil {
				fmt.Println(fmt.Errorf("workers.Run: %w\n", err))
				response.Body.Close() // лучше через дефер, где-ннибудь да забудешь написать
				break selectLoop
			}

			for _, newPrice := range newPrices {
				message := fmt.Sprintf("%s price:%s", newPrice.Symbol, newPrice.Value)

				oldPrice, exists := worker.prices[newPrice.Symbol]
				if !exists || oldPrice != newPrice.Value {
					worker.prices[newPrice.Symbol] = newPrice.Value
					message += " changed"
				}

				messageChan <- message
			}
		}
	}
}

func (worker *Worker) GetRequestsCount() int {
	return worker.requestsCount
}
