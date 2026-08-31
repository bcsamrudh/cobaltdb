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

## Run the server

```sh
go run ./cmd/cobalt server
```

The server listens on `127.0.0.1:6380` by default.

You can also listen on a different address:

```sh
go run ./cmd/cobalt server -addr=127.0.0.1:7000
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

Each TCP connection is handled concurrently. The `kv` client commands will be
added in the next task.

The planned client experience is:

```sh
cobalt kv put name samrudh
cobalt kv get name
cobalt kv exists name
cobalt kv delete name
```

> The `cobalt kv` commands are not implemented yet.
