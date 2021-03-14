package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/ilyareist/NR-challenge/internal/config"
)

type numConn struct {
	mu sync.Mutex
	i  int
}

type numMap struct {
	mu     sync.Mutex
	numMap map[string]int
}

var data = make(chan string)
var terminate = make(chan struct{})

func main() {
	listener, err := net.Listen("tcp", config.ListenAddr())
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	connections := numConn{
		mu: sync.Mutex{},
		i:  0,
	}
	numbers := numMap{
		mu:     sync.Mutex{},
		numMap: make(map[string]int),
	}

	go updateStats(&numbers)
	go writeToLogFile(config.LogName())

	go func() {
		select {
		case <-terminate:
			listener.Close()
		}
	}()

	patternNumbers := regexp.MustCompile(`^[0-9]{9,9}$`)

	for {
		if isTerminated() {
			break
		}
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(&numbers, patternNumbers, conn, &connections)
	}
}

// Function to check was app terminated or not
func isTerminated() bool {
	select {
	case <-terminate:
		return true
	default:
		return false
	}
}

// Function to write uniq numbers from `data` channel to the file
func writeToLogFile(fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	for d := range data {
		_, err = fmt.Fprintln(f, d)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
		}
	}

	f.Close()
}

// updateStats prints a report to the STDOUT every config.ReportInt() period
func updateStats(uniqNumbers *numMap) {
	prevUniqTotal := 0
	prevDupTotal := 0
	for {
		curDupTotal := 0

		uniqNumbers.mu.Lock()
		curUniqTotal := len(uniqNumbers.numMap)
		for _, c := range uniqNumbers.numMap {
			if c > 1 {
				curDupTotal++
			}
		}
		uniqNumbers.mu.Unlock()

		curUniq := curUniqTotal - prevUniqTotal
		curDup := curDupTotal - prevDupTotal

		fmt.Printf("Received %d unique numbers, %d duplicates.  Unique total: %d\n", curUniq, curDup, curUniqTotal)

		prevUniqTotal = curUniqTotal
		prevDupTotal = curDupTotal

		time.Sleep(time.Duration(config.ReportInt()) * time.Second)
	}
}

// Function processInputLine checks valid numbers
// also responsible for "terminate" input line.
func processInputLine(uniqNumbers *numMap, pattern *regexp.Regexp, digits string, c net.Conn) {
	if pattern.MatchString(digits) {
		uniqNumbers.Inc(digits)
		if uniqNumbers.Value(digits) == 1 {
			data <- digits
		}
	} else {
		if digits == "terminate" {
			close(terminate)
			fmt.Println("Terminating...")
		}

		c.Close()
	}
}

// Handling connections
func handleConn(uniqNumbers *numMap, digitsPattern *regexp.Regexp, c net.Conn, numConn *numConn) {
	if numConn.Value() >= config.ClientsLimit() {
		c.Write([]byte("Too many concurrent connections" + "\n"))
		c.Close()
		return
	} else {
		numConn.Inc()
		if !isTerminated() {
			input := bufio.NewScanner(c)
			for input.Scan() {
				go processInputLine(uniqNumbers, digitsPattern, input.Text(), c)
				if isTerminated() {
					c.Close()
					numConn.Dec()
					return
				}
			}
		}
		numConn.Dec()
		c.Close()
	}
}

// Helpers to be concurrency safe
func (c *numConn) Inc() {
	c.mu.Lock()
	c.i++
	c.mu.Unlock()
}

func (c *numConn) Dec() {
	c.mu.Lock()
	c.i--
	c.mu.Unlock()
}

func (c *numConn) Value() (x int) {
	c.mu.Lock()
	x = c.i
	c.mu.Unlock()
	return
}

func (c *numMap) Inc(number string) {
	c.mu.Lock()
	c.numMap[number]++
	c.mu.Unlock()
}

func (c *numMap) Value(number string) int {
	c.mu.Lock()
	x := c.numMap[number]
	c.mu.Unlock()
	return x
}