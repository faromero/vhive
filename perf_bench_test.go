// MIT License
//
// Copyright (c) 2020 Yuchen Niu
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ease-lab/vhive/profile"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	vmNum           = flag.Int("vm", 2, "The number of VMs")
	targetReqPerSec = flag.Int("requestPerSec", 10, "The target number of requests per second")
	executionTime   = flag.Int("executionTime", 1, "The execution time of the benchmark in seconds")
	funcNames       = flag.String("funcNames", "helloworld", "Name of the functions to benchmark")

	perfExecTime = flag.Float64("perfExecTime", 0, "The execution time of perf command in seconds (run `perf stat` if bigger than 0)")
	perfInterval = flag.Int("perfInterval", 100, "Print count deltas every N milliseconds")
	perfEvents   = flag.String("perfEvents", "", "Perf events")
)

const (
	averageExecutionTime  = "average-execution-time"
	realRequestsPerSecond = "real-requests-per-second"
)

func TestBenchMultiVMRequestPerSecond(t *testing.T) {
	baseRPS := getCPUIntenseRPS()[*funcName]
	*funcNames = *funcName

	for i := 4; i <= 8; i += 4 {
		*vmNum = i
		*targetReqPerSec = *vmNum * baseRPS
		*funcNames = *funcName
		testName := fmt.Sprintf("%d VMs,%d RPS", *vmNum, *targetReqPerSec)
		t.Run(testName, TestBenchRequestPerSecond)
	}

	profile.CSVPlotter(getOutFile("benchRPS.csv"), *benchDir)
}

func TestBenchRequestPerSecond(t *testing.T) {
	var (
		servedTh      uint64
		pinnedFuncNum int
		vmID          int
		isSyncOffload bool = true
		isRunPerf          = *perfExecTime > 0
		images             = getImages(t)
		timeInterval       = time.Duration(time.Second.Nanoseconds() / int64(*targetReqPerSec))
		totalRequests      = *executionTime * *targetReqPerSec
		perfStat           = profile.NewPerfStat(*perfEvents, "perf-tmp.data", *perfExecTime, *perfInterval)
	)
	log.SetLevel(log.InfoLevel)

	checkInputValidation(t)

	createResultsDir()

	log.SetLevel(log.InfoLevel)
	bootStart := time.Now()

	funcPool = NewFuncPool(!isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	createResultsDir()

	// Pull images
	for _, imageName := range images {
		resp, _, err := funcPool.Serve(context.Background(), "plr_fnc", imageName, "record")
		require.NoError(t, err, "Function returned error")
		require.Equal(t, resp.Payload, "Hello, record_response!")
	}

	// Boot VMs
	for i := 0; i < *vmNum; i++ {
		vmIDString := strconv.Itoa(i)
		_, err := funcPool.AddInstance(vmIDString, images[i%len(images)])
		require.NoError(t, err, "Function returned error")
	}
	log.Debugf("All VMs booted in %d ms", time.Since(bootStart).Milliseconds())

	var (
		avgExecTime, realRPS int64
		vmGroup              sync.WaitGroup
		ticker               = time.NewTicker(timeInterval)
		tickerDone           = make(chan bool, 1)
		remainingRequests    = totalRequests
		warmupThreshold      = totalRequests - *vmNum - 1
	)

	if isRunPerf {
		err := perfStat.Run()
		require.NoError(t, err, "Start perf stat returned error")
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				vmGroup.Add(1)
				remainingRequests--
				if remainingRequests == warmupThreshold {
					perfStat.SetWarmTime()
				}

				imageName := images[vmID%len(images)]
				vmIDString := strconv.Itoa(vmID)

				go serveVM(t, vmIDString, imageName, &vmGroup, isSyncOffload, &avgExecTime, &realRPS)

				vmID = (vmID + 1) % *vmNum
			case <-tickerDone:
				ticker.Stop()
				vmGroup.Wait()
				return
			}
		}
	}()

	for remainingRequests > 0 {
	}
	tickerDone <- true
	vmGroup.Wait()
	perfStat.SetTearDownTime()

	// Shutdown VMs
	message, err := funcPool.RemoveInstance("plr_fnc", images[0], isSyncOffload)
	require.NoError(t, err, "Function returned error, "+message)
	for i := 0; i < *vmNum; i++ {
		vmIDString := strconv.Itoa(i)
		message, err := funcPool.RemoveInstance(vmIDString, images[i%len(images)], isSyncOffload)
		require.NoError(t, err, "Function returned error, "+message)
	}

	// Collect results
	serveMetrics := make(map[string]float64)
	serveMetrics[averageExecutionTime] = float64(avgExecTime) / float64(totalRequests)
	serveMetrics[realRequestsPerSecond] = float64(realRPS)

	if isRunPerf {
		result, err := perfStat.GetResult()
		require.NoError(t, err, "Stop perf stat returned error: %v", err)
		for eventName, value := range result {
			log.Debugf("%s: %f\n", eventName, value)
			serveMetrics[eventName] = value
		}
	}
	log.Debugf("average-execution-time: %f\n", serveMetrics[averageExecutionTime])
	log.Debugf("real-requests-per-second: %f\n", serveMetrics[realRequestsPerSecond])

	writeResultToCSV(t, "benchRPS.csv", serveMetrics)
}

func serveVM(t *testing.T, vmIDString, imageName string, vmGroup *sync.WaitGroup, isSyncOffload bool, avgExecTime, realRPS *int64) {
	defer vmGroup.Done()

	tStart := time.Now()
	resp, _, err := funcPool.Serve(context.Background(), vmIDString, imageName, "replay")
	require.Equal(t, resp.IsColdStart, false)
	if err != nil {
		log.Warn("Function returned error")
	} else {
		atomic.AddInt64(realRPS, 1)
	}
	if resp.Payload != "Hello, replay_response!" {
		log.Warnf("Function returned invalid: %s", resp.Payload)
	}

	execTime := time.Since(tStart).Milliseconds()
	atomic.AddInt64(avgExecTime, execTime)
	log.Debugf("VM %s: returned in %d milliseconds", vmIDString, execTime)
}

// Returns a list of image names
func getImages(t *testing.T) []string {
	var (
		images = map[string]string{
			"helloworld":   "ustiugov/helloworld:var_workload",
			"chameleon":    "ustiugov/chameleon:var_workload",
			"pyaes":        "ustiugov/pyaes:var_workload",
			"image_rotate": "ustiugov/image_rotate:var_workload",
			"json_serdes":  "ustiugov/json_serdes:var_workload",
			"lr_serving":   "ustiugov/lr_serving:var_workload",
			"cnn_serving":  "ustiugov/cnn_serving:var_workload",
			"rnn_serving":  "ustiugov/rnn_serving:var_workload",
		}
		funcs  = strings.Split(*funcNames, ",")
		result []string
	)

	for _, funcName := range funcs {
		imageName, isPresent := images[funcName]
		require.True(t, isPresent, "Function %s is not supported", funcName)
		result = append(result, imageName)
	}

	return result
}

func writeResultToCSV(t *testing.T, outfile string, metrics map[string]float64) {
	outFile := getOutFile(outfile)

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	require.NoError(t, err, "Failed opening file")
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()
	require.True(t, err == nil || err.Error() == "EOF", "Failed reading file")

	writer := csv.NewWriter(f)
	if headers == nil {
		for k := range metrics {
			headers = append(headers, k)
		}
		err = writer.Write(headers)
		require.NoError(t, err, "Failed writting file")
		writer.Flush()
	}

	var data []string
	for _, header := range headers {
		for k, v := range metrics {
			if k == header {
				vStr := strconv.FormatFloat(v, 'f', -1, 64)
				data = append(data, vStr)
			}
		}
	}
	err = writer.Write(data)
	require.NoError(t, err, "Failed writting file")

	writer.Flush()
}

func getCPUIntenseRPS() map[string]int {
	return map[string]int{
		"helloworld":   10000,
		"chameleon":    400,
		"pyaes":        2500,
		"image_rotate": 600,
		"json_serdes":  600,
		"lr_serving":   5000,
		"cnn_serving":  100,
		"rnn_serving":  600,
	}
}
