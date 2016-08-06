package main

import (
	"fmt"
	"log"
)

const schedulerName = "hightower"

func main() {
	log.Println("Starting custom scheduler...")
	pods, errc := monitorUnscheduledPods()

	for {
		select {
		case err := <-errc:
			log.Println(err)
		case pod := <-pods:
			nodes, err := fit(pod)
			if err != nil {
				log.Println(err)
				continue
			}
			node, err := bestPrice(nodes)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Printf("Assigned to: %s\n", node.Metadata.Name)
			err = bind(pod, node)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
