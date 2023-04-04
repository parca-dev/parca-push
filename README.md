![Build](https://github.com/parca-dev/parca-push/actions/workflows/build.yml/badge.svg)
![Container](https://github.com/parca-dev/parca-push/actions/workflows/container.yml/badge.svg)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# parca-push

A command line utility to push individual pprof formatted profiles to a Parca compatible API.

## Configuration

Flags:

[embedmd]:# (dist/help.txt)
```txt
Usage: parca-push <path>

Arguments:
  <path>    Path to the profile data.

Flags:
  -h, --help                     Show context-sensitive help.
  -l, --labels=KEY=VALUE;...     Labels to attach to the profile data.
                                 For example --labels=__name__=process_cpu
                                 --labels=node=foo
      --normalized               Whether the profile sample addresses are
                                 already normalized by the mapping offset.
      --remote-store-address=STRING
                                 gRPC address to send profiles and symbols to.
      --remote-store-bearer-token=STRING
                                 Bearer token to authenticate with store.
      --remote-store-bearer-token-file=STRING
                                 File to read bearer token from to authenticate
                                 with store.
      --remote-store-insecure    Send gRPC requests via plaintext instead of
                                 TLS.
      --remote-store-insecure-skip-verify
                                 Skip TLS certificate verification.
```
