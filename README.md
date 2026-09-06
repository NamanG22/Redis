# Redis

A from-scratch Redis clone. This repo starts with a TCP echo server — the networking layer Redis sits on — before protocol parsing, commands, and storage. The long-term target is a Java Redis server; this first slice is in Go so the socket model is small and easy to inspect.

## Current status

`TCPEchoServer` is a **synchronous TCP echo server**:

- Listens on `0.0.0.0:7379` by default (same port family as Redis, offset so it does not collide with a real Redis on `6379`)
- Accepts one connection at a time (the accept loop is blocked while a client is being served)
- Reads up to 512 bytes from the client and writes the same bytes back
- Logs connect, disconnect, and each received payload

There is no RESP parsing, persistence, or Redis commands yet.

## Layout

```
TCPEchoServer/
  main.go              # flags and process entry
  config/config.go     # host and port
  server/sync_tcp.go   # listen, accept, echo loop
  go.mod
```

## Prerequisites

- Go 1.27+

## Run

From the repo root:

```bash
cd TCPEchoServer
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

## What comes next

1. Keep the TCP accept/read/write path, then parse Redis Serialization Protocol (RESP)
2. Implement a small command set (`PING`, `GET`, `SET`, …)
3. Rebuild the same server in Java
