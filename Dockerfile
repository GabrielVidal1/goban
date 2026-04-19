FROM node:22-alpine AS node-builder
WORKDIR /app/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=node-builder /app/ui/dist ./ui/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /kanban-ui .

FROM scratch
COPY --from=go-builder /kanban-ui /kanban-ui
VOLUME ["/kanban"]
EXPOSE 8080
ENTRYPOINT ["/kanban-ui"]
