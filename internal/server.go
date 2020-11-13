package main

import (
	"io"
	"log"
	"net"
	"os"
)

func startServer(out *chan Frame, addConnection func(key string, conn net.Conn)) {

	l, err := net.Listen(ConnType, os.Getenv(ServerAddress)+":"+os.Getenv(ServerPort))
	if err != nil {
		log.Fatal("Error starting server ", err.Error())
	}
	defer l.Close()

	log.Printf("Listening on %s", os.Getenv(ServerAddress)+":"+os.Getenv(ServerPort))
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting connection ", err.Error())
		}
		addConnection(conn.RemoteAddr().String(), conn)
		go readDataAndSaveForExternalGet(out, conn)
	}
}

func readDataAndSaveForExternalGet(c *chan Frame, conn net.Conn) {

	for isEOF := false; !isEOF; {

		buf := make([]byte, BufferSize)
		read, err := conn.Read(buf)

		switch err {
		case nil:
		case io.EOF:
			log.Printf("[%s] Closed", conn.RemoteAddr())
			isEOF = true
		default:
			log.Print("Error reading ", err.Error())
			isEOF = true
		}

		if read > 0 || isEOF {

			frame := Frame{
				conn.RemoteAddr().String(),
				buf[:read],
				isEOF,
			}

			*c <- frame
		}
	}
}
