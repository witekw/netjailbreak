package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	ApiUrl     = "API_URL"
	RemoteHost = "REMOTE_HOST"
	RemotePort = "REMOTE_PORT"
)

func main() {

	_, p1 := os.LookupEnv(RemoteHost)
	_, p2 := os.LookupEnv(RemotePort)
	_, p3 := os.LookupEnv(ApiUrl)

	if !p1 || !p2 || !p3 {
		log.Fatalf(
			"Environment variables %s, %s and %s should be set before running this program.",
			RemoteHost,
			RemotePort,
			ApiUrl,
		)
	} else {
		log.Printf("Starting polling %s and forwarding connections to %s:%s", os.Getenv(ApiUrl), os.Getenv(RemoteHost), os.Getenv(RemotePort))
	}

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

	go pollApiAndSendDataToRemote(getConnection, removeConnection, addConnection)

	<-done
}

func pollApiAndSendDataToRemote(getConnection func(key string) net.Conn, removeConnection func(key string), addConnection func(key string, conn net.Conn)) {
	client := &http.Client{}
	for true { //TODO pass main signal loop
		get, err := client.Get(os.Getenv(ApiUrl))
		if err != nil {
			log.Print("Error getting data to send to remote ", err.Error())
			time.Sleep(5 * time.Second)
		} else if "none" == get.Header.Get("Connection-Correlation") {
			// no data returned
		} else if "true" == get.Header.Get("Connection-Disconnect") {
			removeConnection(get.Header.Get("Connection-Correlation"))
		} else {
			conn := createOrGetConnection(get.Header.Get("Connection-Correlation"), getConnection, removeConnection, addConnection)
			written, err := io.Copy(conn, get.Body)
			if err != nil {
				log.Print("Error sending ...", err.Error())
			} else {
				log.Printf("Sent %v bytes", written)
			}
		}
	}
}

func createOrGetConnection(key string, getConnection func(key string) net.Conn, removeConnection func(key string), addConnection func(key string, conn net.Conn)) net.Conn {
	if getConnection(key) == nil {
		conn, err := net.Dial("tcp", os.Getenv(RemoteHost)+":"+os.Getenv(RemotePort))
		if err != nil {
			log.Fatal("Error connecting to remote ", err.Error())
		}
		addConnection(key, conn)
		go receiveFromRemoteAndPostToApi(key, getConnection, removeConnection)
	}
	return getConnection(key)
}

func receiveFromRemoteAndPostToApi(key string, getConnection func(key string) net.Conn, removeConnection func(key string)) {

	client := &http.Client{}

	conn := getConnection(key)
	defer conn.Close()
	defer removeConnection(key)
	isClosed := false

	for !isClosed {

		buf := make([]byte, 1<<14)
		read, err2 := conn.Read(buf)
		if err2 != nil { //&& err2 != io.EOF
			log.Print("Connection is closed.", err2.Error())
			isClosed = true
		}

		if read > 0 || isClosed { //TODO necessary check this is?

			log.Printf("Received %v bytes", read)

			req, err0 := http.NewRequest("POST", os.Getenv(ApiUrl), bytes.NewReader(buf[:read]))
			if err0 != nil {
				log.Print("Error creating request.", err0.Error())
			}

			req.Header.Add("Connection-Correlation", key)
			req.Header.Add("Connection-Closed", strconv.FormatBool(isClosed))
			_, err1 := client.Do(req)

			if err1 != nil {
				log.Print("Error submitting received data ", err1.Error())
			}
		}
	}
}
