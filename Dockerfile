# Development image: compile/run Go inside the container.
# Production multi-stage target is added in a later step.
FROM golang:1.26-alpine AS development

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "./cmd/web"]
