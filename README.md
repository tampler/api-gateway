API gateway service

Features:

- REST input
- Requst validation
- JWT validation
- Various HTTP controls (req/resp timeout, req rate, req max size, etc)
- Implements NWS SDK Cloud Control API
- Integrates NATS JetStream as a Message Bus
- Integrates NATS JetStream as a key-value storage
- TBD: Integrates NATS JetStream as an object binary storage

### Dependencies

- NATS JetStream server
- Casdoor IAM
- NWS SDK

### Configuration

This uses `app.toml` to specify different params. For devevelopment, use HTTP without Authorization,
settings `app.toml` -> `auth.auth_enabled = false`.

For prod use, `Echo Server` provides JWT validation and extracts `userID` from the `JWT token`.
This `userID` is widely used in the system as a primary key to user identifiation.

### Testing

1. Launch `NATS JetStream` and `NWS SDK`
2. Run `go test -v -run Test_offerings` to allow caching `CloudStack` internals and IDs
3. Run all other tests as usual
