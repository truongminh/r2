
# vim /etc/network/interfaces
# # Configure the WAN port to get IP via DHCP
# auto eth0
# iface eth0 inet dhcp
# # Configure the LAN port
# auto eth1
# iface eth1 inet static
#     address 192.168.10.1 
#     netmask 255.255.255.0
ip a
systemctl restart networking
systemctl status networking

# dns
apt-get install bind9 
# dhcp
apt-get install isc-dhcp-server
# /etc/dhcp/dhcpd.conf
# subnet 192.168.10.0 netmask 255.255.255.0 {
  # option domain-name "lan";
  # option domain-name-servers 192.168.10.1,192.168.0.1;
  # authoritative;
  # range 192.168.10.100 192.168.10.200;
  # option routers 192.168.10.1;
# }
# setup ipv4 interface 
# vim /etc/default/isc-dhcp-server
systemctl restart isc-dhcp-server
