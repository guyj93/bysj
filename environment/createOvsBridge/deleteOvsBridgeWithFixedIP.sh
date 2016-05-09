# exit when use unset variable
set -u;
# exit after an error command
set -e;

if [ $# -eq "4" ]; then
	brName=$1;
	portName=$2;
	ipAddr=$3;
	gwAddr=$4;
	sudo ifconfig $brName 0.0.0.0 down;
	sudo ovs-vsctl del-br $brName;
	sudo ifconfig $portName $ipAddr;
	sudo route add default gw $gwAddr;
else
	echo "usage: deleteOvsBridgeWithFixedIP.sh bridgeName portName ipAddrWithMaskLen defaultGatewayAddr";
	echo "eg. deleteOvsBridgeWithFixedIP.sh br0 eth0 10.0.0.100/24 10.0.0.1";
fi
