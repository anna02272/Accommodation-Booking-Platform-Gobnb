FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o server

FROM alpine
COPY --from=build_container /app/server /usr/bin
COPY server.crt /app/
COPY decrypted_server.pem /app/
EXPOSE 8080
ENTRYPOINT ["server", "https"]