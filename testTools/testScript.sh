# exit when use unset variable
set -u;
# exit after an error command
#set -e;

if [ $# -eq "5" ]; then
	serverAddr=$1
	netperfPort=$2
	iperfPort=$3
	redisPort=$4
	latencyTesterPort=$5
	pwd=$(pwd)
	scriptPath=$(cd $(dirname "${BASH_SOURCE[0]}");pwd;)
	
	echo "start netperf"
	netperf -c -C -l 60 -H $serverAddr -p $netperfPort -t TCP_RR -- -r 100,200 -P ",$(expr ${netperfPort} + 1)"> "${pwd}/netperf_${serverAddr}:${netperfPort}.txt";	

	echo "start iperf3"
	iperf3 -c $serverAddr -p $iperfPort > "${pwd}/iperf3_${serverAddr}:${iperfPort}.txt";

	echo "start redis-benchmark"
	redis-cli -h $serverAddr -p $redisPort flushall > /dev/null
	redis-benchmark -h $serverAddr -p $redisPort --csv > "${pwd}/redis_${serverAddr}:${redisPort}.txt";

	echo "start testChangeRequestSize"
	${scriptPath}/netLatencyTester/testChangeRequestSize.sh -q -ad "${serverAddr}:${latencyTesterPort}" -r 1000 -rp 1ms -cpus 1 > "${pwd}/changeRequestSize_${serverAddr}:${latencyTesterPort}.txt";

	echo "start testChangeRequestPeriod"
	${scriptPath}/netLatencyTester/testChangeRequestPeriod.sh -q -ad "${serverAddr}:${latencyTesterPort}" -r 1000 -cpus 1 > "${pwd}/changeRequestPeriod_${serverAddr}:${latencyTesterPort}.txt";
	
	echo "start large sample test"
	${scriptPath}/netLatencyTester/netLatencyTester -ad "${serverAddr}:${latencyTesterPort}" -rp 1ms -r 100000 -cpus 1 -fc "${pwd}/largeSample_conn_${serverAddr}:${latencyTesterPort}.txt" -fr "${pwd}/largeSample_rtt_${serverAddr}:${latencyTesterPort}.txt"> "${pwd}/largeSample_${serverAddr}:${latencyTesterPort}.txt";

else
	echo "usage: testScript.sh serverAddr iperfPort redisPort netLatencyTesterPort"
	echo "eg. testScript.sh 127.0.0.1 2865 5201 6379 1234"
fi
