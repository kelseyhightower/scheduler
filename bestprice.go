package main

import (
	"log"
	"strconv"
)

func bestPrice(nodes []Node) (Node, error) {
	type NodePrice struct {
		Node  Node
		Price float64
	}

	var bestNodePrice *NodePrice
	for _, n := range nodes {
		price, ok := n.Metadata.Annotations["hightower.com/cost"]
		if !ok {
			continue
		}
		f, err := strconv.ParseFloat(price, 32)
		if err != nil {
			return Node{}, err
		}
		log.Printf("Price %v Node %v", f, n.Metadata.Name)
		if bestNodePrice == nil {
			bestNodePrice = &NodePrice{n, f}
			continue
		}
		if f < bestNodePrice.Price {
			bestNodePrice.Node = n
			bestNodePrice.Price = f
		}
	}

	if bestNodePrice == nil {
		bestNodePrice = &NodePrice{nodes[0], 0}
	}
	return bestNodePrice.Node, nil
}
