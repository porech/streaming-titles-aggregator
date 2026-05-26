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
