// MIT License
//
// Copyright (c) 2020 Dmitrii Ustiugov, Plamen Petrov
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
	"os"
	"strconv"
	"sync"
	"testing"

	ctrdlog "github.com/containerd/containerd/log"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	ctriface "github.com/ustiugov/fccd-orchestrator/ctriface"
)

const (
	isTestModeConst   = true
	isSaveMemoryConst = true
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: ctrdlog.RFC3339NanoFixed,
		FullTimestamp:   true,
	})
	//log.SetReportCaller(true) // FIXME: make sure it's false unless debugging

	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)

	var envSetting = os.Getenv("GOORCHSNAPSHOTS")
	if envSetting == "" {
		envSetting = "false"
	}

	log.Infof("Orchestrator snapshots enabled: %s", envSetting)

	envBool, err := strconv.ParseBool(envSetting)
	if err != nil {
		log.Panic("Invalid snapshots mode environment variable")
	}

	orch = ctriface.NewOrchestrator("devmapper", ctriface.WithTestModeOn(true), ctriface.WithSnapshots(envBool))

	ret := m.Run()

	err = orch.StopActiveVMs()
	if err != nil {
		log.Printf("Failed to stop VMs, err: %v\n", err)
	}

	orch.Cleanup()

	os.Exit(ret)
}

func TestSendToFunctionSerial(t *testing.T) {
	fID := "1"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64
		pinnedFuncNum int
	)
	funcPool = NewFuncPool(!isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	for i := 0; i < 2; i++ {
		resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
		require.NoError(t, err, "Function returned error")
		if i == 0 {
			require.Equal(t, resp.IsColdStart, true)
		}

		require.Equal(t, resp.Payload, "Hello, world!")
	}

	message, err := funcPool.RemoveInstance(fID, imageName, true)
	require.NoError(t, err, "Function returned error, "+message)
}

func TestSendToFunctionParallel(t *testing.T) {
	fID := "2"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64
		pinnedFuncNum int
	)
	funcPool = NewFuncPool(!isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	var vmGroup sync.WaitGroup
	for i := 0; i < 100; i++ {
		vmGroup.Add(1)

		go func(i int) {
			defer vmGroup.Done()
			resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
			require.NoError(t, err, "Function returned error")
			require.Equal(t, resp.Payload, "Hello, world!")
		}(i)

	}
	vmGroup.Wait()

	message, err := funcPool.RemoveInstance(fID, imageName, true)
	require.NoError(t, err, "Function returned error, "+message)
}

func TestStartSendStopTwice(t *testing.T) {
	fID := "3"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64 = 1
		pinnedFuncNum int    = 2
	)
	funcPool = NewFuncPool(!isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	for i := 0; i < 2; i++ {
		for k := 0; k < 2; k++ {
			resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
			require.NoError(t, err, "Function returned error")
			require.Equal(t, resp.Payload, "Hello, world!")
		}

		message, err := funcPool.RemoveInstance(fID, imageName, true)
		require.NoError(t, err, "Function returned error, "+message)
	}

	servedGot := funcPool.stats.statMap[fID].served
	require.Equal(t, 4, int(servedGot), "Cold start (served) stats are wrong")
	startsGot := funcPool.stats.statMap[fID].started
	require.Equal(t, 2, int(startsGot), "Cold start (starts) stats are wrong")
}

func TestStatsNotNumericFunction(t *testing.T) {
	fID := "not_cld"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64 = 1
		pinnedFuncNum int    = 2
	)
	funcPool = NewFuncPool(isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
	require.NoError(t, err, "Function returned error")
	require.Equal(t, resp.Payload, "Hello, world!")

	message, err := funcPool.RemoveInstance(fID, imageName, true)
	require.NoError(t, err, "Function returned error, "+message)

	servedGot := funcPool.stats.statMap[fID].served
	require.Equal(t, 1, int(servedGot), "Cold start (served) stats are wrong")
	startsGot := funcPool.stats.statMap[fID].started
	require.Equal(t, 1, int(startsGot), "Cold start (starts) stats are wrong")
}

func TestStatsNotColdFunction(t *testing.T) {
	fID := "4"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64 = 1
		pinnedFuncNum int    = 4
	)
	funcPool = NewFuncPool(isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
	require.NoError(t, err, "Function returned error")
	require.Equal(t, resp.Payload, "Hello, world!")

	message, err := funcPool.RemoveInstance(fID, imageName, true)
	require.NoError(t, err, "Function returned error, "+message)

	servedGot := funcPool.stats.statMap[fID].served
	require.Equal(t, 1, int(servedGot), "Cold start (served) stats are wrong")
	startsGot := funcPool.stats.statMap[fID].started
	require.Equal(t, 1, int(startsGot), "Cold start (starts) stats are wrong")
}

func TestSaveMemorySerial(t *testing.T) {
	fID := "5"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64 = 40
		pinnedFuncNum int    = 2
	)
	funcPool = NewFuncPool(isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	for i := 0; i < 100; i++ {
		resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
		require.NoError(t, err, "Function returned error")
		require.Equal(t, resp.Payload, "Hello, world!")
	}

	startsGot := funcPool.stats.statMap[fID].started
	require.Equal(t, 3, int(startsGot), "Cold start (starts) stats are wrong")

	message, err := funcPool.RemoveInstance(fID, imageName, true)
	require.NoError(t, err, "Function returned error, "+message)
}

func TestSaveMemoryParallel(t *testing.T) {
	fID := "6"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64 = 40
		pinnedFuncNum int    = 2
	)
	funcPool = NewFuncPool(isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	var vmGroup sync.WaitGroup
	for i := 0; i < 100; i++ {
		vmGroup.Add(1)

		go func(i int) {
			defer vmGroup.Done()

			resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
			require.NoError(t, err, "Function returned error")
			require.Equal(t, resp.Payload, "Hello, world!")
		}(i)

	}
	vmGroup.Wait()

	startsGot := funcPool.stats.statMap[fID].started
	require.Equal(t, 3, int(startsGot), "Cold start (starts) stats are wrong")

	message, err := funcPool.RemoveInstance(fID, imageName, true)
	require.NoError(t, err, "Function returned error, "+message)
}

func TestDirectStartStopVM(t *testing.T) {
	fID := "7"
	imageName := "ustiugov/helloworld:runner_workload"
	var (
		servedTh      uint64
		pinnedFuncNum int
	)
	funcPool = NewFuncPool(!isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	message, err := funcPool.AddInstance(fID, imageName)
	require.NoError(t, err, "This error should never happen (addInstance())"+message)

	resp, _, err := funcPool.Serve(context.Background(), fID, imageName, "world")
	require.NoError(t, err, "Function returned error")
	require.Equal(t, resp.Payload, "Hello, world!")

	message, err = funcPool.RemoveInstance(fID, imageName, true)
	require.NoError(t, err, "Function returned error, "+message)
}

func TestAllFunctions(t *testing.T) {
	images := []string{
		"ustiugov/helloworld:var_workload",
		"ustiugov/chameleon:var_workload",
		"ustiugov/pyaes:var_workload",
		"ustiugov/image_rotate:var_workload",
		"ustiugov/json_serdes:var_workload",
		"ustiugov/lr_serving:var_workload",
		"ustiugov/cnn_serving:var_workload",
		"ustiugov/rnn_serving:var_workload",
		"ustiugov/lr_training:var_workload",
	}
	var (
		servedTh      uint64
		pinnedFuncNum int
	)
	funcPool = NewFuncPool(!isSaveMemoryConst, servedTh, pinnedFuncNum, isTestModeConst)

	for i := 0; i < 2; i++ {
		var vmGroup sync.WaitGroup
		for fID, imageName := range images {
			reqs := []string{"world", "record", "replay"}
			resps := []string{"world", "record_response", "replay_response"}
			for k := 0; k < 3; k++ {
				vmGroup.Add(1)
				go func(fID int, imageName, request, response string) {
					defer vmGroup.Done()

					resp, _, err := funcPool.Serve(context.Background(), strconv.Itoa(8+fID), imageName, request)
					require.NoError(t, err, "Function returned error")

					require.Equal(t, resp.Payload, "Hello, "+response+"!")
				}(fID, imageName, reqs[k], resps[k])
			}
			vmGroup.Wait()
		}
	}

	var vmGroup sync.WaitGroup
	for fID, imageName := range images {
		vmGroup.Add(1)
		go func(fID int, imageName string) {
			defer vmGroup.Done()

			message, err := funcPool.RemoveInstance(strconv.Itoa(8+fID), imageName, true)
			require.NoError(t, err, "Function returned error, "+message)
		}(fID, imageName)
	}
	vmGroup.Wait()
}
