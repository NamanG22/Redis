package core

import (
	"time"
	"strconv"
)

var RESP_NIL []byte = []byte("$-1\r\n")

func evalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode("ERR wrong number of arguments for 'ping' command", false)
	}
	if len(args) == 0 {
		b = Encode("PONG",true);
	} else {
		b = Encode(args[0],false);
	}
	return b
}

func evalSET(args []string) []byte {
	if len(args) < 2 {
		return Encode("ERR wrong number of arguments for 'set' command", false)
	}
	key := args[0]
	value := args[1]
	var expirationMs int64 = -1
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex", "Ex", "eX":
			i++
			if i >= len(args) {
				return Encode("ERR wrong number of arguments for 'set' command", false)
			}
			expirationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode("ERR invalid expiration time '" + args[i] + "'", false)
			}
			expirationMs = expirationSec * 1000
		default:
			return Encode("ERR unknown option '" + args[i] + "'", false)
		}
	}

	Put(key, NewObj(value, expirationMs))
	return Encode("OK", true)
}

func evalGET(args []string) []byte {
	if len(args) != 1 {
		return Encode("ERR wrong number of arguments for 'get' command", false)
	}
	key := args[0]
	obj := Get(key)
	if obj == nil {
		return RESP_NIL
	}
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		return RESP_NIL
	}
	return Encode(obj.Value, false)
}

func evalTTL(args []string) []byte {
	if len(args) != 1 {
		return Encode("ERR wrong number of arguments for 'ttl' command", false)
	}
	key := args[0]
	obj := Get(key)
	if obj == nil {
		return Encode(int64(-2), true)
	}
	if obj.ExpiresAt == -1 {
		return Encode(int64(-1), true)
	}
	ttl := obj.ExpiresAt - time.Now().UnixMilli()
	if ttl < 0 {
		return Encode(int64(-2), true)
	}
	return Encode(int64(ttl/1000), true)
}

func evalDEL(args []string) []byte {
	if len(args) < 1 {
		return Encode("ERR wrong number of arguments for 'del' command", false)
	}
	deleted := 0
	for _, key := range args {
		if Delete(key) {
			deleted++
		}
	}
	return Encode(int64(deleted), true)
}

func evalEXPIRE(args []string) []byte {
	if len(args) <= 1 {
		return Encode("ERR wrong number of arguments for 'expire' command", false)
	}
	key := args[0]
	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode("ERR invalid expiration time '" + args[1] + "'", false)
	}
	obj := Get(key)

	if obj == nil {
		return Encode(int64(0), true)
	}

	obj.ExpiresAt = time.Now().UnixMilli() + duration * 1000
	return Encode(int64(1), true)
}

func EvalAndRespond(cmd *RedisCmd) []byte {
	switch cmd.Command {
	case "PING":
		return evalPING(cmd.Args)
	case "SET":
		return evalSET(cmd.Args)
	case "GET":
		return evalGET(cmd.Args)
	case "TTL":
		return evalTTL(cmd.Args)
	case "DEL":
		return evalDEL(cmd.Args)
	case "EXPIRE":
		return evalEXPIRE(cmd.Args)
	default:
		return Encode("ERR unknown command '" + cmd.Command + "'", false)
	}
}