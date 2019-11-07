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
func WrappedReader(reader *bufio.Reader) <-chan []string {
	output := make(chan []string)
	go func() {
		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("EOF")
					break
				}
			}
			line := strings.ToLower(strings.TrimSpace(string(data)))
			output <- strings.Split(line, ":")
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

type HandlerFn func(DbService, ...string) string

var HandlerMap map[string]HandlerFn = map[string]HandlerFn{}

func CreateDb(svc DbService, args ...string) string {
	err := svc.CreateDb()
	if err != nil {
		return fmt.Sprint("Failed to create db: ", err)
	}
	return "successfully created db\n"
}

func DropDb(svc DbService, args ...string) string {
	err := svc.DropDb()
	if err != nil {
		return fmt.Sprint("failed to drop db: ", err)
	}
	return "successfully dropped db\n"

}

func CreateTable(svc DbService, args ...string) string {
	err := svc.CreateTable()
	if err != nil {
		return fmt.Sprint("Create table failed: ", err)
	} else {
		return "successfully created table\n"
	}
}

func DropTable(svc DbService, args ...string) string {
	err := svc.DropTable()
	if err != nil {
		return fmt.Sprint("Drop table failed: ", err)
	} else {
		return "successfully dropped table\n"
	}
}

func Help(_ DbService, args ...string) string {
	help := "Available Commands: \n"
	for k, _ := range HandlerMap {
		help += fmt.Sprintf("\t%s\n", k)
	}
	return help
}

func InsertText(svc DbService, args ...string) string {
	if len(args) != 1 {
		return "One and only one argument allowed\n"
	}
	err := svc.InsertText(args[0])
	if err != nil {
		return fmt.Sprint("Insert failed: ", err)
	}
	return fmt.Sprintf("successfully inserted %s\n", args[0])
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
			fmt.Println("Read ", incoming[0])

			if incoming[0] == "exit" {
				return
			}
			handler, ok := HandlerMap[incoming[0]]
			if !ok {
				WriteAndFlush(writer, fmt.Sprintf("%s\n", incoming))
			} else {
				svc, ok := ctx.Value("svc").(DbService)
				if ok {
					res := handler(svc, incoming[1:]...)
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
	HandlerMap["help"] = Help
	HandlerMap["insert"] = InsertText
}
