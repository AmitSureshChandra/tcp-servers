package main

import (
	"log"
	"net"
	"runtime"
)

func main() {
	if err := runServer(); err != nil {
		panic(err.Error())
	}
}

var clients = 0

func runServer() error {

	listen, err := net.Listen("tcp", "127.0.0.1:8080")

	defer listen.Close()

	if err != nil {
		panic(err.Error())
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			panic(err.Error())
		}

		go func() {
			err := handleConn(conn)
			if err != nil {
				log.Println(err.Error())
			}
		}()
	}
}

func closeConn(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		log.Println(err.Error())
	}
}

func handleConn(conn net.Conn) error {

	log.Println("no of threads used ", runtime.NumGoroutine())

	defer closeConn(conn)

	log.Println("connected clients : ", clients)

	buffer := make([]byte, 2048)

	for {
		_, err := conn.Read(buffer)
		if err != nil {
			return err
		}

		log.Println("read : ", string(buffer))

		_, err = conn.Write(buffer)
		if err != nil {
			return err
		}
		log.Println("write : ", string(buffer))
	}
}
