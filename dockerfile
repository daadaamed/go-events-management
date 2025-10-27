FROM golang:alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o events-api .

FROM alpine:latest
COPY --from=build /app/events-api /

EXPOSE 8080
CMD ["/events-api"]