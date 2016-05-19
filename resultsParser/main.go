// resultParser project main.go
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Table struct {
	Name          string
	MergeFunc     func(location, test, timestamp, deployment, path string, to io.Writer) error
	PrintHeadFunc func(to io.Writer) error
	fp            *os.File
	bw            *bufio.Writer
	lock          sync.Mutex
}

func (t *Table) Merge(location, test, timestamp, deployment, path string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.MergeFunc(location, test, timestamp, deployment, path, t.bw)
}

func (t *Table) PrintHead() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.PrintHeadFunc(t.bw)
}

func NewTable(
	name string,
	mergeFunc func(location, test, timestamp, deployment, path string, to io.Writer) error,
	printHeadFunc func(to io.Writer) error,
) *Table {
	return &Table{
		Name:          name,
		MergeFunc:     mergeFunc,
		PrintHeadFunc: printHeadFunc,
	}
}

type StringSlice []string

func (p StringSlice) Len() int {
	return len(p)
}
func (p StringSlice) Less(i, j int) bool {
	li := len(p[i])
	lj := len(p[j])
	if li == lj {
		return p[i] < p[j]
	} else {
		return li < lj
	}
}
func (p StringSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type TableSlice []*Table

func (p TableSlice) Len() int {
	return len(p)
}
func (p TableSlice) Less(i, j int) bool {
	li := len(p[i].Name)
	lj := len(p[j].Name)
	if li == lj {
		return p[i].Name < p[j].Name
	} else {
		return li < lj
	}
}
func (p TableSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func MergeResults(resultsDir, mergedResultsDir string, locations, tests, deployments []string, tables []*Table) {
	sort.Sort(sort.Reverse(StringSlice(locations)))
	sort.Sort(sort.Reverse(StringSlice(tests)))
	sort.Sort(sort.Reverse(StringSlice(deployments)))
	sort.Sort(sort.Reverse(TableSlice(tables)))

	err := os.MkdirAll(mergedResultsDir, 0777)
	if err != nil {
		panic(err)
	}
	defer func(tables []*Table) {
		for _, t := range tables {
			if t.bw != nil {
				t.bw.Flush()

			}
			if t.fp != nil {
				t.fp.Close()
			}
		}
	}(tables)
	for _, t := range tables {
		fp, err := os.Create(mergedResultsDir + "/" + t.Name + ".csv")
		if err != nil {
			panic(err)
		}
		t.fp = fp
		t.bw = bufio.NewWriter(fp)
		if _, err = t.bw.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil { //write BOM of UTF8 for windows
			panic(err)
		}
		if err = t.PrintHead(); err != nil {
			panic(err)
		}

	}

	processResultFunc := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file := filepath.Base(path)
		tmp := filepath.Dir(path)
		deployment := filepath.Base(tmp)
		tmp = filepath.Dir(tmp)
		timestamp := filepath.Base(tmp)
		tmp = filepath.Dir(tmp)
		testAndLocation := filepath.Base(tmp)

		for _, l := range locations {
			if strings.Contains(testAndLocation, l) {
				for _, t := range tests {
					if strings.Contains(testAndLocation, t) {
						for _, d := range deployments {
							if strings.Contains(deployment, d) {
								for _, tab := range tables {
									if strings.Contains(file, tab.Name) {
										fmt.Println(path)
										if err := tab.Merge(l, t, timestamp, d, path); err != nil {
											return err
										}
										break
									}
								}
								break
							}
						}
						break
					}
				}
				break
			}
		}
		return nil
	}
	if err = filepath.Walk(resultsDir, processResultFunc); err != nil {
		panic(err)
	}

}

func main() {
	tables := make([]*Table, 0)
	var netperfMerge = func(location, test, timestamp, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(string(data), "\n")
		words := strings.Fields(lines[6])
		tps := words[5]
		_, err = to.Write([]byte(location + "," + test + "," + timestamp + "," + deployment + "," + tps + "\n"))
		return err
	}
	var netperfHead = func(to io.Writer) error {
		_, err := to.Write([]byte("Location,Test,timestamp,Deployment,Tps\n"))
		return err
	}
	tables = append(tables, NewTable("netperf", netperfMerge, netperfHead))

	var iperf3Merge = func(location, test, timestamp, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		tail := strings.LastIndex(content, "sender")
		content = content[:tail]
		head := strings.LastIndex(content, "\n") + 1
		words := strings.Fields(content[head:])
		retry := words[len(words)-1]
		unit := words[len(words)-2]
		bandwidth := words[len(words)-3]
		bandwidthFloat, err := strconv.ParseFloat(bandwidth, 64)
		if err != nil {
			return err
		}
		switch unit {
		case "Gbits/sec":
			bandwidthFloat *= 1000
		case "Mbits/sec":
			bandwidthFloat *= 1
		case "Kbits/sec":
			bandwidthFloat /= 1000
		default:
			return errors.New("Unrecognized unit: " + unit)
		}
		_, err = to.Write([]byte(location + "," + test + "," + timestamp + "," + deployment + "," + strconv.FormatFloat(bandwidthFloat, 'E', 4, 64) + "," + retry + "\n"))
		return err
	}
	var iperf3Head = func(to io.Writer) error {
		_, err := to.Write([]byte("Location,Test,timestamp,Deployment,Bandwidth,Retry\n"))
		return err
	}
	tables = append(tables, NewTable("iperf3", iperf3Merge, iperf3Head))

	var redisMerge = func(location, test, timestamp, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.SplitAfter(string(data), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				if _, err = to.Write([]byte(location + "," + test + "," + timestamp + "," + deployment + ",")); err != nil {
					return err
				}
				if _, err = to.Write([]byte(line)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	var redisHead = func(to io.Writer) error {
		_, err := to.Write([]byte("Location,Test,timestamp,Deployment,Command,Tps\n"))
		return err
	}
	tables = append(tables, NewTable("redis", redisMerge, redisHead))

	var changeRequestSizeMerge = func(location, test, timeStamp, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.SplitAfter(string(data), "\n")
		for _, line := range lines {
			if trimed := strings.TrimSpace(line); trimed != "" {
				words := strings.Fields(trimed)
				if _, err = to.Write([]byte(location + "," + test + "," + timeStamp + "," + deployment)); err != nil {
					return err
				}
				for _, word := range words {
					if du, err := time.ParseDuration(word); err != nil || word == "0" {
						if _, err = to.Write([]byte("," + word)); err != nil {
							return err
						}
					} else { //is a duration
						if _, err = to.Write([]byte("," + strconv.FormatFloat(float64(du), 'E', -1, 64))); err != nil {
							return err
						}
					}
				}
				if _, err = to.Write([]byte("\n")); err != nil {
					return err
				}
			}
		}
		return nil
	}
	var changeRequestSizeHead = func(to io.Writer) error {
		_, err := to.Write([]byte("Location,Test,timestamp,Deployment,RequestSize,NumValidRtt,MinRtt,AvgRtt,MaxRtt,StdRtt,Tps,TxBandwidth,RxBandwidth\n"))
		return err
	}
	tables = append(tables, NewTable("changeRequestSize", changeRequestSizeMerge, changeRequestSizeHead))

	var changeRequestPeriodMerge = changeRequestSizeMerge
	var changeRequestPeriodHead = func(to io.Writer) error {
		_, err := to.Write([]byte("Location,Test,timestamp,Deployment,RequestPeriod,NumValidRtt,MinRtt,AvgRtt,MaxRtt,StdRtt,Tps,TxBandwidth,RxBandwidth\n"))
		return err
	}
	tables = append(tables, NewTable("changeRequestPeriod", changeRequestPeriodMerge, changeRequestPeriodHead))

	var largeSampleRttMerge = func(location, test, timestamp, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.SplitAfter(string(data), "\n")
		lines = lines[1:] //ignore the first line of table head

		rttNumLimit := 10000
		var du time.Duration
		for key, line := range lines {
			if !(key < rttNumLimit) {
				break
			}
			if trimed := strings.TrimSpace(line); trimed != "" {
				words := strings.Fields(trimed)
				if len(words) != 3 {
					return errors.New("File format error!")
				}
				if du, err = time.ParseDuration(words[2]); err != nil {
					return err
				}
				//conn id
				if _, err = to.Write([]byte(location + "," + test + "," + timestamp + "," + deployment)); err != nil {
					return err
				}
				//rtt id
				if _, err = to.Write([]byte("," + words[0] + "," + words[1])); err != nil {
					return err
				}
				//rtt
				if _, err = to.Write([]byte("," + strconv.FormatFloat(float64(du), 'E', -1, 64))); err != nil {
					return err
				}
				if _, err = to.Write([]byte("\n")); err != nil {
					return err
				}
			}
		}
		return nil
	}
	var largeSampleRttHead = func(to io.Writer) error {
		_, err := to.Write([]byte("Location,Test,timestamp,Deployment,ConnID,RttID,Rtt\n"))
		return err
	}
	tables = append(tables, NewTable("largeSample_rtt", largeSampleRttMerge, largeSampleRttHead))

	var largeSampleMerge = func(location, test, timestamp, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.SplitAfter(string(data), "\n")
		line := lines[3]
		start := strings.Index(line, ":") + 1
		numValidRtt := strings.TrimSpace(line[start:])

		line = lines[4]
		start = strings.Index(line, ":") + 1
		line = line[start:]
		stop := strings.Index(line, "/")
		minRtt := strings.TrimSpace(line[:stop])
		du, err := time.ParseDuration(minRtt)
		if err != nil {
			return err
		}
		minRtt = strconv.FormatFloat(float64(du), 'E', -1, 64)

		start = stop + 1
		line = line[start:]
		stop = strings.Index(line, "/")
		avgRtt := strings.TrimSpace(line[:stop])
		du, err = time.ParseDuration(avgRtt)
		if err != nil {
			return err
		}
		avgRtt = strconv.FormatFloat(float64(du), 'E', -1, 64)

		start = stop + 1
		line = line[start:]
		stop = strings.Index(line, "/")
		maxRtt := strings.TrimSpace(line[:stop])
		du, err = time.ParseDuration(maxRtt)
		if err != nil {
			return err
		}
		maxRtt = strconv.FormatFloat(float64(du), 'E', -1, 64)

		start = stop + 1
		line = line[start:]
		stdRtt := strings.TrimSpace(line)
		du, err = time.ParseDuration(stdRtt)
		if err != nil {
			return err
		}
		stdRtt = strconv.FormatFloat(float64(du), 'E', -1, 64)

		line = lines[5]
		start = strings.Index(line, ":") + 1
		tps := strings.TrimSpace(line[start:])

		if _, err = to.Write([]byte(location + "," + test + "," + timestamp + "," + deployment)); err != nil {
			return err
		}
		if _, err = to.Write([]byte("," + numValidRtt)); err != nil {
			return err
		}
		if _, err = to.Write([]byte("," + minRtt)); err != nil {
			return err
		}
		if _, err = to.Write([]byte("," + avgRtt)); err != nil {
			return err
		}
		if _, err = to.Write([]byte("," + maxRtt)); err != nil {
			return err
		}
		if _, err = to.Write([]byte("," + stdRtt)); err != nil {
			return err
		}
		if _, err = to.Write([]byte("," + tps)); err != nil {
			return err
		}
		if _, err = to.Write([]byte("\n")); err != nil {
			return err
		}
		return nil
	}
	var largeSampleHead = func(to io.Writer) error {
		_, err := to.Write([]byte("Location,Test,timestamp,Deployment,NumValidRtt,MinRtt,AvgRtt,MaxRtt,StdRtt,Tps\n"))
		return err
	}
	tables = append(tables, NewTable("largeSample", largeSampleMerge, largeSampleHead))
	var largeSampleConnMerge = func(location, test, timestamp, deployment, path string, to io.Writer) error {
		return nil //do nothing
	}
	var largeSampleConnHead = func(to io.Writer) error {
		_, err := to.Write([]byte("\n"))
		return err
	}
	tables = append(tables, NewTable("largeSample_conn", largeSampleConnMerge, largeSampleConnHead))

	var waitResponseRttMerge = largeSampleRttMerge
	var waitResponseRttHead = largeSampleRttHead
	tables = append(tables, NewTable("waitResponse_rtt", waitResponseRttMerge, waitResponseRttHead))

	var waitResponseConnMerge = largeSampleConnMerge
	var waitResponseConnHead = largeSampleConnHead
	tables = append(tables, NewTable("waitResponse_conn", waitResponseConnMerge, waitResponseConnHead))

	var waitResponseMerge = largeSampleMerge
	var waitResponseHead = largeSampleHead
	tables = append(tables, NewTable("waitResponse", waitResponseMerge, waitResponseHead))

	deployments := []string{
		"physical", "lxcNetworkDefault", "lxcBridgeBr0", "lxcBridgeOvsbr0",
		"kvmNetworkDefault", "kvmBridgeBr0", "kvmBridgeOvsbr0", "kvmRtl8139NetworkDefault",
		"osvBridgeBr0",
	}

	tests := []string{
		"basic", "ovs", "ovs2.5.0", "rtl8139", "unikernel",
	}

	locations := []string{
		"Local", "Remote", "10gRemote",
	}

	resultsDir := "results"
	mergedResultsDir := resultsDir + "/" + "merged"
	MergeResults(resultsDir, mergedResultsDir, locations, tests, deployments, tables)
	fmt.Println("-----------------finished-----------------")
	fmt.Println("Merged results stored in: " + mergedResultsDir)
}
