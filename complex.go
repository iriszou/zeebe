package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zeebe-io/zeebe/clients/go/entities"
	"github.com/zeebe-io/zeebe/clients/go/worker"
	"github.com/zeebe-io/zeebe/clients/go/zbc"
)

const brokerAddr string = "172.16.0.6:26500"

var zbClient zbc.ZBClient

func main() {

	var err error
	zbClient, err = zbc.NewZBClient(brokerAddr)
	if err != nil {
		panic(err)
	}

	// deploy workflow
	response, err := zbClient.NewDeployWorkflowCommand().AddResourceFile("complex.bpmn").Send()
	if err != nil {
		panic(err)
	}
	fmt.Println(response.String())

	go func() {
		cJobWorker := zbClient.NewJobWorker().JobType("create-service").Handler(handleCreateJob).Open()
		defer cJobWorker.Close()
		cJobWorker.AwaitClose()
	}()

	go func() {
		lowJobWorker := zbClient.NewJobWorker().JobType("low-service").Handler(handleLowJob).Open()
		defer lowJobWorker.Close()
		lowJobWorker.AwaitClose()
	}()

	go func() {
		highJobWorker := zbClient.NewJobWorker().JobType("high-service").Handler(handleHighJob).Open()
		defer highJobWorker.Close()
		highJobWorker.AwaitClose()
	}()

	go func() {
		replyJobWorker := zbClient.NewJobWorker().JobType("reply-service").Handler(handleReplyJob).Open()
		defer replyJobWorker.Close()
		replyJobWorker.AwaitClose()
	}()

	smux := http.NewServeMux()
	smux.HandleFunc("/new", HandleWorkflow)
	smux.HandleFunc("/message", HandleMessage)
	s := http.Server{
		Addr:    ":28000",
		Handler: smux,
	}
	s.ListenAndServe()

}

func HandleMessage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.Write([]byte(`read body err`))
		return
	}

	var reqBody map[string]interface{}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		w.Write([]byte(`json Unmarshal fail`))
		return
	}

	query := r.URL.Query()
	msgName := query.Get("name")
	msgKey := query.Get("key")

	payloadCmd, err := zbClient.NewPublishMessageCommand().MessageName(msgName).CorrelationKey(msgKey).PayloadFromMap(reqBody)
	if err != nil {
		fmt.Println("payloadCmd fail:", err)
		w.Write([]byte(`payload fail`))
		return
	}

	result, err := payloadCmd.Send()
	if err != nil {
		fmt.Println("request send fail", err)
		w.Write([]byte(`request.Send fail`))
		return
	}
	fmt.Println("send msg:", result.String())

	w.Write([]byte(`recv message and send`))
	return
}

func HandleWorkflow(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.Write([]byte(`read body err`))
		return
	}

	var reqBody map[string]interface{}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		w.Write([]byte(`json Unmarshal fail`))
		return
	}

	request, err := zbClient.NewCreateInstanceCommand().BPMNProcessId("complex").LatestVersion().PayloadFromMap(reqBody)
	if err != nil {
		fmt.Println("NewCreateInstanceCommand:", err)
		w.Write([]byte(`NewCreateInstanceCommand fail`))
		return
	}

	result, err := request.Send()
	if err != nil {
		fmt.Println("request send fail", err)
		w.Write([]byte("request.Send fail"))
		return
	}

	fmt.Println(result.String())
	w.Write([]byte(result.String()))
	return
}

func handleLowJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	ret := 0
	for {
		payload, err := job.GetPayloadAsMap()
		if err != nil {
			ret = 2
			break
		}

		// TODO sth
		payload["low"] = time.Now().Unix()
		// DONE sth
		request, err := zbClient.NewCompleteJobCommand().JobKey(jobKey).PayloadFromMap(payload)
		if err != nil {
			ret = 3
			break
		}

		request.Send()
		fmt.Println("Complete job", jobKey, "of type", job.Type)
		fmt.Println("Processing:", payload)
		break
	}

	if ret != 0 {
		fmt.Println("Failed to complete job", job.GetKey())
		client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send()
	}
	return
}

func handleHighJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	ret := 0
	for {
		payload, err := job.GetPayloadAsMap()
		if err != nil {
			ret = 2
			break
		}

		// TODO sth
		payload["high"] = time.Now().Unix()
		// DONE sth
		request, err := zbClient.NewCompleteJobCommand().JobKey(jobKey).PayloadFromMap(payload)
		if err != nil {
			ret = 3
			break
		}

		request.Send()
		fmt.Println("Complete job", jobKey, "of type", job.Type)
		fmt.Println("Processing:", payload)
		break
	}

	if ret != 0 {
		fmt.Println("Failed to complete job", job.GetKey())
		client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send()
	}
	return
}

func handleReplyJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	ret := 0
	for {
		payload, err := job.GetPayloadAsMap()
		if err != nil {
			ret = 2
			break
		}

		// TODO sth
		payload["reply"] = time.Now().Unix()
		// DONE sth
		request, err := zbClient.NewCompleteJobCommand().JobKey(jobKey).PayloadFromMap(payload)
		if err != nil {
			ret = 3
			break
		}

		request.Send()
		fmt.Println("Complete job", jobKey, "of type", job.Type)
		fmt.Println("Processing:", payload)
		break
	}

	if ret != 0 {
		fmt.Println("Failed to complete job", job.GetKey())
		client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send()
	}
	return
}

func handleCreateJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	ret := 0
	for {
		payload, err := job.GetPayloadAsMap()
		if err != nil {
			ret = 2
			break
		}

		// TODO sth
		payload["create"] = time.Now().Unix()
		// DONE sth
		request, err := zbClient.NewCompleteJobCommand().JobKey(jobKey).PayloadFromMap(payload)
		if err != nil {
			ret = 3
			break
		}

		request.Send()
		fmt.Println("Complete job", jobKey, "of type", job.Type)
		fmt.Println("Processing:", payload)
		break
	}

	if ret != 0 {
		fmt.Println("Failed to complete job", job.GetKey())
		client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send()
	}
	return
}
