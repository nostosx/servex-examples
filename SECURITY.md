# Security

These examples are pre-release demos and have no supported release series.
The HTTP and gRPC demos use loopback defaults and do not implement application
authentication or TLS. They demonstrate lifecycle and transport APIs; exposure
to an untrusted network requires application security policy.

For a vulnerability in example wiring, use this repository's GitHub
**Security → Report a vulnerability** channel if available. If unavailable,
open an issue asking for a private contact without exploit details. Do not
include credentials, private sessions or access-bearing URLs. No response-time
guarantee is stated. Include the source revision and a synthetic reproducer.

For a defect in an upstream library, follow that library's SECURITY.md instead:
[servex](https://github.com/nostosx/servex/blob/main/SECURITY.md) or
[servex-contrib](https://github.com/nostosx/servex-contrib/blob/main/SECURITY.md).
These links identify the intended repository paths, not a claim that a remote
reporting channel is already enabled. Ordinary bugs can use public issues.
