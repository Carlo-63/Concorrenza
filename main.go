package main

// Questo programma simula il comportamento di un semaforo per la lettura di un file da parte di più goroutine (thread).
// Il semaforo viene impostato di base a 0, simulando un file occupato, e viene liberato dopo 2 secondi.
// Le goroutine competono per accedervi, svolgono il loro lavoro e poi liberano il semaforo.
// Il semaforo è implementato tramite un mutex, che garantisce la mutua esclusione tra le goroutine.

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// 1 = libero, 0 = occupato
// L'implementazione del semaforo può essere effettuata anche tramite variabili booleane (spin lock), ma in questo
// caso ho scelto di utilizzare un unsigned integer di 8 bit.
var (
	wg       sync.WaitGroup
	semaforo uint8                  = 0
	mutex    sync.Mutex             // Mutex per la mutua esclusione della risorsa semaforo
	cond     = sync.NewCond(&mutex) // Condizione per la sincronizzazione dell'accesso al semaforo
)

func readFile(fileName string, PID uint8) {
	defer wg.Done()

	mutex.Lock()
	fmt.Printf("\n[THREAD %v] Semaforo occupato\n", PID)

	for semaforo == 0 {
		fmt.Printf("[THREAD %v] Il file è già in uso, attendo...\n", PID)
		cond.Wait()
	}

	semaforo = 0
	mutex.Unlock()

	fmt.Printf("\n[THREAD %v] Lettura del file...\n", PID)

	// Alloca spazio per il contenuto
	content := make([]byte, 1024)

	// Apre il file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("\n[THREAD %v] Errore nell'apertura del file: %v", PID, err)

		goto fine
	}

	defer file.Close()

	// Legge il contenuto
	_, err = file.Read(content)
	if err != nil {
		fmt.Printf("[THREAD %v] Errore nella lettura del file: %v", PID, err)

		goto fine
	}

	fmt.Printf("\n[THREAD %v, CONTENUTO]\n", PID)
	fmt.Println(string(content))

fine:
	mutex.Lock()
	semaforo = 1
	cond.Broadcast()
	mutex.Unlock()

	fmt.Printf("\n[THREAD %v] Semaforo liberato\n", PID)
}

func main() {
	wg.Add(3)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Nome file: ")
	fileName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Errore nella lettura del nome del file:", err)
		return
	}

	fileName = strings.TrimSpace(fileName)

	go readFile(fileName, 0)
	go readFile(fileName, 1)

	go func() {
		defer wg.Done()

		time.Sleep(2 * time.Second)

		mutex.Lock()
		semaforo = 1
		cond.Broadcast()
		mutex.Unlock()

		fmt.Println("\n\nSemaforo liberato")
	}()

	wg.Wait()
}
