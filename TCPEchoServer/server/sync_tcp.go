package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/NamanG22/Redis/TCPEchoServer/config"
)

func respond(conn net.Conn, cmd string) error {
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
}

func readCommand(conn net.Conn) (string, error) {
	var buf []byte = make([]byte, 512)
	n, err := conn.Read(buf[:]) // system call to read from the network socket, blocks until data is available
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func RunSyncTCPServer() {
	log.Println("Starting synchronous TCP server on", config.Host, config.Port)

	var con_clients int = 0

	lsnr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port)) // instance of the server socket
	if err != nil {
		panic(err)
	}

	for {
		conn, err := lsnr.Accept() // blocking call until a new connection is established
		if err != nil {
			panic(err)
		}

		con_clients++
		log.Println("New client connected:", conn.RemoteAddr(), "concurrent clients:", con_clients)

		for {
			cmd, err := readCommand(conn)
			if err != nil {
				conn.Close()
				con_clients--
				log.Println("Client disconnected:", conn.RemoteAddr(), "concurrent clients:", con_clients)
				if err == io.EOF {
					break
				}
				log.Println("Error reading command:", err)
			}
			log.Println("Command received:", cmd)
			if err = respond(conn, cmd); err != nil {
				log.Println("Error responding to client:", err)
			}
		}
	}

}
