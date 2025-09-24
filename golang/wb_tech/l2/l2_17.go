package l2

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

func Example_L2_17() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout (default 10s)")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--timeout duration] host port\n", os.Args[0])
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Printf("Connected to %s\n", address)

	done := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
		}
		close(done)
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			_, err := fmt.Fprintln(conn, scanner.Text())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
				break
			}
		}
		conn.Close()
	}()

	<-done
	fmt.Println("Connection closed.")
}
