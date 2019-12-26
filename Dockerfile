FROM golang:1.13 as builder

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o slack-notify .


FROM drone/ca-certs

WORKDIR /apps
COPY README.md .
COPY resources .

COPY --from=builder /src/slack-notify . 

CMD ["/apps/slack-notify"]
