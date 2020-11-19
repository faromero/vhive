// MIT License
//
// Copyright (c) 2020 Plamen Petrov
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

package cri

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	hpb "github.com/ustiugov/fccd-orchestrator/helloworld"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	coord         *coordinator
	gatewayURL    = flag.String("gatewayURL", "192.168.1.240.xip.io", "URL of the gateway")
	namespaceName = flag.String("namespace", "default", "name of namespace in which services exists")
)

func TestMain(m *testing.M) {
	coord = newCoordinator(nil, withoutOrchestrator())

	flag.Parse()

	ret := m.Run()
	os.Exit(ret)
}

func TestSingleInvoke(t *testing.T) {
	functionURL := getFuncURL("helloworld")
	invoke(t, functionURL)
}

func TestParallelInvoke(t *testing.T) {
	functionURL := getFuncURL("helloworld")
	n := 50
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			invoke(t, functionURL)
		}()
	}
	wg.Wait()
}

func TestMultipleFuncInvoke(t *testing.T) {
	var wg sync.WaitGroup
	funcs := []string{
		"helloworld",
		"pyaes",
		"rnnserving",
	}

	for _, funcName := range funcs {
		wg.Add(1)
		functionURL := getFuncURL(funcName)

		go func(functionURL string) {
			defer wg.Done()
			invoke(t, functionURL)
		}(functionURL)
	}

	wg.Wait()
}

func TestBench(t *testing.T) {
	start := time.Now()
	functionURL := getFuncURL("helloworld")

	dropPageCache()
	invoke(t, functionURL)
	end := time.Since(start)
	fmt.Printf("First invocation took %d ms\n", end.Milliseconds())

	dropPageCache()
	start = time.Now()
	invoke(t, functionURL)
	end = time.Since(start)
	fmt.Printf("Second invocation took %d ms\n", end.Milliseconds())
}

func invoke(t *testing.T, functionURL string) {
	reqPayload := "record"
	respPayload := "Hello, " + reqPayload + "_response!"

	client, conn, err := getClient(functionURL)
	require.NoError(t, err, "Failed to dial function URL")
	defer conn.Close()
	ctxFwd, cancel := context.WithDeadline(context.Background(), time.Now().Add(20*time.Second))
	defer cancel()

	resp, err := client.SayHello(ctxFwd, &hpb.HelloRequest{Name: reqPayload})
	require.NoError(t, err, "Failed to get response from function")
	require.Equal(t, respPayload, resp.Message, "Incorrect response payload")
}

func getClient(functionURL string) (hpb.GreeterClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(functionURL, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	return hpb.NewGreeterClient(conn), conn, nil
}

func dropPageCache() {
	cmd := exec.Command("sudo", "/bin/sh", "-c", "sync; echo 1 > /proc/sys/vm/drop_caches")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to drop caches: %v", err)
	}
}

func getFuncURL(funcName string) string {
	return funcName + "." + *namespaceName + "." + *gatewayURL + ":80"
}