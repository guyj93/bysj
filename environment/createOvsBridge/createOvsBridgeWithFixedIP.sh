# exit when use unset variable
set -u;
# exit after an error command
set -e;

if [ $# -eq "4" ]; then
	brName=$1;
	portName=$2;
	ipAddr=$3;
	gwAddr=$4;
	sudo ovs-vsctl add-br $brName;
	sudo ovs-vsctl add-port $brName $portName;
	sudo ifconfig $portName 0.0.0.0;
	sudo ifconfig $brName $ipAddr up;
	sudo route add default gw $gwAddr;
else
	echo "usage: createOvsBridgeWithFixedIP.sh bridgeName portName ipAddrWithMaskLen defaultGatewayAddr";
	echo "eg. createOvsBridgeWithFixedIP.sh br0 eth0 10.0.0.100/24 10.0.0.1";
fi
