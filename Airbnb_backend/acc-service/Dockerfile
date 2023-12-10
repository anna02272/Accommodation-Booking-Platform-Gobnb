FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o server

FROM alpine
COPY --from=build_container /app/server /usr/bin
COPY accomm-service.crt /app/
COPY accomm_decrypted_key.pem /app/
EXPOSE 8083
ENTRYPOINT ["server", "https"]