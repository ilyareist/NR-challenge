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

	// Starting updatingReport routine
	go updateReport(&numbers)
	// Starting writing to the log file routine
	go writeToLogFile(config.LogName())

	// Checking if app should be terminated
	go func() {
		<-terminate
		// Checking if all active connections are closed:
		for {
			if connections.Value() == 0 {
				listener.Close()
			}
		}
	}()

	// numbers pattern to check
	patternNumbers := regexp.MustCompile(`^[0-9]{9,9}$`)

	// handling new connections
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

// updateReport prints a report to the STDOUT every config.ReportInt() period
func updateReport(uniqNumbers *numMap) {
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

// Function checkNumbers checks valid numbers
// Also responsible for "terminate" input line.
func checkNumbers(uniqNumbers *numMap, pattern *regexp.Regexp, digits string, c net.Conn, numConn *numConn) {
	if isTerminated() {
		closeConnection(c)
		numConn.Dec()
	}
	if pattern.MatchString(digits) {
		uniqNumbers.Inc(digits)
		if uniqNumbers.Value(digits) == 1 {
			select {
			case <-data:
				closeConnection(c)
			default:
				data <- digits
			}
		}
	} else {
		if digits == "terminate" {
			fmt.Println("Terminating...")
			closeConnection(c)
			close(terminate)
		}
		closeConnection(c)
	}
}

// Handling connections
func handleConn(uniqNumbers *numMap, digitsPattern *regexp.Regexp, c net.Conn, numConn *numConn) {
	go func() {
		<-terminate
		c.Close()
	}()
	if numConn.Value() >= config.ClientsLimit() {
		_, err := c.Write([]byte("Too many concurrent connections" + "\n"))
		if err != nil {
			log.Println(err)
		}
		c.Close()
		return
	} else {
		numConn.Inc()
		if !isTerminated() {
			input := bufio.NewScanner(c)
			for input.Scan() {
				go checkNumbers(uniqNumbers, digitsPattern, input.Text(), c, numConn)
				if isTerminated() {
					closeConnection(c)
					numConn.Dec()
					return
				}
			}
		}
		numConn.Dec()
		closeConnection(c)
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

func closeConnection(c net.Conn) {
	if c != nil {
		c.Close()
		c = nil
	}
}
