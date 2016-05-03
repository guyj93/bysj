# exit when use unset variable
set -u;
# exit after an error command
set -e;

if [ $# -eq "2" ]; then
	brName=$1;
	portName=$2;
	sudo brctl addbr $brName;
	sudo brctl addif $brName $portName;
	sudo dhclient -r $portName;
	sudo ifconfig $portName 0.0.0.0;
	sudo ifconfig $brName up;
	sudo dhclient $brName;
else
	echo "usage: createBridgeWithDhcp.sh bridgeName portName";
	echo "eg. createBridgeWithDhcp.sh br0 eth0";
fi
