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
	echo "current path: $pwd"
	echo "script path: $scriptPath"
	
	echo "start netperf"
	netperf -c -C -l 60 -H $serverAddr -p $netperfPort -t TCP_RR -- -r 100,200 -P ",$(expr ${netperfPort} + 1)"> "${pwd}/${serverAddr}_${netperfPort}_netperf.txt";
	echo "finish netperf"	

	echo "start iperf3"
	iperf3 -c $serverAddr -p $iperfPort > "${pwd}/${serverAddr}_${iperfPort}_iperf.txt";
	echo "finish iperf3"

	echo "start redis-benchmark"
	redis-benchmark -h $serverAddr -p $redisPort --csv > "${pwd}/${serverAddr}_${redisPort}_redis.txt";
	echo "finish redis-benckmark"

	echo "start testChangeRequestSize"
	${scriptPath}/netLatencyTester/testChangeRequestSize.sh -q -ad "${serverAddr}:${latencyTesterPort}" -r 1000 -rp 1ms -cpus 1 > "${pwd}/${serverAddr}_${latencyTesterPort}_changeRequestSize.txt";
	echo "finish testChangeRequestSize"

	echo "start testChangeRequestPeriod"
	${scriptPath}/netLatencyTester/testChangeRequestPeriod.sh -q -ad "${serverAddr}:${latencyTesterPort}" -r 1000 -cpus 1 > "${pwd}/${serverAddr}_${latencyTesterPort}_changeRequestPeriod.txt";
	echo "finish testChangeRequestPeriod"
	
	echo "start large sample test"
	${scriptPath}/netLatencyTester/netLatencyTester -ad "${serverAddr}:${latencyTesterPort}" -rp 1ms -fc "${pwd}/${serverAddr}:${latencyTesterPort}_largeSample_conn.txt" -fr "${pwd}/${serverAddr}:${latencyTesterPort}_largeSample_rtt.txt" -r 100000 -cpus 1 > "${pwd}/${serverAddr}_${latencyTesterPort}_largeSample.txt";
	echo "finish large sample test"	

	echo "All work done!"
else
	echo "usage: testScript.sh serverAddr iperfPort redisPort netLatencyTesterPort"
	echo "eg. testScript.sh 127.0.0.1 2865 5201 6379 1234"
fi
