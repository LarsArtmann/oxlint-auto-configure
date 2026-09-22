# Security Policy

## Supported Versions

Only the latest tagged release receives security fixes:

| Version | Supported |
| ------- | --------- |
| latest release on [GitHub Releases](https://github.com/LarsArtmann/oxlint-auto-configure/releases) | ✅ |
| older releases / `master` between releases | ❌ |

## Reporting a Vulnerability

**Do not open a public issue for security vulnerabilities.**

Use GitHub's private vulnerability reporting:
<https://github.com/LarsArtmann/oxlint-auto-configure/security/advisories/new>

Include what you can of:

- A description of the issue and its impact
- Steps to reproduce (config, project layout, command line)
- The version or commit you tested against
- Any known workarounds

You will get an acknowledgment within 7 days, and a fix or a mitigation plan
within 90 days. Credit is given in the release notes unless you prefer to
stay anonymous.

## Scope

This tool generates `.oxlintrc.json` files. It runs `oxlint` as a subprocess
and writes a single config file — reports about oxlint itself belong in the
[oxlint repository](https://github.com/oxc-project/oxc/security/policy).
