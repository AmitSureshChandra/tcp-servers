package main

import (
	"io"
	"log"
	"runtime"
	"syscall"
	"time"
)

func main() {
	runServer()
}

var clients = make([]int, 0)

func runServer() {

	fd, err := setUpSocket()

	if err != nil {
		panic(err.Error())
	}

	defer syscall.Close(fd)

	for {
		// check for new conn
		err := acceptConn(fd)
		if err != nil {
			log.Println(err.Error())
		}

		// check if read available
		handleNFDForReadWrite()

		// to check system is single threaded
		if runtime.NumGoroutine() > 1 {
			log.Println("runtime.NumGoroutine()", runtime.NumGoroutine())
		}

		time.Sleep(2 * time.Second)
	}
}

func handleNFDForReadWrite() {
	for i, client := range clients {
		err := handleNFD(client)
		if err != nil {
			// close the current nfd & remove from clients
			clients = append(clients[:i], clients[i+1:]...)
			println("client connection closed, total clients ", len(clients))
			log.Println(err.Error())
		}
	}
}

func acceptConn(fd int) error {
	nfd, _, err := syscall.Accept(fd)

	// skip if server_fd is not accepted any connection
	if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
		return nil
	}

	if err := syscall.SetNonblock(nfd, true); err != nil {
		return err
	}

	if err == nil {
		clients = append(clients, nfd)
		println("client connection created, total clients ", len(clients))
		return nil
	}

	return err
}

func setUpSocket() (int, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}

	if err := syscall.SetNonblock(fd, true); err != nil {
		return 0, err
	}

	err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: 8080,
		Addr: [4]byte{127, 0, 0, 1},
	})

	if err != nil {
		return 0, err
	}

	if err := syscall.Listen(fd, 100); err != nil {
		return 0, err
	}

	return fd, nil
}

func handleNFD(nfd int) error {
	buffer := make([]byte, 2048)
	n, err := syscall.Read(nfd, buffer)
	if err != nil {
		// skip if server_fd is not accepted any connection
		if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
			return nil
		}
		return err
	}

	if n == 0 {
		return io.EOF
	}

	if n > 0 {
		_, err = syscall.Write(nfd, buffer[:n])
		if err != nil {
			return err
		}
	}
	return nil
}
