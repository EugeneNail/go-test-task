package main

import (
	"bufio"
	"fmt"
	"github.com/EugeneNail/go-test-task/internal/config"
	"github.com/EugeneNail/go-test-task/internal/workers"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	config, err := config.Load()
	if err != nil {
		panic(err)
	}

	workers := getWorkers(config)
	stopChan := make(chan string)
	messageChan := make(chan string)
	var wg sync.WaitGroup

	runWorkers(workers, &wg, stopChan, messageChan)
	go runLogging(messageChan)
	go runRequestCounter(workers)

	waitForStop(&wg, stopChan)
}

func getWorkers(config config.Config) []*workers.Worker {
	activeWorkers := make([]*workers.Worker, config.MaxWorkers)

	for i, group := range splitSymbols(config) {
		activeWorkers[i] = workers.New(group)
	}

	return activeWorkers
}

// не расширяемо - ты кидаешь сюда прям конфигу, зачем функции знать о конфиге - она должна разбивать сиволы на N групп (я бы явно передал сюда эти параметры) - если тебе придется бить еще какие-то символы или будешь копипастить эту фукнцию чтобы передать жддругую конфигу или будешь рафакторить эту функцию
func splitSymbols(config config.Config) [][]string {
	groups := make([][]string, config.MaxWorkers)
	for i, symbol := range config.Symbols {
		groupOffset := i % config.MaxWorkers // зщапаникует если MaxWorkers=0
		groups[groupOffset] = append(groups[groupOffset], symbol)
	}

	return groups
}

func runWorkers(workers []*workers.Worker, wg *sync.WaitGroup, stopChan chan string, messageChan chan string) {
	for _, worker := range workers {
		wg.Add(1)

		go func() {
			defer wg.Done()
			worker.Run(stopChan, messageChan)
		}()
	}
}

func runLogging(messageChan <-chan string) {
	for {
		fmt.Println(<-messageChan) // лучше через range - если канал закроется, не будешь бесконечно сыпать пустые сообщения
	}
}

func runRequestCounter(workers []*workers.Worker) {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			requestsTotal := 0
			for _, worker := range workers {
				requestsTotal += worker.GetRequestsCount()
			}
			fmt.Printf("worker requests total: %d\n", requestsTotal)
		}
	}
}

func waitForStop(wg *sync.WaitGroup, stopChan chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() && strings.TrimSpace(scanner.Text()) == "STOP" {
			close(stopChan)
			wg.Wait()
			os.Exit(1)
		}
	}
}
