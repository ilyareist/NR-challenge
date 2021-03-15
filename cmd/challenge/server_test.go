package main_test

import (
	"testing"
	"net"
	"bufio"
	"fmt"
	"log"
	"time"
	"os"
)

func TestUniqueNumbers(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:4000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	w := bufio.NewWriter(conn)

	fmt.Fprintf(w, "475948576\n")
	fmt.Fprintf(w, "654236485\n")
	fmt.Fprintf(w, "475948576\n")

	fmt.Fprintf(w, "terminate\n")

	w.Flush()

	time.Sleep(2 * time.Second)

	f, err := os.Open("../../numbers.log")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	lines := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines++
	}

	if lines == 2 {
		t.Log(`Success`)
	}
}
