package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	ChannelLength int = 1 << 12
	BufferSize    int = 1 << 15
	ApiUrl            = "API_URL"
	ServerAddress     = "SERVER_ADDRESS"
	ServerPort        = "SERVER_PORT"
	ConnType          = "tcp"
)

type Frame struct {
	connection string
	buffer     []byte
	disconnect bool
}

func main() {

	_, p1 := os.LookupEnv(ServerAddress)
	_, p2 := os.LookupEnv(ServerPort)
	_, p3 := os.LookupEnv(ApiUrl)

	if !p1 || !p2 || !p3 {
		log.Fatalf(
			"Environment variables: %s, %s and %s must be set before running this program.", ServerAddress, ServerPort, ApiUrl)
	}

	log.Printf("Api url is %s", os.Getenv(ApiUrl))

	//handle signals
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	//~handle signals

	out := make(chan Frame, ChannelLength)

	connections := make(map[string]net.Conn)

	getConnection := func(key string) net.Conn {
		return connections[key]
	}

	removeConnection := func(key string) {
		if connections[key] != nil {
			connections[key].Close()
			delete(connections, key)
			log.Printf("Removed connection %s, still %v active", key, len(connections))
		}
	}

	addConnection := func(key string, conn net.Conn) {
		connections[key] = conn
		log.Printf("Added connection %s", key)
	}

	go startApiServer(&out, getConnection, removeConnection)
	go startServer(&out, addConnection)

	<-done
	log.Print("Exiting")
}
