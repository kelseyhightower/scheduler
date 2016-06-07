package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

func getUnscheduledPod() (Pod, error) {
	var unscheduledPod Pod
	var podList PodList
	resp, err := http.Get("http://127.0.0.1:8001/api/v1/pods?fieldSelector=spec.nodeName=")
	if err != nil {
		return unscheduledPod, err
	}
	err = json.NewDecoder(resp.Body).Decode(&podList)
	if err != nil {
		return unscheduledPod, err
	}

	for _, pod := range podList.Items {
		if pod.Metadata.Annotations["scheduler.alpha.kubernetes.io/name"] == schedulerName {
			unscheduledPod = pod
			break
		}
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
		fmt.Println(node.Metadata.Name)
		cpu := node.Status.Allocatable["cpu"]
		cpuFloat, err := strconv.ParseFloat(cpu, 32)
		if err != nil {
			return nil, err
		}

		freeSpace := (int(cpuFloat * 1000) - resourceUsage[node.Metadata.Name].CPU)
		if freeSpace > spaceRequired {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}
