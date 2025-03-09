FROM golang:1.23 AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -mod=readonly -o /go/bin/app ./cmd/server/main.go

FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/app /
CMD ["/app"]