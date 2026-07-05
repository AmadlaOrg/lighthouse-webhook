# lighthouse-webhook

Lighthouse plugin for sending notifications via HTTP webhooks.

## Configuration

Set environment variables:
- `LIGHTHOUSE_WEBHOOK_URL` (required) - Webhook endpoint URL
- `LIGHTHOUSE_WEBHOOK_METHOD` - HTTP method (default: POST)
- `LIGHTHOUSE_WEBHOOK_TIMEOUT` - Request timeout (default: 10s)
- `LIGHTHOUSE_WEBHOOK_HEADER_*` - Custom headers (e.g. `LIGHTHOUSE_WEBHOOK_HEADER_Authorization=Bearer token`)

## Usage

```bash
echo '{"source":"test","name":"alert","severity":"warning"}' | lighthouse-webhook send
lighthouse-webhook info
```

## License

MIT
