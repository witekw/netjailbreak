package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func startApiServer(out *chan Frame, getConnection func(key string) net.Conn, removeConnection func(key string)) {
	http.HandleFunc("/", handleApiRequest(out, getConnection, removeConnection))
	log.Fatal(http.ListenAndServe(os.Getenv(ApiUrl), nil))
}

func handleApiRequest(out *chan Frame, getConnection func(key string) net.Conn, removeConnection func(key string)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			returnData(w, out)
		case http.MethodPost:
			acceptData(r, getConnection, removeConnection)
		default:
			log.Fatal("Method not supported")
		}
	}
}

func acceptData(r *http.Request, getConnection func(key string) net.Conn, removeConnection func(key string)) {
	key := r.Header.Get("Connection-Correlation")
	conn := getConnection(key)
	if conn != nil {

		if _, err := io.Copy(conn, r.Body); err != nil {
			log.Print("Error moving accepted data into connection ", err.Error())
		}

		if r.Header.Get("Connection-Closed") == "true" {
			log.Print("Connection closed in external part.")
		}
	}
}

func returnData(w http.ResponseWriter, out *chan Frame) {
	select {
	case frame := <-*out:
		w.Header().Add("Connection-Correlation", frame.connection)
		w.Header().Add("Connection-Disconnect", strconv.FormatBool(frame.disconnect))
		w.Write(frame.buffer) //:frame.size
	case <-time.After(250 * time.Millisecond):
		w.Header().Add("Connection-Correlation", "none") //TODO change to HTTP statuses
		w.Header().Add("Connection-Disconnect", "false")
	}
}
