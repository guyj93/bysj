# exit when use unset variable
set -u;
# exit after an error command
set -e;


if [ $# -eq "2" ]; then
	brName=$1;
	portName=$2;
	sudo ovs-vsctl add-br $brName;
	sudo ovs-vsctl add-port $brName $portName;
	sudo dhclient -r $portName;
	sudo ifconfig $portName 0.0.0.0;
	sudo ifconfig $brName up;
	sudo dhclient $brName;
else
	echo "usage: createOvsBridgeWithDhcp.sh bridgeName portName";
	echo "eg. createOvsBridgeWithDhcp.sh br0 eth0";
fi
