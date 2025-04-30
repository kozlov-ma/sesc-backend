FROM golang:1.24

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/pgsetup/main.go .
ADD cmd/pgsetup/migrations ./migrations
RUN go build -o pgsetup ./main.go

ENTRYPOINT [ "./pgsetup" ]
