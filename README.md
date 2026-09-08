# Redis

A from-scratch Redis clone. This repo starts with a TCP server — the networking layer Redis sits on — then protocol parsing, commands, and storage. The long-term target is a Java Redis server; this first slice is in Go so the socket model is small and easy to inspect.

## Current status

The TCP path, RESP codec, and first command are wired together:

**Synchronous TCP server**

- Listens on `0.0.0.0:7379` by default (same port family as Redis, offset so it does not collide with a real Redis on `6379`)
- Accepts one connection at a time (the accept loop is blocked while a client is being served)
- Reads a RESP array from the client, turns it into a `RedisCmd`, evaluates it, and writes a RESP reply

**RESP codec** (`core.Decode` / `core.DecodeOne` / `core.Encode`)

- Parses simple strings (`+`), errors (`-`), integers (`:`), bulk strings (`$`), and arrays (`*`), including nested arrays
- `DecodeArrayString` flattens a RESP array into `[]string` for command tokens
- `Encode` writes simple strings (`+PONG`) and bulk strings (`$5\r\nhello\r\n`)

**Commands**

- `PING` → `+PONG`
- `PING <message>` → bulk-string echo of `<message>`
- Wrong arity for `PING` → RESP error

There is no key-value store, persistence, or further command set yet.

## Layout

```
main.go              # flags and process entry
config/config.go     # host and port
server/sync_tcp.go   # listen, accept, decode, respond
core/cmd.go          # RedisCmd (command + args)
core/eval.go         # command dispatch (PING)
core/resp.go         # RESP encode / decode
core/resp_test.go    # table-driven decode tests
go.mod
```

## Prerequisites

- Go 1.27+

## Run

From the repo root:

```bash
go run .
```

Or with explicit bind address:

```bash
go run . -host 127.0.0.1 -port 7379
```

You should see logs like:

```
rolling the dice
Starting synchronous TCP server on 0.0.0.0 7379
```

## Try it

The server speaks RESP, so a Redis client works:

```bash
redis-cli -p 7379 PING
```

Expected: `PONG`

```bash
redis-cli -p 7379 PING hello
```

Expected: `hello`

Or raw RESP over `nc`:

```bash
printf '*1\r\n$4\r\nPING\r\n' | nc 127.0.0.1 7379
```

Expected: `+PONG`

Disconnect with `Ctrl-D` (or close the client); the server then accepts the next connection.

Flags:

| Flag   | Default   | Meaning        |
|--------|-----------|----------------|
| `-host`  | `0.0.0.0` | bind address |
| `-port`  | `7379`    | bind port    |

## Test

```bash
go test ./core/
```

## What comes next

1. Implement a small command set (`GET`, `SET`, …) and reject unknown commands
2. Add an in-memory store
3. Rebuild the same server in Java
