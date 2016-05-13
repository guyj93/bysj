# exit when use unset variable
set -u

if [ $# -ge "4" ]; then
	testName=$1
	local_IP=$2
	lxcBridgeBr0_IP=$3
	kvmBridgeBr0_IP=$4

	if [ $# -ge "5" ]; then
		lxcNetworkDefault_IP=$5
	else
		lxcNetworkDefault_IP=$local_IP
	fi
	if [ $# -ge "6" ]; then
		kvmNetworkDefault_IP=$6
	else
		kvmNetworkDefault_IP=$local_IP
	fi
	
	pwd=$(pwd)
	scriptPath=$(cd $(dirname "${BASH_SOURCE[0]}");pwd;)
	resultPath=results/${testName}/$(date +'%Y%m%d_%H-%M')
	mkdir -p $resultPath
	cd $resultPath
	if [ $? -ne "0" ]; then
		exit
	fi

	echo "-----local-----"
	mkdir -p local
	cd local
	${scriptPath}/testTools/testScript.sh $local_IP 2865 5201 6379 1234
	cd ..

	echo "-----lxcNetworkDefault-----"
	mkdir -p lxcNetworkDefault
	cd lxcNetworkDefault
	${scriptPath}/testTools/testScript.sh $lxcNetworkDefault_IP 12865 15201 16379 11234
	cd ..

	echo "-----lxcBridgeBr0-----"
	mkdir -p lxcBridgeBr0
	cd lxcBridgeBr0
	${scriptPath}/testTools/testScript.sh $lxcBridgeBr0_IP 2865 5201 6379 1234
	cd ..

	echo "-----kvmNetworkDefault-----"
	mkdir -p kvmNetworkDefault
	cd kvmNetworkDefault
	${scriptPath}/testTools/testScript.sh $kvmNetworkDefault_IP 22865 25201 26379 21234
	cd ..

	echo "-----kvmBridgeBr0-----"
	mkdir -p kvmBridgeBr0
	cd kvmBridgeBr0
	${scriptPath}/testTools/testScript.sh $kvmBridgeBr0_IP 2865 5201 6379 1234
	cd ..
	
	cd $pwd
	echo "-----finish-----"
	echo "The result is stored in ${resultPath}."
else
	echo "usage: basicTest.sh testName local_IP lxcBridgeBr0_IP kvmBridgeBr0_IP [lxcNetworkDefault_IP] [kvmNetworkDefault_IP]"
	echo "Normally local_IP==lxcNetworkDefault_IP==kvmNetworkDefault_IP, so optionaly give them."
fi
