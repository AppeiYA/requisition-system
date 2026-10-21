// Requisition System worker.
//
// The worker drains the asynq task queues backed by Redis. The API enqueues
// tasks (for example, sending an invitation email) and this process executes
// them, so integrations that do not need to be synchronous never add latency
// to a request.
//
//	go run ./cmd/worker
package main

import app "github.com/AppeiYA/requisition-system/internal"

func main() {
	err := app.NewWorker().Run()
	if err != nil {
		panic(err)
	}
}
