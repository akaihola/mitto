# fix: evaluate CSP per-request and preserve UI settings through RC file merge

Two related bugs fixed in this PR:

## Problem 1 – CSP evaluated once at server startup

`cspHandler` in `internal/web/csp_nonce.go` captured `allowExternalImages` as a
plain `bool` at server startup time. If the user later toggled the setting in the UI,
the updated value was never picked up. The Content Security Policy was therefore stale
for the lifetime of the server process.

**Fix:** Change the field to `func() bool` and evaluate it on every request by calling
`MittoConfig.Conversations.AreExternalImagesEnabled()` inside the handler closure.

## Problem 2 – RC file merge silently drops UI-saved settings

When settings are loaded from a `.mittorc` file, `mergeConfigs` in
`internal/config/settings.go` did not propagate `Conversations.ExternalImages` and
`Conversations.ActionButtons` from the base `settings.json` into the merged result.
Any value the user had saved via the Settings dialog was silently discarded on the
next config reload.

**Fix:** Explicitly copy those two fields in `mergeConfigs` so UI-persisted values
survive an RC file merge.

## Files changed

- `internal/config/settings.go` — copy `ExternalImages` and `ActionButtons` in merge
- `internal/web/csp_nonce.go` — `allowExternalImages bool` → `func() bool`
- `internal/web/server.go` — pass closure instead of captured bool
- `internal/web/csp_nonce_test.go` — update tests for new signature

## Testing

- All existing CSP tests pass with the updated `func() bool` signature.
- Manually verified: toggling external images in Settings now takes effect immediately
  without restarting the server.
