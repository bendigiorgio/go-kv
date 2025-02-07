FROM golang:1.23 AS builder

WORKDIR /app

# Needs to install npm to be able to install the tailwind modules
RUN apt-get update && apt-get install -y npm

COPY go.mod go.sum ./

# Downloads Go packages
RUN go mod download

# Copies project's files to container
COPY . .
# Installs exact versions of tailwind dependencies as specified in package-lock.json
RUN npm ci
# Builds the single binary for a linux and amd64 architecture
# It also builds the tailwind css files
RUN make build

FROM alpine:latest
WORKDIR /root/

# Copies binary from previous container
COPY --from=builder /app/bin/main .

EXPOSE 8080

CMD ["./main"]