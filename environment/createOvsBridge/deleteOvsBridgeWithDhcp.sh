# exit when use unset variable
set -u;
# exit after an error command
set -e;

if [ $# -eq "2" ]; then
	brName=$1;
	portName=$2;
	sudo dhclient -r $brName;
	sudo ifconfig $brName down;
	sudo ovs-vsctl del-br $brName;
	sudo dhclient $portName;
else
	echo "usage: deleteOvsBridgeWithDhcp.sh bridgeName portName";
	echo "eg. deleteOvsBridgeWithDhcp.sh br0 eth0";
fi
