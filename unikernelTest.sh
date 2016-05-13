# exit when use unset variable
set -u

if [ $# -ge "2" ]; then
	testName=$1
	osvBridgeBr0_IP=$2
	
	pwd=$(pwd)
	scriptPath=$(cd $(dirname "${BASH_SOURCE[0]}");pwd;)
	resultPath=results/${testName}/$(date +'%Y%m%d_%H-%M')
	mkdir -p $resultPath
	cd $resultPath
	if [ $? -ne "0" ]; then
		exit
	fi

	echo "-----osvBridgeBr0-----"
	mkdir -p osvBridgeBr0
	cd osvBridgeBr0
	${scriptPath}/testTools/testScript.sh $osvBridgeBr0_IP 2865 5201 6379 1234
	cd ..
	
	cd $pwd
	echo "-----finish-----"
	echo "The result is stored in ${resultPath}."
else
	echo "usage: unikernelTest.sh testName osvBridgeBr0_IP"
fi
