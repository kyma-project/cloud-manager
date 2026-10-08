#!/usr/bin/env bash
#
# filter-logs-pod.sh — reduce a verbose cloud-manager log export to a readable timeline.
#
# Intended for logs copied directly from a running cloud-manager pod (e.g.
# `kubectl logs`). Each line is a single JSON object with the log fields at the
# top level (no jsonPayload wrapper) and the timestamp under "time". The export
# is either a JSON array of entries or newline-delimited JSON (one object per
# line); this script accepts both.
#
# It extracts the fields most useful for triaging a reconcile timeline and drops
# the rest (labels, scope, stacktrace, per-controller payload fields, etc.).
#
# Usage:
#   ./tools/scripts/filter-logs-pod.sh <input.json> [output.json]
#
# Requires: jq
set -euo pipefail

input="${1:?usage: filter-logs-pod.sh <input.json> [output.json]}"
output="${2:-filtered-logs.json}"

# Pod logs come as either a top-level JSON array or newline-delimited JSON.
# `jq -s` slurps both into an array; if that array is itself a single wrapped
# array (the export-as-array case) we unwrap it before mapping.
jq -s '
  (if length == 1 and (.[0] | type) == "array" then .[0] else . end)
  | map({
      timestamp: .time,                 # pod logs use "time"
      severity,                          # top-level INFO / ERROR
      controller,                        # top-level in pod logs
      message,
      reconcileID,
      error                              # null on non-error entries
    })
' "$input" > "$output"

echo "Wrote $(jq length "$output") entries to $output"
