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

import (
	"fmt"
	"log"
	"time"
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
			err = bind(pod, node)
			if err != nil {
				log.Println(err)
				continue
			}

			timestamp := time.Now().UTC().Format(time.RFC3339)
			event := Event{
				Count:          1,
				Message:        fmt.Sprintf("Successfully assigned %s to %s", pod.Metadata.Name, node.Metadata.Name),
				Reason:         "Scheduled",
				LastTimestamp:  timestamp,
				FirstTimestamp: timestamp,
				Type:           "Normal",
				Source:         EventSource{Component: "hightower-scheduler"},
				InvolvedObject: ObjectReference{
					Kind:      "Pod",
					Name:      pod.Metadata.Name,
					Namespace: "default",
				},
			}
			event.Metadata.GenerateName = pod.Metadata.Name + "-"

			err = createEvent(event)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
