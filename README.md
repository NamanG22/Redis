# Redis

A from-scratch Redis clone. This repo starts with a TCP server — the networking layer Redis sits on — then protocol parsing, commands, and storage. The long-term target is a Java Redis server; this first slice is in Go so the socket model is small and easy to inspect.

## Current status

The TCP path, RESP codec, in-memory store, and first commands are wired together. `main` starts the **async** server.

**Async TCP server** (default, macOS / BSD `kqueue`)

- Listens on `0.0.0.0:7379` by default (same port family as Redis, offset so it does not collide with a real Redis on `6379`)
- Non-blocking sockets + `kqueue` so many clients can be ready at once (up to 20,000 registered FDs)
- Accepts a connection, registers its FD, then on each readable event decodes one or more RESP values into `RedisCmds`, evaluates them, and writes a concatenated RESP reply
- I/O on raw FDs goes through `core.FDComm` (`Read` / `Write` via `syscall`)
- About once a second (when the event loop wakes) runs `core.CheckExpire` to sample and drop overdue keys

**Synchronous TCP server** (still in tree, not started by `main`)

- Accepts one connection at a time; the accept loop is blocked while that client is being served
- Same decode → eval → reply path, over `net.Conn`

**Pipelining**

- `core.Decode` walks the whole read buffer and returns every top-level RESP value (e.g. two command arrays from one `redis-cli --pipe` or pipelined write)
- Each value becomes a `RedisCmd`; replies are buffered and flushed in one `Write`

**RESP codec** (`core.Decode` / `core.DecodeOne` / `core.Encode`)

- Parses simple strings (`+`), errors (`-`), integers (`:`), bulk strings (`$`), and arrays (`*`), including nested arrays
- `Decode` returns `[]interface{}` (one entry per top-level value in the buffer)
- `Encode` writes simple strings (`+PONG`), bulk strings (`$5\r\nhello\r\n`), and integers (`:-1\r\n`)
- Command handlers return encoded `[]byte`; the server writes them (errors are encoded as bulk/simple strings today, not `-ERR` prefixes)

**In-memory store** (`core.Put` / `core.Get` / `core.Delete`)

- Process-local `map[string]*Obj` (value + optional expiry in unix milliseconds)
- **Lazy expire:** `Get` deletes the key if `ExpiresAt` is in the past
- **Active expire:** `CheckExpire` walks a sample of up to 20 keys and deletes overdue ones; it keeps sampling while more than 25% of the sample was expired (Redis-style)
- **Eviction:** `Put` calls `evict` when `len(store) >= config.MaxKeyLimit` (default **5**). The current policy deletes one arbitrary key (`evictFirst` — first key from map iteration). Not LRU / LFU, and not Redis `maxmemory`
- No disk persistence

**Commands**

- `PING` → `+PONG`
- `PING <message>` → bulk-string echo of `<message>`
- `SET key value` → `+OK`
- `SET key value EX seconds` → `+OK` with a TTL
- `GET key` → bulk string, or `$-1` if missing / expired
- `TTL key` → seconds remaining, `-1` if no expiry, `-2` if missing / expired
- `DEL key [key ...]` → count of keys removed
- `EXPIRE key seconds` → `1` if a TTL was set, `0` if the key is missing
- Unknown command or wrong arity → RESP error

There is no disk persistence or further command set yet. The async path uses `kqueue`, so it is not portable to Linux (`epoll`) yet.

## Layout

```
main.go              # flags; starts RunAsyncTCPServer
config/config.go     # host, port, MaxKeyLimit (default 5)
server/async_tcp.go  # kqueue listen / accept / read loop
server/sync_tcp.go   # blocking listen / accept / read loop
core/comm.go         # FDComm (syscall Read/Write)
core/cmd.go          # RedisCmd / RedisCmds (command batch)
core/store.go        # in-memory map + lazy expire on Get + evict on Put
core/expire.go       # sampled active expire (CheckExpire)
core/eviction.go     # evictFirst when at MaxKeyLimit
core/eval.go         # command dispatch (PING, SET, GET, TTL, DEL, EXPIRE)
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

```bash
redis-cli -p 7379 SET k v
redis-cli -p 7379 GET k
redis-cli -p 7379 SET tmp v EX 10
redis-cli -p 7379 TTL tmp
```

Expected: `OK`, `v`, `OK`, then a TTL of `10` or just under.

```bash
redis-cli -p 7379 EXPIRE k 30
redis-cli -p 7379 DEL k tmp
```

Expected: `1`, then `2` if both keys existed.

Or raw RESP over `nc`:

```bash
printf '*1\r\n$4\r\nPING\r\n' | nc 127.0.0.1 7379
```

Expected: `+PONG`

Pipelined raw RESP (two commands, one write):

```bash
printf '*1\r\n$4\r\nPING\r\n*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n' | nc 127.0.0.1 7379
```

Expected replies concatenated: `+PONG` then `+OK`.

Disconnect with `Ctrl-D` (or close the client). Other connections can stay open; the event loop keeps serving them.

Flags:

| Flag   | Default   | Meaning        |
|--------|-----------|----------------|
| `-host`  | `0.0.0.0` | bind address |
| `-port`  | `7379`    | bind port    |

`MaxKeyLimit` is `5` in `config/config.go` (not a flag). A sixth distinct `SET` evicts one existing key.

## Test

```bash
go test ./core/
```

## What comes next

1. More commands (`EXISTS`, …)
2. Better eviction (LRU / LFU) and a flag for `MaxKeyLimit`
3. Persistence (RDB / AOF)
4. Linux `epoll` (or a portable multiplexer) so the async server is not macOS-only
5. Rebuild the same server in Java
