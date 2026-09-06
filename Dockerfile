FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/llmtrim ./cmd/llmtrim

FROM scratch
COPY --from=build /out/llmtrim /llmtrim
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/llmtrim", "serve"]
