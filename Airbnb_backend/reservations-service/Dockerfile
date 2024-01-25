FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o server

FROM alpine
COPY --from=build_container /app/server /usr/bin
COPY reservation-service.crt /app/
COPY reservation_decrypted_key.pem /app/
COPY gobnb-409715-26445f8b186e.json /app/
EXPOSE 8082
ENTRYPOINT ["server", "https"]