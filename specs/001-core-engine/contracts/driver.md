# Contract: Driver Interface

**Date**: 2026-06-19
**Version**: v1.0
**Stability**: Stable
**Consumers**: `pkg/engine/Runner`
**Implementations**: `pkg/http.Driver` (v1), `pkg/ws.Driver` (future), `pkg/grpc.Driver` (future)

## Interface

```go
type Driver interface {
    // Name returns the human-readable protocol name.
    // Example: "http", "websocket", "grpc"
    Name() string

    // Execute sends a single Step and returns the Result.
    // ctx carries the request-scoped deadline and cancellation signal.
    // baseURL is the scenario's BaseURL field.
    // step contains method, path, headers, body, assertions to apply.
    // sc provides template data (VU.ID, VU.Iteration, Vars, Env).
    //
    // The driver MUST:
    //   - Build and send the request per the Step configuration.
    //   - Record the latency from before the request to after the response is fully read.
    //   - Run all assertions on the response.
    //   - Populate Result.Success based on assertion outcomes.
    //   - Populate Result.Error with a descriptive string on failure.
    //   - Handle nil Body gracefully (send empty body).
    //
    // The driver SHOULD NOT:
    //   - Modify the Step or StepContext.
    //   - Panic on transport errors (return error instead).
    Execute(ctx context.Context, baseURL string, step Step, sc *StepContext) (*Result, error)

    // Close releases any persistent resources (connections, channels).
    // Called once per scenario after all phases complete.
    // MUST be safe to call multiple times (idempotent).
    Close() error
}
```

## Semantics

### Execute Contract

1. **Timing**: `Result.StartTime` is recorded before any network activity. `Result.EndTime` is recorded after the response body is fully consumed or an error occurs. `Result.Latency = EndTime - StartTime`.
2. **Assertions**: The driver is responsible for running assertions because it has access to the response data. The engine does not run assertions separately.
3. **Error handling**:
   - Transport error (timeout, connection refused, DNS failure): `Result.Success = false`, `Result.Error = "transport: ..."`.
   - Assertion failure: `Result.Success = false`, `Result.Error = "assertion: ..."`.
   - Unexpected panic: driver MUST recover and return a failed Result.
4. **Body consumption**: For HTTP, the driver MUST consume and discard the response body to prevent connection pool leaks.

### Close Contract

1. Called exactly once per scenario, after all phases and teardown complete.
2. For HTTP: no-op (net/http.Client manages its own pool).
3. For WebSocket (future): closes the persistent connection.
4. For gRPC (future): closes the gRPC channel.

## Version History

| Version | Date | Changes |
|---|---|---|
| v1.0 | 2026-06-19 | Initial contract. HTTP driver implements this. |
