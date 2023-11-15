# build stage
FROM golang:1.21-alpine AS build
WORKDIR /build
COPY . .
RUN go get -d -v ./...
RUN go build -o app -v

# final stage
FROM alpine:latest

WORKDIR /dist
COPY --from=build /build .
COPY config.yaml .
EXPOSE 3000
CMD ["/dist/app", "app", "-c", "config.yaml"]