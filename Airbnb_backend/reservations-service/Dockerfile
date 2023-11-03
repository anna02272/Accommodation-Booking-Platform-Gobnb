FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o SOA_NoSQL_IB-MRS-2023-2024

FROM alpine
COPY --from=build_container /app/SOA_NoSQL_IB-MRS-2023-2024 /usr/bin/SOA_NoSQL_IB-MRS-2023-2024
EXPOSE 8080
ENTRYPOINT ["SOA_NoSQL_IB-MRS-2023-2024"]
