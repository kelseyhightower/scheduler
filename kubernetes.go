package main

import (
	"encoding/json"
	"fmt"
	"math"
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
	CPU float64
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
				ru.CPU += (float64(cores) / 1000)
			}
		}
	}

	var nodes []Node

	var spaceRequired float64
	for _, c := range pod.Spec.Containers {
		if strings.HasSuffix(c.Resources.Requests["cpu"], "m") {
			milliCores := strings.TrimSuffix(c.Resources.Requests["cpu"], "m")
			cores, err := strconv.Atoi(milliCores)
			if err != nil {
				return nil, err
			}
			spaceRequired += (float64(cores) / 1000)
		}
	}

	for _, node := range nodeList.Items {
		fmt.Println(node.Metadata.Name)
		cpu := node.Status.Allocatable["cpu"]
		cpuInt, err := strconv.Atoi(cpu)
		if err != nil {
			return nil, err
		}

		freeSpace := round((float64(cpuInt) - resourceUsage[node.Metadata.Name].CPU), 2)
		if freeSpace > spaceRequired {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func round(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
}
