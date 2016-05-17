// resultParser project main.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Table struct {
	Name      string
	MergeFunc func(location, test, time, deployment, path string, to io.Writer) error
	fp        *os.File
	bw        *bufio.Writer
	lock      sync.Mutex
}

func (t *Table) Merge(location, test, time, deployment, path string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.MergeFunc(location, test, time, deployment, path, t.bw)
}

func NewTable(
	name string,
	mergeFunc func(location, test, time, deployment, path string, to io.Writer) error,
) *Table {
	return &Table{
		Name:      name,
		MergeFunc: mergeFunc,
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

	var netperfMerge = func(location, test, time, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(string(data), "\n")
		words := strings.Fields(lines[6])
		tps := words[5]
		_, err = to.Write([]byte(location + "," + test + "," + time + "," + deployment + "," + tps + "\n"))
		return err
	}
	tables = append(tables, NewTable("netperf", netperfMerge))

	var iperf3Merge = func(location, test, time, deployment, path string, to io.Writer) error {
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
		_, err = to.Write([]byte(location + "," + test + "," + time + "," + deployment + "," + bandwidth + " " + unit + "," + retry + "\n"))
		return err
	}
	tables = append(tables, NewTable("iperf3", iperf3Merge))

	var redisMerge = func(location, test, time, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.SplitAfter(string(data), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				if _, err = to.Write([]byte(location + "," + test + "," + time + "," + deployment + ",")); err != nil {
					return err
				}
				if _, err = to.Write([]byte(line)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	tables = append(tables, NewTable("redis", redisMerge))

	var changeRequestSizeMerge = func(location, test, time, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.SplitAfter(string(data), "\n")
		for _, line := range lines {
			if trimed := strings.TrimSpace(line); trimed != "" {
				words := strings.Fields(trimed)
				if _, err = to.Write([]byte(location + "," + test + "," + time + "," + deployment)); err != nil {
					return err
				}
				for _, word := range words {
					if _, err = to.Write([]byte("," + word)); err != nil {
						return err
					}
				}
				if _, err = to.Write([]byte("\n")); err != nil {
					return err
				}
			}
		}
		return nil
	}
	tables = append(tables, NewTable("changeRequestSize", changeRequestSizeMerge))

	var changeRequestPeriodMerge = changeRequestSizeMerge
	tables = append(tables, NewTable("changeRequestPeriod", changeRequestPeriodMerge))

	var largeSampleRttMerge = func(location, test, time, deployment, path string, to io.Writer) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.SplitAfter(string(data), "\n")
		lines = lines[1:] //ignore the first line of table head

		rttNumLimit := 10000
		for key, line := range lines {
			if !(key < rttNumLimit) {
				break
			}
			if trimed := strings.TrimSpace(line); trimed != "" {
				words := strings.Fields(trimed)
				if _, err = to.Write([]byte(location + "," + test + "," + time + "," + deployment)); err != nil {
					return err
				}
				for _, word := range words {
					if _, err = to.Write([]byte("," + word)); err != nil {
						return err
					}
				}
				if _, err = to.Write([]byte("\n")); err != nil {
					return err
				}
			}
		}
		return nil
	}
	tables = append(tables, NewTable("largeSample_rtt", largeSampleRttMerge))

	var largeSampleMerge = func(location, test, time, deployment, path string, to io.Writer) error {
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
		start = stop + 1
		line = line[start:]
		stop = strings.Index(line, "/")
		avgRtt := strings.TrimSpace(line[:stop])
		start = stop + 1
		line = line[start:]
		stop = strings.Index(line, "/")
		maxRtt := strings.TrimSpace(line[:stop])
		start = stop + 1
		line = line[start:]
		stdRtt := strings.TrimSpace(line)

		line = lines[5]
		start = strings.Index(line, ":") + 1
		tps := strings.TrimSpace(line[start:])

		if _, err = to.Write([]byte(location + "," + test + "," + time + "," + deployment)); err != nil {
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
	tables = append(tables, NewTable("largeSample", largeSampleMerge))

	tables = append(tables, NewTable("largeSample_conn", func(location, test, time, deployment, path string, to io.Writer) error { return nil }))

	var waitResponseRttMerge = largeSampleRttMerge
	tables = append(tables, NewTable("waitResponse_rtt", waitResponseRttMerge))

	tables = append(tables, NewTable("waitResponse_conn", func(location, test, time, deployment, path string, to io.Writer) error { return nil }))

	var waitResponseMerge = largeSampleMerge
	tables = append(tables, NewTable("waitResponse", waitResponseMerge))

	deployments := []string{
		"physical", "lxcNetworkDefault", "lxcBridgeBr0", "lxcBridgeOvsbr0",
		"kvmNetworkDefault", "kvmBridgeBr0", "kvmBridgeOvsbr0", "kvmRtl8139NetworkDefault",
		"osvBridgeBr0",
	}

	tests := []string{
		"basic", "ovs", "ovs2.5.0", "rtl8139", "unikernel",
	}

	locations := []string{
		"Local", "Remote",
	}

	resultsDir := "results"
	mergedResultsDir := resultsDir + "/" + "merged"
	MergeResults(resultsDir, mergedResultsDir, locations, tests, deployments, tables)
	fmt.Println("-----------------finished-----------------")
	fmt.Println("Merged results stored in: " + mergedResultsDir)
}
