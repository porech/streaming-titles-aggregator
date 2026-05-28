# Streaming Titles Aggregator

Aggregates currently-playing song titles from multiple streaming sources (Shoutcast, Icecast, Indiplay Stream+) and exposes them via a simple HTTP API.

## Features

- **Multi-source**: supports Shoutcast, Icecast, and Indiplay Stream+ sources in a single configuration. Support for additional providers is trivial to add: open an issue or send a PR if you need it
- **Per-stream endpoint**: fetch the current title of any configured stream at `GET /title/{name}.json`
- **Config hot-reload**: send `SIGHUP` to reload the configuration file without restarting — new streams are added, removed ones disappear, changed ones are updated, untouched ones keep working
- **Thread-safe**: requests are never interrupted during reload

## Configuration

Create a `config.json` file:

```json
{
    "listen_address": ":8080",
    "user_agent": "MyCustomAgent/1.0",
    "cors": {
        "allowed_origins": ["https://example.com"],
        "allowed_methods": ["GET", "OPTIONS"],
        "allowed_headers": ["Content-Type"],
        "exposed_headers": [],
        "allow_credentials": false,
        "max_age": 3600
    },
    "streams": {
        "example_shoutcast": {
            "kind": "shoutcast",
            "base_url": "http://shoutcast.example.com:7239",
            "sid": 1
        },
        "example_icecast": {
            "kind": "icecast",
            "base_url": "https://icecast.example.com",
            "mount": "/stream.aac"
        },
        "example_indiplay_id": {
            "kind": "indiplay",
            "base_url": "https://indiplay.example.com",
            "stream_id": "some-stream-id"
        },
        "example_indiplay_mount": {
            "kind": "indiplay",
            "base_url": "https://indiplay.example.com",
            "mount": "/stream"
        },
        "example_indiplay_mount_port": {
            "kind": "indiplay",
            "base_url": "https://indiplay.example.com",
            "mount": "/stream",
            "port": 8083
        }
    }
}
```

The top-level fields are all optional:

- `listen_address` – default `:8080`
- `user_agent` – default `streaming-titles-aggregator/1.0`; each stream can override it with its own `user_agent` field
- `cors` – disabled by default (no CORS headers emitted). See the **CORS** section below.

### CORS

If a top-level `cors` object is present, the server emits CORS headers on responses and handles `OPTIONS` preflight requests. If the section is absent, no CORS handling occurs (default).

| Field | Type | Default | Notes |
|---|---|---|---|
| `allowed_origins` | string array | — (required if `cors` is set) | Use `["*"]` for a wildcard, or list explicit origins (`["https://a", "https://b"]`). Matched as exact strings. |
| `allowed_methods` | string array | `["GET", "OPTIONS"]` | Sent in `Access-Control-Allow-Methods` on preflight. |
| `allowed_headers` | string array | — | Sent in `Access-Control-Allow-Headers` on preflight if non-empty. |
| `exposed_headers` | string array | — | Sent in `Access-Control-Expose-Headers` on normal responses if non-empty. |
| `allow_credentials` | bool | `false` | Sends `Access-Control-Allow-Credentials: true` when `true`. With `allowed_origins: ["*"]`, the server echoes the request `Origin` (since `*` + credentials is invalid per spec). |
| `max_age` | int (seconds) | `0` (header omitted) | Sent in `Access-Control-Max-Age` on preflight when greater than zero. Must be `>= 0`. |

When the request's `Origin` does not match any entry in `allowed_origins` (and the list is not `["*"]`), no CORS headers are emitted — the browser enforces the policy.

Changes to the `cors` section are picked up by `SIGHUP` reload without restarting the listener.

### Supported kinds

| Kind | Required fields | Optional fields | Notes |
|---|---|---|---|
| `shoutcast` | `base_url`, `sid` | — | Queries `/stats?sid=N&json=1` |
| `icecast` | `base_url`, `mount` | — | Queries `/status-json.xsl`; the mount is matched against the `listenurl` of each source in the response |
| `indiplay` | `base_url`, plus **one of** `stream_id` or `mount` | `port` | If `stream_id` is set it is used (and `mount`/`port` are ignored). Otherwise queries `/metadata/title?mount=...&port=...`, omitting `port` when not set |
| `icy` | `stream_url` | — | Connects to a stream URL, reads the first ICY metadata block and disconnects. Use only as a last resort: every title request produces a brief listen that will be counted by the streaming server. |

## Usage

```
streaming-titles-aggregator [config.json]
```

The only argument is an optional path to the configuration file. Defaults to `config.json` in the current directory.

### Examples

```sh
# Start the server with the default config.json
streaming-titles-aggregator

# Specify a different config file
streaming-titles-aggregator /etc/sta/config.json

# Fetch the current title for a stream
curl http://localhost:8080/title/example_shoutcast.json

# Response:
{"title":"Artist - Song Title"}
```

### Reloading configuration

```sh
kill -HUP <pid>
```

Or with systemd:

```sh
systemctl reload streaming-titles-aggregator
```

On reload the configuration file is re-read. Streams that were added, changed, or removed in the file are reflected immediately in subsequent requests.

## Install

### From source

```sh
go install github.com/porech/streaming-titles-aggregator/cmd/streaming-titles-aggregator@latest
```

### systemd

Copy `streaming-titles-aggregator.service` to `/etc/systemd/system/`, adjust paths and user, then:

```sh
systemctl daemon-reload
systemctl enable --now streaming-titles-aggregator
```

## License

MIT
