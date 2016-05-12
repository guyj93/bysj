# exit when use unset variable
set -u

if [ $# -eq "2" ]; then
	lxcBridgeOvsbr0_IP=$1
	kvmBridgeOvsbr0_IP=$2
	
	pwd=$(pwd)
	scriptPath=$(cd $(dirname "${BASH_SOURCE[0]}");pwd;)
	resultPath=results/ovs/$(date +'%Y%m%d_%H-%M')
	mkdir -p $resultPath
	cd $resultPath

	echo "-----lxcBridgeOvsbr0-----"
	mkdir -p lxcBridgeOvsbr0
	cd lxcBridgeOvsbr0
	${scriptPath}/testTools/testScript.sh $lxcBridgeOvsbr0_IP 2865 5201 6379 1234
	cd ..

	echo "-----kvmBridgeOvsbr0-----"
	mkdir -p kvmBridgeOvsbr0
	cd kvmBridgeOvsbr0
	${scriptPath}/testTools/testScript.sh $kvmBridgeOvsbr0_IP 2865 5201 6379 1234
	cd ..
	
	cd $pwd
	echo "-----finish-----"
	echo "The result is stored in ${resultPath}."
else
	echo "usage: ovsTest.sh lxcBridgeOvsbr0_IP kvmBridgeOvsbr0_IP"
fi
