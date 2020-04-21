// MIT License
//
// Copyright (c) 2020 Dmitrii Ustiugov
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
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	hpb "github.com/ustiugov/fccd-orchestrator/helloworld"
)

//////////////////////////////// FunctionPool type //////////////////////////////////////////

// FuncPool Pool of functions
type FuncPool struct {
	sync.Mutex
	funcMap map[string]*Function
}

// NewFuncPool Initializes a pool of functions. Functions can only be added
// but never removed from the map.
func NewFuncPool() *FuncPool {
	p := new(FuncPool)
	p.funcMap = make(map[string]*Function)

	return p
}

// GetFunction Returns a ptr to a function or creates it unless it exists
func (p *FuncPool) GetFunction(fID, imageName string) *Function {
	p.Lock()
	defer p.Unlock()

	_, found := p.funcMap[fID]
	if !found {
		p.funcMap[fID] = NewFunction(fID, imageName)
	}

	return p.funcMap[fID]
}

//////////////////////////////// Function type //////////////////////////////////////////////

// Function type
type Function struct {
	sync.RWMutex
	Once           *sync.Once
	fID            string
	imageName      string
	vmIDList       []string // FIXME: only a single VM per function is supported
	lastInstanceID int
	isActive       bool
}

// NewFunction Initializes a function
func NewFunction(fID, imageName string) *Function {
	f := new(Function)
	f.fID = fID
	f.imageName = imageName
	f.Once = new(sync.Once)

	return f
}

// IsActive Returns if function is active and ready to serve requests.
func (f *Function) IsActive() bool {
	return f.isActive
}

func (f *Function) getInstanceVMID() string {
	return f.vmIDList[0] // TODO: load balancing when many instances are supported
}

// FwdRPC Forward the RPC to an instance, then forwards the response back.
func (f *Function) FwdRPC(ctx context.Context, reqPayload string) (*hpb.HelloReply, error) {
	f.RLock()
	defer f.RUnlock()

	funcClientPtr, err := orch.GetFuncClient(f.getInstanceVMID())
	if err != nil {
		return &hpb.HelloReply{Message: "Failed to get function client"}, err
	}

	funcClient := *funcClientPtr

	resp, err := funcClient.SayHello(ctx, &hpb.HelloRequest{Name: reqPayload})

	return resp, err
}

// AddInstance Starts a VM, waits till it is ready.
func (f *Function) AddInstance() {
	f.Lock()
	defer f.Unlock()

	vmID := fmt.Sprintf("%s_%d", f.fID, f.lastInstanceID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	message, _, err := orch.StartVM(ctx, vmID, f.imageName)
	if err != nil {
		log.Panic(message, err)
	}

	f.vmIDList = append(f.vmIDList, vmID)
	f.lastInstanceID++

	f.isActive = true
}

// RemoveInstance Stops an instance (VM) of the function.
func (f *Function) RemoveInstance() (string, error) {
	f.Lock()
	var vmID string
	vmID, f.vmIDList = f.vmIDList[0], f.vmIDList[1:]

	if len(f.vmIDList) == 0 {
		f.isActive = false
		f.Once = new(sync.Once)
	} else {
		panic("List of function's instance is not empty after stopping an instance!")
	}

	f.Unlock()

	message, err := orch.StopSingleVM(context.Background(), vmID)
	if err != nil {
		log.Warn(message, err)
	}

	return message, err
}
