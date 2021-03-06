scriptPath=$(cd $(/usr/bin/dirname "${BASH_SOURCE[0]}");pwd;)
echo "will start netperf, iperf3, redis-server and netLatencyTester servers"
/usr/local/bin/netserver -p 2865
/usr/bin/iperf3 -s -D
/usr/local/bin/redis-server $scriptPath/testTools/redis/redis.conf
$scriptPath/testTools/netLatencyTester/netLatencyTester -s -laddr ":1234" -cpus 1 > /dev/null &
sleep 1
