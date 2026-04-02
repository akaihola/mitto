# Tailscale Funnel

[Tailscale Funnel](https://tailscale.com/kb/1223/tailscale-funnel/) exposes your server
through Tailscale's network with automatic HTTPS certificates.

## Prerequisites

1. Install [Tailscale](https://tailscale.com/download)
2. Enable Funnel in your Tailscale admin console
3. Run `tailscale up` to connect to your tailnet

## Configuration

```yaml
web:
  port: 8080
  hooks:
    up:
      command: "tailscale funnel ${PORT}"
      name: "tailscale-funnel"
```

## Benefits

- **Automatic HTTPS** - Valid TLS certificates without configuration
- **No separate auth needed** - Tailscale handles identity (optional)
- **Tailnet integration** - Accessible only to your Tailscale network or publicly
- **No account/limits** - Part of your Tailscale plan

### Performance

Tailscale Funnel may feel slower than alternatives like Cloudflare Tunnel. This is
because Tailscale is designed primarily for mesh VPN (direct WireGuard connections
between devices), and Funnel adds HTTP proxying through Tailscale's DERP relay
network. DERP relays have fewer points of presence than Cloudflare's 300+ edge
cities, so traffic often travels further before reaching a relay. Additionally,
the WireGuard encapsulation adds overhead for HTTP traffic that isn't present in
purpose-built HTTP tunnel solutions. For latency-sensitive use, consider
[Cloudflare Tunnel](cloudflare.md) as an alternative.

## Security Considerations

- You can restrict access to your Tailnet only (no public exposure)
- ACLs control who can access the funnel
- No additional Mitto authentication required if restricted to trusted users
