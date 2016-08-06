// Copyright 2016 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	apiHost            = "http://127.0.0.1:8001"
	eventsEndpoint     = "/api/v1/namespaces/default/events"
	nodesEndpoint      = "/api/v1/nodes"
	watchPodEndpoint   = "/api/v1/pods?watch=true&fieldSelector=spec.nodeName="
	podsEndpoint       = "/api/v1/pods?fieldSelector=spec.nodeName="
)

func createEvent(event Event) error {
	var b []byte
	body := bytes.NewBuffer(b)
	err := json.NewEncoder(body).Encode(event)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiHost+eventsEndpoint, "application/json", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		data, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		log.Println(string(data))
		return errors.New("Event: Unexpected HTTP status code" + resp.Status)
	}
	return nil

}

func getNodes() (*NodeList, error) {
	var nodeList NodeList
	resp, err := http.Get(apiHost + nodesEndpoint)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&nodeList)
	if err != nil {
		return nil, err
	}
	return &nodeList, nil
}

func monitorUnscheduledPods() (<-chan Pod, <-chan error) {
	pods := make(chan Pod)
	errc := make(chan error, 1)

	go func() {
		for {
			resp, err := http.Get(apiHost + watchPodEndpoint)
			if err != nil {
				errc <- err
				time.Sleep(5 * time.Second)
				continue
			}

			if resp.StatusCode != 200 {
				errc <- errors.New("Invalid status code: " + resp.Status)
				time.Sleep(5 * time.Second)
				continue
			}

			decoder := json.NewDecoder(resp.Body)
			for {
				var event PodWatchEvent
				err = decoder.Decode(&event)
				if err != nil {
					errc <- err
					break
				}

				if event.Type == "ADDED" {
					pods <- event.Object
				}
			}
		}
	}()

	return pods, errc
}

func getUnscheduledPods() ([]*Pod, error) {
	ups := make([]*Pod, 0)

	var podList PodList
	resp, err := http.Get(apiHost + podsEndpoint)
	if err != nil {
		return ups, err
	}
	err = json.NewDecoder(resp.Body).Decode(&podList)
	if err != nil {
		return ups, err
	}

	for _, pod := range podList.Items {
		if pod.Metadata.Annotations["scheduler.alpha.kubernetes.io/name"] == schedulerName {
			ups = append(ups, &pod)
		}
	}
	return ups, nil
}

func getRunningPods() (*PodList, error) {
	var podList PodList
	resp, err := http.Get("http://127.0.0.1:8001/api/v1/pods?fieldSelector=status.phase=Running")
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&podList)
	if err != nil {
		return nil, err
	}
	return &podList, nil
}

type ResourceUsage struct {
	CPU int
}

func fit(pod *Pod) ([]Node, error) {
	nodeList, err := getNodes()
	if err != nil {
		return nil, err
	}

	podList, err := getRunningPods()
	if err != nil {
		return nil, err
	}

	resourceUsage := make(map[string]*ResourceUsage)
	for _, node := range nodeList.Items {
		resourceUsage[node.Metadata.Name] = &ResourceUsage{}
	}

	for _, p := range podList.Items {
		for _, c := range p.Spec.Containers {
			if strings.HasSuffix(c.Resources.Requests["cpu"], "m") {
				milliCores := strings.TrimSuffix(c.Resources.Requests["cpu"], "m")
				cores, err := strconv.Atoi(milliCores)
				if err != nil {
					return nil, err
				}
				ru := resourceUsage[p.Spec.NodeName]
				ru.CPU += cores
			}
		}
	}

	var nodes []Node
	fitFailures := make([]string, 0)

	var spaceRequired int
	for _, c := range pod.Spec.Containers {
		if strings.HasSuffix(c.Resources.Requests["cpu"], "m") {
			milliCores := strings.TrimSuffix(c.Resources.Requests["cpu"], "m")
			cores, err := strconv.Atoi(milliCores)
			if err != nil {
				return nil, err
			}
			spaceRequired += cores
		}
	}

	for _, node := range nodeList.Items {
		cpu := node.Status.Allocatable["cpu"]
		cpuFloat, err := strconv.ParseFloat(cpu, 32)
		if err != nil {
			return nil, err
		}

		freeSpace := (int(cpuFloat*1000) - resourceUsage[node.Metadata.Name].CPU)
		if freeSpace < spaceRequired {
			m := fmt.Sprintf("fit failure on node (%s): Insufficient CPU", node.Metadata.Name)
			fitFailures = append(fitFailures, m)
			continue
		}
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		// Emit a Kubernetes event that the Pod was scheduled successfully.
		timestamp := time.Now().UTC().Format(time.RFC3339)
		event := Event{
			Count:          1,
			Message:        fmt.Sprintf("pod (%s) failed to fit on any node\n%s", pod.Metadata.Name, strings.Join(fitFailures, "\n")),
			Metadata:       Metadata{GenerateName: pod.Metadata.Name + "-"},
			Reason:         "FailedScheduling",
			LastTimestamp:  timestamp,
			FirstTimestamp: timestamp,
			Type:           "Warning",
			Source:         EventSource{Component: "hightower-scheduler"},
			InvolvedObject: ObjectReference{
				Kind:      "Pod",
				Name:      pod.Metadata.Name,
				Namespace: "default",
				Uid:       pod.Metadata.Uid,
			},
		}
		createEvent(event)
		return nodes, errors.New("no fit")
	}

	return nodes, nil
}

func bind(pod *Pod, node Node) error {
	binding := Binding{
		ApiVersion: "v1",
		Kind:       "Binding",
		Metadata:   Metadata{Name: pod.Metadata.Name},
		Target: Target{
			ApiVersion: "v1",
			Kind:       "Node",
			Name:       node.Metadata.Name,
		},
	}

	var b []byte
	body := bytes.NewBuffer(b)
	err := json.NewEncoder(body).Encode(binding)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/v1/namespaces/default/pods/%s/binding/", pod.Metadata.Name)
	resp, err := http.Post(apiHost+path, "application/json", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errors.New("Binding: Unexpected HTTP status code" + resp.Status)
	}

	// Emit a Kubernetes event that the Pod was scheduled successfully.
	timestamp := time.Now().UTC().Format(time.RFC3339)
	event := Event{
		Count:          1,
		Message:        fmt.Sprintf("Successfully assigned %s to %s", pod.Metadata.Name, node.Metadata.Name),
		Metadata:       Metadata{GenerateName: pod.Metadata.Name + "-"},
		Reason:         "Scheduled",
		LastTimestamp:  timestamp,
		FirstTimestamp: timestamp,
		Type:           "Normal",
		Source:         EventSource{Component: "hightower-scheduler"},
		InvolvedObject: ObjectReference{
			Kind:      "Pod",
			Name:      pod.Metadata.Name,
			Namespace: "default",
			Uid:       pod.Metadata.Uid,
		},
	}
	return createEvent(event)
}
