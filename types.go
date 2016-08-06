// Copyright 2016 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

type PodList struct {
	ApiVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   ListMetadata `json:"metadata"`
	Items      []Pod        `json:"items"`
}

type PodEvent struct {
	Type   string `json:"type"`
	Object Pod    `json:"object"`
}

type Pod struct {
	Metadata Metadata `json:"metadata"`
	Spec     PodSpec  `json:"spec"`
}

type PodSpec struct {
	NodeName   string      `json:"nodeName"`
	Containers []Container `json:"containers"`
}

type Container struct {
	Name      string               `json:"name"`
	Resources ResourceRequirements `json:"resources"`
}

type ResourceRequirements struct {
	Limits   ResourceList `json:"limits"`
	Requests ResourceList `json:"requests"`
}

type ResourceList map[string]string

type Binding struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Target     Target   `json:"target"`
	Metadata   Metadata `json:"metadata"`
}

type Target struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type NodeList struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Items      []Node
}

type Node struct {
	Metadata Metadata   `json:"metadata"`
	Status   NodeStatus `json:"status"`
}

type NodeStatus struct {
	Capacity    ResourceList `json:"capacity"`
	Allocatable ResourceList `json:"allocatable"`
}

type ListMetadata struct {
	ResourceVersion string `json:"resourceVersion"`
}

type Metadata struct {
	Name            string            `json:"name"`
	ResourceVersion string            `json:"resourceVersion"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
}
