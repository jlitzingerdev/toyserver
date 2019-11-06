package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const CONN_EXPIRED = "Connection has expired\n"

// Helper to allow selecting to wait for data from input stream.  This
// likely has overhead that exceeds managing the reads directly, but
// provides a selectable interface
func WrappedReader(reader *bufio.Reader) <-chan string {
	output := make(chan string)
	go func() {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("EOF")
				return
			}
		}
		output <- strings.ToLower(strings.TrimSpace(string(data)))
	}()
	return output
}

// Function to write the specified string and flush, possibly logging
// errors along the way
func WriteAndFlush(w *bufio.Writer, data string) {
	bw, err := w.WriteString(data)
	if bw != len(data) {
		fmt.Println("Unable to write data ", err)
	}
	w.Flush()
}

// Handler for a single connection
func HandleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	recvCh := WrappedReader(bufio.NewReader(conn))

	for {
		select {
		case <-ctx.Done():
			WriteAndFlush(writer, CONN_EXPIRED)
			return
		case incoming := <-recvCh:
			fmt.Println("Read ", incoming)
			if incoming == "exit" {
				return
			} else {
				WriteAndFlush(writer, fmt.Sprintf("%s\n", incoming))
			}
		}
	}
}

func main() {
	fmt.Println("Starting server on 10000")
	listener, err := net.Listen("tcp", ":10000")
	if err != nil {
		fmt.Println("Unable to listen on 10000, aborting ", err)
		return
	}
	ctx := context.Background()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed accepting ", err)
		}

		go func() {
			deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			HandleConn(deadline, conn)
		}()
	}
}
