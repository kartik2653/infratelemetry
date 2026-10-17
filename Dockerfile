FROM golang:1.23-bookworm AS api
WORKDIR /src
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM node:20-bookworm AS web
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
ENV NEXT_PUBLIC_API_BASE_URL=
ENV BACKEND_URL=http://127.0.0.1:8080
RUN npm run build

FROM node:20-bookworm-slim
WORKDIR /app
ENV NODE_ENV=production
ENV ADDR=:8080
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
COPY --from=api /out/server /app/server
COPY --from=web /src/.next/standalone /app/web
COPY --from=web /src/.next/static /app/web/.next/static
EXPOSE 3000
CMD ["sh", "-c", "/app/server & exec node /app/web/server.js"]
