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
		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("EOF")
					break
				}
			}
			output <- strings.ToLower(strings.TrimSpace(string(data)))
		}
		close(output)
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

type HandlerFn func(svc DbService) string

var HandlerMap map[string]HandlerFn = map[string]HandlerFn{}

func CreateDb(svc DbService) string {
	svc.CreateDb()
	return "successfully created db\n"
}

func DropDb(svc DbService) string {

	svc.DropDb()
	return "successfully dropped db\n"

}

func CreateTable(svc DbService) string {
	err := svc.CreateTable()
	if err != nil {
		return fmt.Sprint("Create table failed: ", err)
	} else {
		return "successfully created table\n"
	}
}

func DropTable(svc DbService) string {
	err := svc.DropTable()
	if err != nil {
		return fmt.Sprint("Drop table failed: ", err)
	} else {
		return "successfully dropped table\n"
	}
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
			}
			handler, ok := HandlerMap[incoming]
			if !ok {
				WriteAndFlush(writer, fmt.Sprintf("%s\n", incoming))
			} else {
				svc, ok := ctx.Value("svc").(DbService)
				if ok {
					res := handler(svc)
					WriteAndFlush(writer, res)
				} else {
					WriteAndFlush(writer, "context is not a db interface\n")
				}
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

	svc, err := NewDbService()
	ctx := context.WithValue(context.Background(), "svc", svc)
	if err != nil {
		fmt.Println("Failed to start db service ", err)
		return
	}
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

func init() {
	HandlerMap["createdb"] = CreateDb
	HandlerMap["dropdb"] = DropDb
	HandlerMap["createtable"] = CreateTable
	HandlerMap["droptable"] = DropTable
}
