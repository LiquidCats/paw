FROM golang:1.27.3-alpine AS build

ARG MODULE_NAME=rater

WORKDIR /app

ADD ./ /app

ENV CGO_ENABLED=0

RUN go mod download
RUN go build -o main ./cmd/${MODULE_NAME}/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /

EXPOSE 8080

COPY --from=build /app/main /main

CMD ["/main"]
