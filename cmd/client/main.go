package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var gateway = "http://localhost:9091"
var token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImJkZWQ4ZjUxODlhODJlZDllNzQ1ZGVjNWQ5MGVkNDMzIn0.eyJpc3MiOiJodHRwczovL2lkcC5sb2NhbCIsImF1ZCI6Im15X2NsaWVudF9hcHAiLCJzdWIiOiI1YmU4NjM1OTA3M2M0MzRiYWQyZGEzOTMyMjIyZGFiZSIsImV4cCI6MTI1NzkzMDAwMCwiaWQiOjQ4OTM0fQ.C5CUnvsLCYq3BdO9-ZcCf5QpnEAmc2VkBvKBeMnDf6K4TBvG5xNZDk80G22GXeE57vbY4aUtGIh4n7pH-2dc8FD4Z9JZNA9-6yKb9W6xM3BzeERkvczqk-EiqX6ybR-cDXl2itreu6hi2A-G0iFqUlY0TZQPLlB-RP-ILkKoc0VKB0a6BIDethsWNTb-Ib8Xy0ht9BuFbbUrrzIwUY1OBbSu1Qj5VlsvDqpLfrA84SEQXELzWoKHjdZo-AW-WYQi9PcSpEaB-u5K_M8fZaH4M3Xin-8ZZ2cTm1hMuLUUwVoSLbzzBDSfdB7Ke1hXSHrbemwbZf7z_k5glBF0L_uZFOgn6whX_Dyq4vuJoXu0XUoEDnZ0Sr0m2eGHuOkW5ksODPdBYIWUioNWjRJjmYd7gfIsGylp1RjCG7OAAR2-IQWUlVtQrLD9pGu4ANV_huNnxNyD2BUkVJExs2dS1LU4769fnmPvZYz1q3G43HrYufQTPQCSfMJR_M0fFKMxzVaa"

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg.Add(20)
	for range 20 {
		delay := time.Duration(rand.Intn(9)+1*1000) * time.Millisecond
		time.Sleep(delay)
		go func() {
			doRequest(ctx, &wg)
		}()
	}

	wg.Wait()
}

func doRequest(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	path := fmt.Sprintf("%s/product-categories?limit=20&offset=0", gateway)
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("x-service-name", "users")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("erro: ", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode > 299 {
		body, _ := io.ReadAll(response.Body)
		fmt.Printf("%d: %s", response.StatusCode, string(body))
		return
	}
	var data interface{}
	json.NewDecoder(response.Body).Decode(&data)
	//b, _ := json.MarshalIndent(data, "", " ")
	fmt.Println(response.StatusCode)
}
