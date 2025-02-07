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
	"context"
	"fmt"
	"os"
	"time"

	oteltrace "go.opentelemetry.io/otel/trace"
	criapi "k8s.io/cri-api/pkg/apis"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/klog/v2"
	criremote "k8s.io/kubernetes/pkg/kubelet/cri/remote"

	runtimeimage "github.com/openkruise/kruise/pkg/daemon/criruntime/imageruntime"
	daemonutil "github.com/openkruise/kruise/pkg/daemon/util"
)

const (
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants_windows.go
	// criSocketFilePipePath is the prefix of the CRI socket file path
	criSocketFilePipePath = "npipe://./pipe/"

	// CRISocketContainerd is the containerd CRI endpoint
	criSocketContainerd = `npipe:\\.\pipe\containerd-containerd`

	// DefaultCRISocket defines the default CRI socket
	defaultCRISocket = criSocketContainerd
)

func NewFactory(varRunPath string, accountManager daemonutil.ImagePullAccountManager) (Factory, error) {
	cfgs := detectRuntime()
	if len(cfgs) == 0 {
		return nil, fmt.Errorf("not found container runtime sock")
	}

	var err error
	f := &factory{}

	var cfg runtimeConfig
	for i := range cfgs {
		cfg = cfgs[i]
		var imageService runtimeimage.ImageService
		var runtimeService criapi.RuntimeService
		var typedVersion *runtimeapi.VersionResponse

		imageService, err = runtimeimage.NewCRIImageService(cfg.runtimeRemoteURI, accountManager)
		if err != nil {
			klog.ErrorS(err, "Failed to new image service", "runtimeType", cfg.runtimeType, "runtimeURI", cfg.runtimeURI, "runtimeRemoteURI", cfg.runtimeRemoteURI)
			continue
		}

		if _, err = imageService.ListImages(context.TODO()); err != nil {
			klog.ErrorS(err, "Failed to list images", "runtimeType", cfg.runtimeType, "runtimeURI", cfg.runtimeURI, "runtimeRemoteURI", cfg.runtimeRemoteURI)
			continue
		}

		runtimeService, err = criremote.NewRemoteRuntimeService(cfg.runtimeRemoteURI, time.Second*5, oteltrace.NewNoopTracerProvider())
		if err != nil {
			klog.ErrorS(err, "Failed to new runtime service", "runtimeType", cfg.runtimeType, "runtimeURI", cfg.runtimeURI, "runtimeRemoteURI", cfg.runtimeRemoteURI)
			continue
		}
		typedVersion, err = runtimeService.Version(context.TODO(), kubeRuntimeAPIVersion)
		if err != nil {
			klog.ErrorS(err, "Failed to get runtime typed version", "runtimeType", cfg.runtimeType, "runtimeURI", cfg.runtimeURI, "runtimeRemoteURI", cfg.runtimeRemoteURI)
			continue
		}

		klog.V(2).InfoS("Add runtime", "runtimeName", typedVersion.RuntimeName, "runtimeURI", cfg.runtimeURI, "runtimeRemoteURI", cfg.runtimeRemoteURI)
		f.impls = append(f.impls, &runtimeImpl{
			cfg:            cfg,
			runtimeName:    typedVersion.RuntimeName,
			imageService:   imageService,
			runtimeService: runtimeService,
		})
	}
	if len(f.impls) == 0 {
		return nil, err
	}

	return f, nil
}

func detectRuntime() (cfgs []runtimeConfig) {
	addr := os.Getenv("CONTAINERD_ADDRESS")
	if addr == "" {
		addr = defaultCRISocket
	}
	cfgs = append(cfgs, runtimeConfig{
		runtimeType:      ContainerRuntimeContainerd,
		runtimeRemoteURI: addr,
	})
	return cfgs
}
