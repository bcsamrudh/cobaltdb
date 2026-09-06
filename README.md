# CobaltDB

CobaltDB is a lightweight in-memory key-value store written in Go, built as a
foundation for a distributed database.

## Current status

Week 2 — write-ahead logging and persistence (`v0.2.0` in progress).

## Features

- Thread-safe in-memory key-value store
- Concurrent TCP clients
- Line-oriented text protocol
- `SET`, `GET`, `DELETE`, and `EXISTS` operations
- Unified `cobalt` server and client command
- Configurable client address through `COBALT_ADDR`
- Append-only write-ahead log with `fsync`
- Automatic recovery when the server restarts
- Recovery from an incomplete final WAL write
- Atomic snapshots and WAL compaction

## Project structure

```text
cobaltdb/
├── cmd/
│   └── cobalt/
├── internal/
│   ├── server/
│   ├── protocol/
│   ├── client/
│   └── store/
├── tests/
├── go.mod
├── README.md
├── LICENSE
└── .gitignore
```

## Testing

```sh
go test ./...
go test -race ./...
```

The test suite includes a socket-level workload with 100 concurrent clients and
more than 20,000 protocol operations. It also covers the store, protocol,
server, client, and command-line behavior.

## Install

```sh
go install ./cmd/cobalt
```

This installs the `cobalt` binary in Go's binary directory. Make sure that
directory is included in your `PATH`.

## Run the server

```sh
cobalt server
```

The server listens on `127.0.0.1:6380` and stores its WAL in `./data` by
default. Choose another data directory or address with flags:

```sh
cobalt server -data-dir=/var/lib/cobaltdb -addr=127.0.0.1:7000
```

For an ephemeral in-memory server whose data is discarded on shutdown:

```sh
cobalt server -dev
```

## Text protocol

CobaltDB uses newline-delimited UTF-8 requests and responses.

| Request | Success response | Missing-key response |
| --- | --- | --- |
| `SET key value` | `OK` | — |
| `GET key` | `VALUE value` | `NOT_FOUND` |
| `DELETE key` | `DELETED` | `NOT_FOUND` |
| `EXISTS key` | `EXISTS` | `NOT_EXISTS` |

Invalid requests return `ERROR message`. Command names are case-insensitive,
keys cannot contain spaces, and `SET` values may contain spaces.

Each TCP connection is handled concurrently.

## Use the client

With the server running in another terminal:

```sh
cobalt kv put name sam
# OK

cobalt kv get name
# sam

cobalt kv exists name
# true

cobalt kv delete name
# OK
```

Client commands connect to `127.0.0.1:6380` by default. To use another server,
set `COBALT_ADDR` once in your shell:

```sh
export COBALT_ADDR=127.0.0.1:7000
cobalt kv get name
```

For a one-off client command, `-addr` can still override `COBALT_ADDR`:

```sh
cobalt kv -addr=127.0.0.1:8000 get name
```

## Architecture

```text
cobalt command -> TCP client -> TCP server -> protocol -> concurrent store
```

The storage engine is independent of networking. The protocol package owns
command parsing and wire responses, while the server gives every connection
its own goroutine and shares one store protected by `sync.RWMutex`.

WAL records are newline-terminated JSON. During recovery, CobaltDB discards an
incomplete final record that may result from a crash. A malformed complete
record still stops startup so earlier corruption is never silently ignored.

On a clean shutdown, CobaltDB atomically writes `cobalt.snapshot` and resets the
WAL. Recovery loads that snapshot first and then replays any newer WAL records.

## Roadmap

- [x] In-memory key-value store
- [x] TCP server
- [x] Concurrent clients
- [x] Text protocol
- [x] Unified command-line client
- [x] Write-ahead log and restart recovery
- [x] Snapshots
- [ ] Replication
- [ ] Consistent hashing
- [ ] Failure detection and cluster management
