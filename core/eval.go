package core

import (
	"time"
	"errors"
	"io"
	"strconv"
)

var RESP_NIL []byte = []byte("$-1\r\n")

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

func evalSET(args []string, conn io.ReadWriter) error {
	if len(args) < 2 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}
	key := args[0]
	value := args[1]
	var expirationMs int64 = -1
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex", "Ex", "eX":
			i++
			if i >= len(args) {
				return errors.New("ERR wrong number of arguments for 'set' command")
			}
			expirationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return errors.New("ERR invalid expiration time '" + args[i] + "'")
			}
			expirationMs = expirationSec * 1000
		default:
			return errors.New("ERR unknown option '" + args[i] + "'")
		}
	}

	Put(key, NewObj(value, expirationMs))
	conn.Write(Encode("OK", true))
	return nil
}

func evalGET(args []string, conn io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'get' command")
	}
	key := args[0]
	obj := Get(key)
	if obj == nil {
		conn.Write(RESP_NIL)
		return nil
	}
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		conn.Write(RESP_NIL)
		return nil
	}
	conn.Write(Encode(obj.Value, false))
	return nil
}

func evalTTL(args []string, conn io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'ttl' command")
	}
	key := args[0]
	obj := Get(key)
	if obj == nil {
		conn.Write(Encode(int64(-2), true))	
		return nil
	}
	if obj.ExpiresAt == -1 {
		conn.Write(Encode(int64(-1), true))
		return nil
	}
	ttl := obj.ExpiresAt - time.Now().UnixMilli()
	if ttl < 0 {
		conn.Write(Encode(int64(-2), true))	
		return nil
	}
	conn.Write(Encode(int64(ttl/1000), true))
	return nil
}

func EvalAndRespond(cmd *RedisCmd, conn io.ReadWriter) error {
	switch cmd.Command {
	case "PING":
		return evalPING(cmd.Args, conn)
	case "SET":
		return evalSET(cmd.Args, conn)
	case "GET":
		return evalGET(cmd.Args, conn)
	case "TTL":
		return evalTTL(cmd.Args, conn)
	default:
		return errors.New("ERR unknown command '" + cmd.Command + "'")
	}
}