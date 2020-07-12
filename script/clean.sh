
# Clean old firewall
iptables -t nat -F
iptables -t nat -X
iptables -t mangle -F
iptables -t mangle -X
