ARG GO_VERSION=1.19

## Build container
FROM docker.io/golang:${GO_VERSION} AS builder

WORKDIR /src

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /example_api /src/cmd/example_api

## Final container
FROM gcr.io/distroless/static-debian11 AS final

COPY --from=builder /example_api /

EXPOSE 8000

ENTRYPOINT ["/example_api"]
