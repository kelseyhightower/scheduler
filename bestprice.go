package main

import (
	"fmt"
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
		fmt.Printf("%s [$%.2f]\n", n.Metadata.Name, f)
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
