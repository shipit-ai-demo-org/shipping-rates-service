# shipping-rates-service

Carrier rate quoting engine for CargoCloud — real-time shipping rate
calculation across carriers. Go 1.22, stdlib-only.

## Role at CargoCloud

Checkout in [web-storefront](https://github.com/shipit-ai-demo-org/web-storefront)
calls this service on every shipping-option step to show customers live carrier
prices. [orders-api](https://github.com/shipit-ai-demo-org/orders-api) re-quotes
at order confirmation to lock the rate that actually gets billed.

## How quoting works

1. `POST /v1/quotes` receives a `Shipment` (origin/destination ZIP, weight,
   dimensions, service level).
2. The engine computes **billable weight** — the max of scale weight and
   dimensional weight (`L×W×H / 5000`).
3. Every registered carrier adapter (`internal/carriers`) prices the shipment
   off its negotiated rate card; a carrier failure degrades the comparison
   instead of failing the request.
4. Quotes are cached with a short TTL so repeated checkout re-quotes don't
   hammer carrier tariff APIs.
5. The cheapest rate is returned as `best`, with all carrier quotes as
   `alternatives`.

## API surface

```
GET  /healthz       liveness
POST /v1/quotes     rate a shipment across all carriers
```

### Example

```bash
curl -s localhost:8081/v1/quotes -d '{
  "originZip": "10001", "destZip": "94107",
  "weightKg": 2.4, "lengthCm": 30, "widthCm": 20, "heightCm": 15,
  "service": "ground"
}'
```

## Carriers

| Adapter | Rate basis |
| ------- | ---------- |
| `internal/carriers/ups.go` | Negotiated 2026 rate card |
| `internal/carriers/fedex.go` | 2026 list rates + volume discount + fuel surcharge |

## Local development

```bash
go run .
```
