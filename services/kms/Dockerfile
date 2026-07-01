FROM golang:1.26.4-alpine AS build

WORKDIR /app

ADD ./ /app

ENV CGO_ENABLED=0

RUN go mod download
RUN go build -o main ./cmd/litehsm/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /

EXPOSE 8080

COPY --from=build /app/main /main

CMD ["/main"]
