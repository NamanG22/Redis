package core

import (
	"errors"
	"io"
)

func evalPING(args []string, conn io.ReadWriter) error {
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

func EvalAndRespond(cmd *RedisCmd, conn io.ReadWriter) error {
	switch cmd.Command {
	case "PING":
		return evalPING(cmd.Args, conn)
	default:
		return evalPING(cmd.Args, conn)
	}
}