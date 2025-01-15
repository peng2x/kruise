//go:build windows
// +build windows

/*
Copyright 2021 The Kruise Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package criruntime

import (
	"fmt"
	"os"

	"k8s.io/klog/v2"
)

const (
	// CRISocketFilePipePath is the prefix of the CRI socket file path 
	criSocketFilePipePath = "npipe:////./pipe/"

	// CRISocketContainerd is the containerd CRI endpoint
	criSocketContainerd = "npipe:////./pipe/containerd-containerd"

	// DefaultCRISocket defines the default CRI socket
	defaultCRISocket = criSocketContainerd
)

func detectRuntime(varRunPath string) (cfgs []runtimeConfig) {
	var err error

	// firstly check if it is configured from flag
	if CRISocketFileName != nil && len(*CRISocketFileName) > 0 {
		filePath := fmt.Sprintf("%s/%s", criSocketFilePipePath, *CRISocketFileName)
		if _, err = os.Stat(filePath); err == nil {
			cfgs = append(cfgs, runtimeConfig{
				runtimeType:      ContainerRuntimeCommonCRI,
				runtimeRemoteURI: fmt.Sprintf("unix://%s/%s", criSocketFilePipePath, *CRISocketFileName),
			})
			klog.InfoS("Find configured CRI socket with given flag", "filePath", filePath)
		} else {
			klog.ErrorS(err, "Failed to stat the CRI socket with given flag", "filePath", filePath)
		}
		return
	}

	// if the flag is not set, then use the default CRI socket

	cfgs = append(cfgs, runtimeConfig{
		runtimeType:      ContainerRuntimeContainerd,
		runtimeRemoteURI: defaultCRISocket,
	})
	return cfgs
}
