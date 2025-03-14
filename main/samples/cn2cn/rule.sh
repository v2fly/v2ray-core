# value
MARK=0Xff
REDIRECT_PORT=3456
LOCAL_NETWORK="10.0.60.0/24 192.168.0.0/16"

# policy route
ip rule add fwmark 1 table 100 
ip route add local 0.0.0.0/0 dev lo table 100

# Framwork
## basic
iptables -t mangle -N V2RAY
iptables -t mangle -A V2RAY -d 127.0.0.1/32 -j RETURN
iptables -t mangle -A V2RAY -d 224.0.0.0/4 -j RETURN 
iptables -t mangle -A V2RAY -d 255.255.255.255/32 -j RETURN
iptables -t mangle -A V2RAY -j RETURN -m mark --mark ${MARK}  

## Main rule
for network in $LOCAL_NETWORK; do
	iptables -t mangle -A V2RAY -d ${network} -p tcp -j RETURN 
	iptables -t mangle -A V2RAY -d ${network} -p udp ! --dport 53 -j RETURN 
done

## redirect
iptables -t mangle -A V2RAY -p udp -j TPROXY --on-ip 127.0.0.1 --on-port ${REDIRECT_PORT} --tproxy-mark 1 
iptables -t mangle -A V2RAY -p tcp -j TPROXY --on-ip 127.0.0.1 --on-port ${REDIRECT_PORT} --tproxy-mark 1 

# apply
iptables -t mangle -A PREROUTING -j V2RAY 
