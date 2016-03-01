# Network forwarding [![GoDoc](https://godoc.org/github.com/goburrow/netforward?status.svg)](https://godoc.org/github.com/goburrow/netforward)

Forward network packets between various protocols, e.g. TCP <-> UDP, TLS <-> TCP.

## Install
```
go get github.com/goburrow/netforward/nf
```

## Usage
```
nf [OPTIONS]

OPTIONS:
  -address string
        listen address (default "localhost:7000")
  -caFile string
        client certificate authorities file
  -certFile string
        certificate file
  -keyFile string
        certificate key file
  -network string
        network protocol (default "tcp")
  -remote.address string
        remote address (default "localhost:8000")
  -remote.caFile string
        server certificate authorities file
  -remote.certFile string
        certificate file
  -remote.keyFile string
        certificate key file
  -remote.network string
        network protocol (default "tcp")
  -remote.skipVerify
        Not to verify remote server certificate
```
With default options, nf will divert TCP packets from localhost:7000 to localhost:8000

## Use cases
### Protect an unsecure HTTP server
- Many budget Internet of Things such as camera or sensor can only be accessed via a plain HTTP server.
- By using port forwarding, they can be viewed from Internet but it is a risk to expose password when login.
- VPN or SSL tunnel is too much overhead.

To mitigate, run a TLS endpoint in front of the unsecure HTTP server:
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /path/to/cert.key -out /path/to/cert.crt

nf -certFile /path/to/cert.crt -keyFile /path/to/cert.key -address :8443 -remote.address 127.0.0.1:8080
```
Then only forward port 8443 in your router.

Certificates can also be acquired from https://letsencrypt.org or self-signed https://github.com/OpenVPN/easy-rsa
