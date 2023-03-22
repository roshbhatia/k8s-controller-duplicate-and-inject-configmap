FROM golang:1.20

WORKDIR /app


COPY . .
RUN go mod download
RUN make build

CMD ["./bin/env-injector-controller"]
