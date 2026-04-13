---
title: Probe
weight: 11
---

Probe is a mechanism for measuring the available bitrate between two peers. One peer sends a probe request with a local bitrate estimate, and the remote peer responds with its own bitrate information. This enables adaptive bitrate decisions without maintaining a persistent subscription.

## Probe Bitrate

To probe the remote peer's bitrate, use the `(moqt.Session).Probe` method:

```go
    localBitrate := uint64(5_000_000) // 5 Mbps

    remoteBitrate, err := sess.Probe(localBitrate)
    if err != nil {
        // Handle error
        return
    }

    fmt.Printf("Remote bitrate: %d bps\n", remoteBitrate)
```

The method opens a bidirectional stream, sends a `PROBE` message with the local bitrate, and reads the remote peer's response containing the remote bitrate.

### Parameters

| Parameter      | Type     | Description                              |
|----------------|----------|------------------------------------------|
| `bitrate`      | `uint64` | The local bitrate estimate in bits per second |

### Return Values

| Value            | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| `remoteBitrate`  | `uint64` | The remote peer's bitrate in bits per second   |
| `err`            | `error`  | An error if the probe failed                   |

## Error Handling

Probe operations may return `moqt.ProbeError` or `moqt.SessionError`:

```go
    remoteBitrate, err := sess.Probe(localBitrate)
    if err != nil {
        var probeErr *moqt.ProbeError
        if errors.As(err, &probeErr) {
            switch probeErr.ProbeErrorCode() {
            case moqt.ProbeErrorCodeNotSupported:
                // Remote peer does not support probing
            case moqt.ProbeErrorCodeTimeout:
                // Probe timed out
            default:
                // Internal error
            }
        }
    }
```

See [Built-in Error Codes](errors/#built-in-error-codes) for `ProbeErrorCode` values.
