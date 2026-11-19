FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /orders-api .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=build /orders-api /orders-api
EXPOSE 8080
ENTRYPOINT ["/orders-api"]
