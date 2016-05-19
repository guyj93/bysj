set -u
pattern=$1
dumplist=$(ls *.pcap)
for dump in $dumplist; do
	echo "============${dump}=============="
	tcpdump -r $dump | grep $pattern
	echo "============${dump}=============="
	echo ""
done
	echo "============Finish==============="
