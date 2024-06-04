package main

import (
	"io"
	"syscall"
)

func main() {
	if err := runServer(); err != nil {
		panic(err.Error())
	}
}

var clients = 0

func runServer() error {

	fd, err := setUpSocket()

	if err != nil {
		return err
	}

	defer syscall.Close(fd)

	for {
		nfd, _, err := syscall.Accept(fd)

		clients++
		println("client connection created, total clients ", clients)

		if err != nil {
			return err
		}

		if err := handleNFD(nfd); err != nil {
			clients--
			println("client connection closed, total clients ", clients)
			if err == io.EOF {
				// connection closed by client => go for new con req
				continue
			}
			return err
		}
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

func handleNFD(nfd int) error {

	defer syscall.Close(nfd)

	for {
		buffer := make([]byte, 2048)
		n, err := syscall.Read(nfd, buffer)
		if err != nil {
			return err
		}

		if n == 0 {
			return io.EOF
		}

		_, err = syscall.Write(nfd, buffer[:n])
		if err != nil {
			return err
		}
	}
}
