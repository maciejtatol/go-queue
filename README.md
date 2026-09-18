# go-queue

A small SQS-inspired message broker built to learn Go concurrency.

## Requirements

- Go 1.27 or newer

## Getting started

Start the HTTP server on port 8080:

```sh
go run .
```

Visit http://localhost:8080 or use `curl http://localhost:8080` to get
`Hello, Go!`. Stop the server with Ctrl+C.

Run the test suite and build the binary:

```sh
make
```

Other useful commands:

```sh
make test  # run tests
make lint  # run gofmt and go vet checks
make clean # remove local build artifacts
```

## Task producer and consumer

The server manually creates a buffered Go channel with capacity for 100 pending
 tasks and starts one background consumer. `POST /tasks` creates a task and the
producer publishes it with the fixed topic `tasks.created`. The consumer processes
it by logging its ID, topic, and name.

```sh
curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"name":"Send welcome email"}'
```

A successful request returns `202 Accepted` with the task's `id`, `name`, `topic`,
and `created_at`. Acceptance means the task was queued; processing happens
asynchronously. Invalid requests return `400`; a full queue returns `503`.

This is a local SQS-inspired learning example; no AWS account or dependencies are
needed. The topic is a message label, not a separate pub/sub broker. Tasks are
stored only in memory, with no persistence or retries. Ctrl+C stops accepting
requests and drains queued tasks before exiting; an abrupt exit loses pending work.
