package server

import (
	"fmt"
	"io"
	"net"
	"log"
	"strconv"
	"strings"
	"bytes"

	"github.com/NamanG22/Redis/config"
	"github.com/NamanG22/Redis/core"
)

func respondError(err error, conn io.ReadWriter) {
	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(conn io.ReadWriter, cmds core.RedisCmds) {
	var response []byte
	buf := bytes.NewBuffer(response)
	for _, cmd := range cmds {
		resp := core.EvalAndRespond(cmd)
		buf.Write(resp)
	}
	conn.Write(buf.Bytes())
}

func readCommands(conn io.ReadWriter) (core.RedisCmds, error) {
	var buf []byte = make([]byte, 512)
	n, err := conn.Read(buf[:]) // system call to read from the network socket, blocks until data is available
	if err != nil {
		return nil, err
	}

	values, err := core.Decode(buf[:n])
	var cmds core.RedisCmds = make([]*core.RedisCmd, 0)
	for _, value := range values {
		tokens, err := toArrayString(value.([]interface{}))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, &core.RedisCmd{
			Command: strings.ToUpper(tokens[0]),
			Args:    tokens[1:],
		})
	}

	return cmds, nil
}

func toArrayString(value []interface{}) ([]string, error) {
	// convert a single array of inputs like [SET,k,v] to individual strings
	as := make([]string, len(value))
	for i := range as {
		as[i] = value[i].(string)
	}
	return as, nil
}

func RunSyncTCPServer() {
	log.Println("Starting synchronous TCP server on", config.Host, config.Port)

	var con_clients int = 0

	lsnr, err := net.Listen("tcp", config.Host + ":" + strconv.Itoa(config.Port))// instance of the server socket
	if err != nil {
		log.Println("Error listening:", err)
		return
	}

	for {
		conn, err := lsnr.Accept() // blocking call until a new connection is established
		if err != nil {
			log.Println("Error accepting connection:", err)
		}

		con_clients++

		for {
			cmds, err := readCommands(conn)
			if err != nil {
				conn.Close()
				con_clients--
				if err == io.EOF {
					break
				}
			}
			respond(conn, cmds);
		}
	}

}
