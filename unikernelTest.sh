# exit when use unset variable
set -u

if [ $# -eq "1" ]; then
	osvBridgeBr0_IP=$1
	
	pwd=$(pwd)
	scriptPath=$(cd $(dirname "${BASH_SOURCE[0]}");pwd;)
	resultPath=results/unikernel/`date -I'minutes'`
	mkdir -p $resultPath
	cd $resultPath

	echo "-----osvBridgeBr0-----"
	mkdir -p osvBridgeBr0
	cd osvBridgeBr0
	${scriptPath}/testTools/testScript.sh $osvBridgeBr0_IP 2865 5201 6379 1234
	cd ..
	
	cd $pwd
	echo "-----finish-----"
	echo "The result is stored in ${resultPath}."
else
	echo "usage: unikernelTest.sh osvBridgeBr0_IP"
fi
