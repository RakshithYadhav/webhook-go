FROM golang:1.26.4 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/webhook-go ./cmd

FROM gcr.io/distroless/static:nonroot

COPY --from=build /out/webhook-go /webhook-go

ENTRYPOINT ["/webhook-go"]