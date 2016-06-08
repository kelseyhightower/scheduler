package main

import (
	"log"
	"os"
)

const schedulerName = "hightower"

func main() {
	log.Println("Starting custom scheduler...")

	pod, err := getUnscheduledPod()
	if err != nil {
		log.Fatal(err)
	}

	if pod == nil {
		log.Println("No pods to schedule.")
		os.Exit(0)
	}

	nodes, err := fit(pod)
	if err != nil {
		log.Fatal(err)
	}
	node, err := bestPrice(nodes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(node.Metadata.Name)
	err = bind(pod, node)
	if err != nil {
		log.Fatal(err)
	}
}
