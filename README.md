# TCP proxy with SNI inspect

## How to use
- `cp config.toml.example config.toml`
- `go build -o r2`
- `sudo ./r2 -port=8080`

# Feature
- Inspect SNI for HTTPS proxy
- Can work as Router with custom DHCP and DNS

## Reference
- https://www.kernel.org/doc/Documentation/networking/tproxy.txt
- https://powerdns.org/tproxydoc/tproxy.md.html
- http://gsoc-blog.ecklm.com/iptables-redirect-vs.-dnat-vs.-tproxy/
- https://github.com/LiamHaworth/go-tproxy
