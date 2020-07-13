# First make a new chain in the mangle table called DIVERT 
# and add a rule to direct any TCP traffic with a matching local socket to the DIVERT chain
iptables -t mangle -N DIVERT
iptables -t mangle -A PREROUTING -p tcp -m socket -j DIVERT

# Then in the DIVERT chain add rules to add routing mark of 1 to packets in the DIVERT chain
# and accept the packets
iptables -t mangle -A DIVERT -j MARK --set-mark 1
iptables -t mangle -A DIVERT -j ACCEPT

# And add routing rules to direct traffic with mark 1 to the local loopback device 
# so the Linux kernal can pipe the traffic into the existing socket.

ip rule add fwmark 1 lookup 100
ip route add local 0.0.0.0/0 dev lo table 100

# Finally add a IPTables rule to catch new traffic on any desired port 
# and send it to the TProxy server

# intranet enp0s8 192.168.10.0/24 via 192.168.10.1
# wan enp0s3 via 10.0.2.2, enable ip forward
iptables -t mangle -A PREROUTING -i enp0s8 -p tcp -j TPROXY --tproxy-mark 0x1/0x1 --on-port 8080
ip r add default via 10.0.2.2 dev enp0s3
ip r add 192.168.10.0/24 via 192.168.10.1

# ip 
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.all.forwarding=1
iptables -t nat -A POSTROUTING -o enp0s3 -j MASQUERADE
