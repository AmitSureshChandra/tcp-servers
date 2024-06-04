package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"net"
	"syscall"
)

func main() {
	if err := runServer(); err != nil {
		panic(err.Error())
	}
}

var clients = 0

func runServer() error {

	listen, err := net.Listen("tcp", "127.0.0.1:8089")
	if err != nil {
		panic(err.Error())
	}

	defer listen.Close()

	epollFD, err := CreateEpoll()

	if err != nil {
		return err
	}

	f, err := listen.(*net.TCPListener).File()

	if err != nil {
		return err
	}

	listenerFD := f.Fd()

	if err := syscall.SetNonblock(int(listenerFD), true); err != nil {
		println(err.Error())
	}

	if err := AddToPoll(epollFD, int(listenerFD)); err != nil {
		return err
	}

	fmt.Println("Server is listening on 127.0.0.1:8080")

	for {
		events, err := WailForEvents(epollFD)

		if err != nil {
			return err
		}

		for _, event := range events {

			if event.Fd == int32(listenerFD) {
				conn, err := listen.Accept()

				if err != nil {
					return err
				}

				f, err := conn.(*net.TCPConn).File()

				if err != nil {
					return err
				}
				connFD := int(f.Fd())

				if err := syscall.SetNonblock(connFD, true); err != nil {
					println(err.Error())
				}

				if err := AddToPoll(epollFD, connFD); err != nil {
					conn.Close()
					return err
				}

				clients++
				log.Println("client connected, clients ", clients)
			} else {
				err := handleConn(int(event.Fd))
				if err != nil {
					log.Println(err.Error())
					err = RemoveFromPoll(epollFD, int(event.Fd))
					if err != nil {
						log.Println(err.Error())
					}
				}
			}
		}
	}
}

func decreaseClient() {
	clients--
	log.Println("client disconnected, clients ", clients)
}

func handleConn(fd int) error {

	buffer := make([]byte, 2048)
	n, err := unix.Read(fd, buffer)

	if err != nil {
		decreaseClient()
		return err
	}

	if n == 0 {
		decreaseClient()
		return io.EOF
	}

	_, err = unix.Write(fd, buffer[:n])
	return err
}
