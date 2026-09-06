# Redis

A from-scratch Redis clone. This repo starts with a TCP echo server — the networking layer Redis sits on — then protocol parsing, commands, and storage. The long-term target is a Java Redis server; this first slice is in Go so the socket model is small and easy to inspect.

## Current status

Two pieces are in place, not yet wired together:

**Synchronous TCP echo server**

- Listens on `0.0.0.0:7379` by default (same port family as Redis, offset so it does not collide with a real Redis on `6379`)
- Accepts one connection at a time (the accept loop is blocked while a client is being served)
- Reads up to 512 bytes from the client and writes the same bytes back
- Logs connect, disconnect, and each received payload

**RESP decoder** (`core.Decode` / `core.DecodeOne`)

- Parses simple strings (`+`), errors (`-`), integers (`:`), bulk strings (`$`), and arrays (`*`), including nested arrays
- Returns the Go value and, for `DecodeOne`, how many bytes were consumed
- Not yet used by the TCP server (the socket path still echoes raw bytes)

There is no persistence or Redis command set yet.

## Layout

```
main.go              # flags and process entry
config/config.go     # host and port
server/sync_tcp.go   # listen, accept, echo loop
core/resp.go         # RESP decode
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

In another terminal:

```bash
nc 127.0.0.1 7379
```

Type a line and press enter. The server echoes it back. Disconnect with `Ctrl-D` (or close the client); the server then accepts the next connection.

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

1. Feed accepted bytes through `core.Decode` instead of echoing them raw
2. Implement a small command set (`PING`, `GET`, `SET`, …)
3. Rebuild the same server in Java
