//go:build !windows
// +build !windows

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

func detectRuntime(varRunPath string) (cfgs []runtimeConfig) {
	var err error

	// firstly check if it is configured from flag
	if CRISocketFileName != nil && len(*CRISocketFileName) > 0 {
		filePath := fmt.Sprintf("%s/%s", varRunPath, *CRISocketFileName)
		if _, err = os.Stat(filePath); err == nil {
			cfgs = append(cfgs, runtimeConfig{
				runtimeType:      ContainerRuntimeCommonCRI,
				runtimeRemoteURI: fmt.Sprintf("unix://%s/%s", varRunPath, *CRISocketFileName),
			})
			klog.InfoS("Find configured CRI socket with given flag", "filePath", filePath)
		} else {
			klog.ErrorS(err, "Failed to stat the CRI socket with given flag", "filePath", filePath)
		}
		return
	}

	// if the flag is not set, then try to find runtime in the recognized types and paths.

	// containerd, with the same behavior of pullImage as commonCRI
	{
		if _, err = os.Stat(fmt.Sprintf("%s/containerd.sock", varRunPath)); err == nil {
			cfgs = append(cfgs, runtimeConfig{
				runtimeType:      ContainerRuntimeContainerd,
				runtimeRemoteURI: fmt.Sprintf("unix://%s/containerd.sock", varRunPath),
			})
		}
		if _, err = os.Stat(fmt.Sprintf("%s/containerd/containerd.sock", varRunPath)); err == nil {
			cfgs = append(cfgs, runtimeConfig{
				runtimeType:      ContainerRuntimeContainerd,
				runtimeRemoteURI: fmt.Sprintf("unix://%s/containerd/containerd.sock", varRunPath),
			})
		}
	}

	// cri-o
	{
		if _, err = os.Stat(fmt.Sprintf("%s/crio.sock", varRunPath)); err == nil {
			cfgs = append(cfgs, runtimeConfig{
				runtimeType:      ContainerRuntimeCommonCRI,
				runtimeRemoteURI: fmt.Sprintf("unix://%s/crio.sock", varRunPath),
			})
		}
		if _, err = os.Stat(fmt.Sprintf("%s/crio/crio.sock", varRunPath)); err == nil {
			cfgs = append(cfgs, runtimeConfig{
				runtimeType:      ContainerRuntimeCommonCRI,
				runtimeRemoteURI: fmt.Sprintf("unix://%s/crio/crio.sock", varRunPath),
			})
		}
	}

	// cri-docker dockerd as a compliant Container Runtime Interface, detail see https://github.com/Mirantis/cri-dockerd
	{
		if _, err = os.Stat(fmt.Sprintf("%s/cri-dockerd.sock", varRunPath)); err == nil {
			cfgs = append(cfgs, runtimeConfig{
				runtimeType:      ContainerRuntimeCommonCRI,
				runtimeRemoteURI: fmt.Sprintf("unix://%s/cri-dockerd.sock", varRunPath),
			})
		}
	}
	return cfgs
}
