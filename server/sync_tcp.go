package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/NamanG22/Redis/config"
	"github.com/NamanG22/Redis/core"
)

func respondError(err error, conn net.Conn) {
	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(conn net.Conn, cmd *core.RedisCmd) {
	log.Println("Command received2:", cmd)
	err := core.EvalAndRespond(cmd, conn)
	if err != nil {
		respondError(err, conn)
	}
}

func readCommand(conn net.Conn) (*core.RedisCmd, error) {
	var buf []byte = make([]byte, 512)
	n, err := conn.Read(buf[:]) // system call to read from the network socket, blocks until data is available
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Command: strings.ToUpper(tokens[0]),
		Args:    tokens[1:],
	}, nil
	
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
			respond(conn, cmd);
		}
	}

}
