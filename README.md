# CobaltDB

CobaltDB is a lightweight in-memory key-value store written in Go, built as a
foundation for a distributed database.

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

The server listens on `127.0.0.1:6380` by default. You can also listen on a
different address:

```sh
cobalt server -addr=127.0.0.1:7000
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
place `-addr` before the operation:

```sh
cobalt kv -addr=127.0.0.1:7000 get name
```
