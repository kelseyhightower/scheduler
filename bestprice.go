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
