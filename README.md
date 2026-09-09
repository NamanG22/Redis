# Redis

A from-scratch Redis clone. This repo starts with a TCP server — the networking layer Redis sits on — then protocol parsing, commands, and storage. The long-term target is a Java Redis server; this first slice is in Go so the socket model is small and easy to inspect.

## Current status

The TCP path, RESP codec, and first command are wired together. `main` starts the **async** server.

**Async TCP server** (default, macOS / BSD `kqueue`)

- Listens on `0.0.0.0:7379` by default (same port family as Redis, offset so it does not collide with a real Redis on `6379`)
- Non-blocking sockets + `kqueue` so many clients can be ready at once (up to 20,000 registered FDs)
- Accepts a connection, registers its FD, then on each readable event decodes a RESP array into a `RedisCmd` and writes a RESP reply
- I/O on raw FDs goes through `core.FDComm` (`Read` / `Write` via `syscall`)

**Synchronous TCP server** (still in tree, not started by `main`)

- Accepts one connection at a time; the accept loop is blocked while that client is being served
- Same decode → eval → reply path, over `net.Conn`

**RESP codec** (`core.Decode` / `core.DecodeOne` / `core.Encode`)

- Parses simple strings (`+`), errors (`-`), integers (`:`), bulk strings (`$`), and arrays (`*`), including nested arrays
- `DecodeArrayString` flattens a RESP array into `[]string` for command tokens
- `Encode` writes simple strings (`+PONG`) and bulk strings (`$5\r\nhello\r\n`)

**Commands**

- `PING` → `+PONG`
- `PING <message>` → bulk-string echo of `<message>`
- Wrong arity for `PING` → RESP error

There is no key-value store, persistence, or further command set yet. The async path uses `kqueue`, so it is not portable to Linux (`epoll`) yet.

## Layout

```
main.go              # flags; starts RunAsyncTCPServer
config/config.go     # host and port
server/async_tcp.go  # kqueue listen / accept / read loop
server/sync_tcp.go   # blocking listen / accept / read loop
core/comm.go         # FDComm (syscall Read/Write)
core/cmd.go          # RedisCmd (command + args)
core/eval.go         # command dispatch (PING)
core/resp.go         # RESP encode / decode
core/resp_test.go    # table-driven decode tests
go.mod
```

## Prerequisites

- Go 1.27+
- macOS or BSD (async server uses `kqueue`)

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
Starting async TCP server on port 0.0.0.0 7379
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

Disconnect with `Ctrl-D` (or close the client). Other connections can stay open; the event loop keeps serving them.

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
3. Linux `epoll` (or a portable multiplexer) so the async server is not macOS-only
4. Rebuild the same server in Java
