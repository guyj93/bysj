# exit when use unset variable
set -u

if [ $# -ge "2" ]; then
	testName=$1
	remote_IP=$2
	
	if [ $# -ge "6"]; then
		netperf_port=$3
		iperf3_port=$4
		redis_port=$5
		netLatencyTester=$6
	else
		netperf_port="2865"
		iperf3_port="5201"
		redis_port="6379"
		netLatencyTester="1234"
	fi
	
	pwd=$(pwd)
	scriptPath=$(cd $(dirname "${BASH_SOURCE[0]}");pwd;)
	resultPath=results/${testName}/$(date +'%Y%m%d_%H-%M')
	mkdir -p $resultPath
	cd $resultPath
	if [ $? -ne "0" ]; then
		exit
	fi

	echo "-----${remote_IP}-----"
	mkdir -p ${remote_IP}
	cd ${remote_IP}
	${scriptPath}/testTools/testScript.sh $remote_IP $netperf_port $iperf3_port $redis_port $netLatencyTester_port
	cd ..
	
	cd $pwd
	echo "-----finish-----"
	echo "The result is stored in ${resultPath}."
else
	echo "usage: standaloneTest.sh testName remote_IP [netperf_port iperf3_port redis_port netLatencyTester_port]"
fi
