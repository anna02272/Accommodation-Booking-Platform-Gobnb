FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o server

FROM alpine
COPY --from=build_container /app/server /usr/bin
COPY blacklist.txt .
COPY auth-service.crt /app/
COPY auth-service.key /app/
EXPOSE 8080
ENTRYPOINT ["server", "https"]