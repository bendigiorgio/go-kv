# Go-KV

Go-KV is a simple key-value store written in Go. It provides an HTTP API for managing key-value pairs with additional features like flushing, compaction, and memory usage tracking.

## Features

- Set, get, delete, and list key-value pairs
- Batch operations for setting and deleting multiple keys
- Memory usage tracking and automatic flushing when memory limits are exceeded
- Data persistence across instances
- Data compaction to merge flushed data into the main database

- Dockerfile for easy deployment

## Installation

To install Go-KV, clone the repository and build the project:

```sh
git clone https://github.com/bendigiorgio/go-kv.git
cd go-kv
make build
```

## Usage

Start the Go-KV server in development mode:

```sh
make dev
```

The server will start on `http://localhost:8080`.
You can also access the Templ proxy for better hot reloading on `http://localhost:8081`.

## API Documentation

For detailed API documentation, please refer to the `openapi.yaml` file in the repository.

## License

This project is licensed under the MIT License.
