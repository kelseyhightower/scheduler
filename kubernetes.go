package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getNodes() (*NodeList, error) {
	var nodeList NodeList
	resp, err := http.Get("http://127.0.0.1:8001/api/v1/nodes")
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
			resp, err := http.Get("http://127.0.0.1:8001/api/v1/pods?watch=true&fieldSelector=spec.nodeName=")
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
				var event PodEvent
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

func getUnscheduledPod() (*Pod, error) {
	var unscheduledPod *Pod

	var podList PodList
	resp, err := http.Get("http://127.0.0.1:8001/api/v1/pods?fieldSelector=spec.nodeName=")
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&podList)
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		if pod.Metadata.Annotations["scheduler.alpha.kubernetes.io/name"] == schedulerName {
			unscheduledPod = &pod
			break
		}
	}
	if unscheduledPod == nil {
		return nil, nil
	}

	return unscheduledPod, nil
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

func fit(pod Pod) ([]Node, error) {
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
		if freeSpace > spaceRequired {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func bind(pod Pod, node Node) error {
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

	b := make([]byte, 0)
	body := bytes.NewBuffer(b)
	err := json.NewEncoder(body).Encode(binding)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/v1/namespaces/default/pods/%s/binding/", pod.Metadata.Name)
	resp, err := http.Post("http://127.0.0.1:8001"+path, "application/json", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errors.New("Binding: Unexpected HTTP status code" + resp.Status)
	}
	return nil
}
