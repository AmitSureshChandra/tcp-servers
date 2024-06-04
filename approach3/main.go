package main

import (
	"io"
	"log"
	"syscall"
)

func main() {
	runServer()
}

var clients = 0

func runServer() {

	errChannel := make(chan error, 100)

	fd, err := setUpSocket()

	defer syscall.Close(fd)

	if err != nil {
		panic(err.Error())
	}

	for {
		go logErrors(errChannel)

		nfd, _, err := syscall.Accept(fd)

		clients++
		println("client connection created, total clients ", clients)

		if err != nil {
			panic(err.Error())
		}

		go func(nfd int) {
			err := handleNFD(nfd, &clients)
			if err != io.EOF {
				errChannel <- err
			}
		}(nfd)
	}
}

func logErrors(errChannel chan error) {
	// logging errors
	for err := range errChannel {
		log.Println(err.Error())
	}
}

func setUpSocket() (int, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
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

func handleNFD(nfd int, clients *int) error {

	defer syscall.Close(nfd)

	for {
		buffer := make([]byte, 2048)
		n, err := syscall.Read(nfd, buffer)
		if err != nil {
			return err
		}

		if n == 0 {
			*clients--
			println("client connection closed, total clients ", *clients)
			return io.EOF
		}

		_, err = syscall.Write(nfd, buffer[:n])
		if err != nil {
			return err
		}
	}
}
