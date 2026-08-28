# CobaltDB

CobaltDB is a lightweight in-memory key-value store written in Go, built as a
foundation for a distributed database.

## Project structure

```text
cobaltdb/
├── cmd/
│   ├── server/
│   └── cli/
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
go run ./cmd/server
```

The server listens on `127.0.0.1:6380` by default. Connect with Netcat in another Terminal:

```sh
nc 127.0.0.1 6380
```

Example session:

```text
SET name cobalt
OK
GET name
cobalt
EXISTS name
true
DELETE name
OK
GET name
(nil)
EXISTS name
false
```

You can also listen on a different address:

```sh
go run ./cmd/server -addr 127.0.0.1:7000
```

It currently accepts `SET`, `GET`, `DELETE`, and `EXISTS`, with each TCP
connection handled concurrently. A dedicated protocol package and interactive
CLI will be added in later tasks.
