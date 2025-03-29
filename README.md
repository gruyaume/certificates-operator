# certificates-operator

Manage X.509 certificates in the Juju ecosystem.

This charm serves as a proof-of-concept for developing charms in Go using [goops](https://github.com/gruyaume/goops).

## Usage

```shell
juju deploy certificates
juju deploy tls-certificates-requirer
juju integrate certificates tls-certificates-requirer
```
