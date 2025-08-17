# mock-openai

A tiny mock OpenAI-compatible server for local integration testing.

Run:

```sh
go run ./mock-openai -addr :8080
```

Endpoints:
- POST /v1/chat/completions
  - Query `?style=chunked` returns raw chunked bytes
  - Otherwise returns SSE-style `data:` events
