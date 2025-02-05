//go:build !windows
// +build !windows

/*
Copyright 2022 The Kruise Authors.
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

package imageruntime

import (
	"context"
	"time"

	daemonutil "github.com/openkruise/kruise/pkg/daemon/util"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/util"
)

// NewCRIImageService create a common CRI runtime
func NewCRIImageService(runtimeURI string, accountManager daemonutil.ImagePullAccountManager) (ImageService, error) {
	klog.V(3).InfoS("Connecting to image service", "endpoint", runtimeURI)
	addr, dialer, err := util.GetAddressAndDialer(runtimeURI)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithContextDialer(dialer), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)))
	if err != nil {
		klog.ErrorS(err, "Connect remote image service failed", "address", addr)
		return nil, err
	}

	imageClientV1, err := determineImageClientAPIVersion(conn)
	if err != nil {
		klog.ErrorS(err, "Failed to determine CRI image API version")
		return nil, err
	}

	return &commonCRIImageService{
		accountManager: accountManager,
		criImageClient: imageClientV1,
	}, nil
}
