# simulate-skr-notification

A local development helper that fires a single fake **runtime-watcher** notification
at cloud-manager's SKR notification listener, so you can exercise the notification
sleeve without a real SKR, Istio, or the runtime-watcher agent.

## What it's for

In production, when a user changes a `cloud-resources.kyma-project.io` resource on an
SKR, the Kyma **runtime-watcher** agent POSTs a notification to cloud-manager's
`SKREventListener` in KCP. cloud-manager extracts the SKR's `runtime-id` (which equals
its `kymaName`) and enqueues it into the fast **notification sleeve** for prompt
reconciliation.

This script reproduces that HTTP POST locally.

## Prerequisites

- `openssl` (cert generation)
- `python3` (URL-encoding the certificate)
- `curl`
- cloud-manager running locally with the notification listener bound (default `:8083`).

## Usage

```bash
./tools/dev/simulate-skr-notification/simulate-skr-notification.sh <kymaName> [addr]
```

| Arg          | Meaning                                                        | Default                  |
|--------------|---------------------------------------------------------------|--------------------------|
| `<kymaName>` | The SKR runtime-id (== the KCP Kyma CR name) to notify.       | *(required)*             |
| `[addr]`     | Listener base URL.                                            | `http://localhost:8083`  |

Example:

```bash
./tools/dev/simulate-skr-notification/simulate-skr-notification.sh my-kyma-runtime-id
```

On success you get an `HTTP/1.1 200 OK` and, on the manager side, the
`cloud_manager_skr_looper_notification_received_total` metric increments and the SKR is
enqueued into the notification sleeve.

## How it works

The endpoint is:

```
POST {addr}/v2/cloud-manager/event
```

The path segment `cloud-manager` is the component name
(`looper.NotificationComponentName`), and must match the `manager:` field of the
Watcher CR (`config/watcher/watcher.yaml`).

The **critical, non-obvious detail**: the listener does **not** read the `runtime-id`
from the request body. It reads it from the **CommonName (CN) of the client
certificate** carried in the `X-Forwarded-Client-Cert` (XFCC) header. In production Istio
injects this header during mTLS termination; the runtime-watcher agent's cert has the
SKR's runtime-id as its CN.

So the script:

1. **Generates a throwaway self-signed cert** whose subject is `/CN=<kymaName>`.
   That CN is what the listener will surface as the `runtime-id`.
2. **URL-encodes the PEM** and wraps it as `Cert="<encoded>"` — the exact shape the
   listener's XFCC parser expects (it looks for the `Cert=` token and URL-unescapes it).
3. **POSTs a minimal `WatchEvent` JSON body**. cloud-manager ignores the body beyond
   identifying the SKR, but the listener still json-unmarshals it, so it must be valid
   JSON with `watched` / `watchedGvk` fields.

The cert and key are created in a temp dir and deleted on exit.

A plain `curl` with just a JSON body will **not** work — without the XFCC header the
listener returns `401 Unauthorized` ("could not get client certificate from request").

## Caveats

- **The `kymaName` must be an already-active SKR.** cloud-manager's `Notify` silently
  drops notifications for SKRs that are not currently active (activation happens only via
  the KCP reconciler when it processes the Kyma CR). If you notify an unknown/inactive
  `kymaName`, the listener still returns `200 OK` and
  `cloud_manager_skr_looper_notification_received_total` still increments (the adapter
  accepted it), but nothing is enqueued. Use a real, active KCP Kyma CR name.
- **This is a dev-only tool.** The forged cert bypasses the real mTLS trust chain; it only
  works because the local listener trusts whatever CN the XFCC header carries. Do not use
  against any non-local environment.
