sudo umount /var/run/netns/*
sudo rm /var/run/netns/*
sudo find /var/lib/cni/networks/mynet/ -name "10.22.0.*" -exec rm {} +
sudo iptables -F
sudo iptables -X
sudo iptables -t nat -F
sudo iptables -t nat -X
sudo iptables -t mangle -F
sudo iptables -t mangle -X
sudo iptables -P INPUT ACCEPT
sudo iptables -P FORWARD ACCEPT
sudo iptables -P OUTPUT ACCEPT

