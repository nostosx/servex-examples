# Security

Security fixes target the latest v1 source snapshot and current development
code. Update to the latest snapshot before reporting a vulnerability.
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
Ordinary bugs can use public issues.
