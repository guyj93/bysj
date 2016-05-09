# exit when use unset variable
set -u;
# exit after an error command
set -e;

if [ $# -eq "2" ]; then
	brName=$1;
	portName=$2;
	sudo dhclient -r $brName;
	sudo ifconfig $brName down;
	sudo brctl delbr $brName;
	sudo dhclient $portName;
else
	echo "usage: deleteBridgeWithDhcp.sh bridgeName portName";
	echo "eg. deleteBridgeWithDhcp.sh br0 eth0";
fi
