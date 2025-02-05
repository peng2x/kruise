//go:build windows
// +build windows

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

	daemonutil "github.com/openkruise/kruise/pkg/daemon/util"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

type unimplementedImageServiceClient struct {
}

func (c *unimplementedImageServiceClient) ListImages(ctx context.Context, in *runtimeapi.ListImagesRequest, opts ...grpc.CallOption) (*runtimeapi.ListImagesResponse, error) {
	return nil, nil
}

func (c *unimplementedImageServiceClient) ImageStatus(ctx context.Context, in *runtimeapi.ImageStatusRequest, opts ...grpc.CallOption) (*runtimeapi.ImageStatusResponse, error) {
	return nil, nil
}

func (c *unimplementedImageServiceClient) PullImage(ctx context.Context, in *runtimeapi.PullImageRequest, opts ...grpc.CallOption) (*runtimeapi.PullImageResponse, error) {
	return nil, nil
}

func (c *unimplementedImageServiceClient) RemoveImage(ctx context.Context, in *runtimeapi.RemoveImageRequest, opts ...grpc.CallOption) (*runtimeapi.RemoveImageResponse, error) {
	return nil, nil
}

func (c *unimplementedImageServiceClient) ImageFsInfo(ctx context.Context, in *runtimeapi.ImageFsInfoRequest, opts ...grpc.CallOption) (*runtimeapi.ImageFsInfoResponse, error) {
	return nil, nil
}

// NewCRIImageService create a common CRI runtime
func NewCRIImageService(runtimeURI string, accountManager daemonutil.ImagePullAccountManager) (ImageService, error) {
	klog.V(3).InfoS("Connecting to image service using dummy CRI Image Client", "endpoint", runtimeURI)

	return &commonCRIImageService{
		accountManager: accountManager,
		criImageClient: &unimplementedImageServiceClient{},
	}, nil
}
