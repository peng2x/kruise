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
	"flag"

	criapi "k8s.io/cri-api/pkg/apis"

	runtimeimage "github.com/openkruise/kruise/pkg/daemon/criruntime/imageruntime"
)

const (
	kubeRuntimeAPIVersion = "0.1.0"
)

var (
	CRISocketFileName = flag.String("socket-file", "", "The name of CRI socket file, and it should be in the mounted /hostvarrun directory.")
)

// Factory is the interface to get container and image runtime service
type Factory interface {
	GetImageService() runtimeimage.ImageService
	GetRuntimeService() criapi.RuntimeService
	GetRuntimeServiceByName(runtimeName string) criapi.RuntimeService
}

type ContainerRuntimeType string

const (
	ContainerRuntimeContainerd = "containerd"
	ContainerRuntimeCommonCRI  = "common-cri"
)

type runtimeConfig struct {
	runtimeType      ContainerRuntimeType
	runtimeURI       string
	runtimeRemoteURI string
}

type factory struct {
	impls []*runtimeImpl
}

type runtimeImpl struct {
	cfg            runtimeConfig
	runtimeName    string
	imageService   runtimeimage.ImageService
	runtimeService criapi.RuntimeService
}

func (f *factory) GetImageService() runtimeimage.ImageService {
	return f.impls[0].imageService
}

func (f *factory) GetRuntimeService() criapi.RuntimeService {
	return f.impls[0].runtimeService
}

func (f *factory) GetRuntimeServiceByName(runtimeName string) criapi.RuntimeService {
	for _, impl := range f.impls {
		if impl.runtimeName == runtimeName {
			return impl.runtimeService
		}
	}
	return nil
}
