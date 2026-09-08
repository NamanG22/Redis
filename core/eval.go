package core

import (
	"errors"
	"net"
	"log"
)

func evalPING(args []string, conn net.Conn) error {
	log.Println("Command received2:", args)
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}
	if len(args) == 0 {
		b = Encode("PONG",true);
	} else {
		b = Encode(args[0],false);
	}
	_, err := conn.Write(b);
	return err
}

func EvalAndRespond(cmd *RedisCmd, conn net.Conn) error {
	log.Println("Command received1:", cmd)
	switch cmd.Command {
	case "PING":
		return evalPING(cmd.Args, conn)
	default:
		return evalPING(cmd.Args, conn)
	}
}