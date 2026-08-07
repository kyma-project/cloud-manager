#!/usr/bin/env bash
#
# filter-logs.sh — reduce a verbose cloud-manager log export to a readable timeline.
#
# Intended for logs downloaded from GCP Logs Explorer (Google Cloud Logging).
# In the Logs Explorer, run your query, then "Download" -> JSON. The export is
# either a JSON array of entries or newline-delimited JSON (one object per line);
# this script accepts both.
#
# It extracts the fields most useful for triaging a reconcile timeline and drops
# the rest (labels, resource, stacktrace, per-controller payload fields, etc.).
#
# Usage:
#   ./tools/scripts/filter-logs.sh <input.json> [output.json]
#
# Requires: jq
set -euo pipefail

input="${1:?usage: filter-logs.sh <input.json> [output.json]}"
output="${2:-filtered-logs.json}"

# GCP exports come as either a top-level JSON array or newline-delimited JSON.
# `jq -s` slurps both into an array; if that array is itself a single wrapped
# array (the export-as-array case) we unwrap it before mapping.
jq -s '
  (if length == 1 and (.[0] | type) == "array" then .[0] else . end)
  | map({
      timestamp,
      severity,                         # top-level INFO / ERROR
      controller: .jsonPayload.controller,
      message: .jsonPayload.message,
      reconcileID: .jsonPayload.reconcileID,
      error: .jsonPayload.error         # null on non-error entries
    })
' "$input" > "$output"

echo "Wrote $(jq length "$output") entries to $output"
