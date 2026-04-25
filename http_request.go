package main

import (
	"fmt"
	"net/http"
)

func DoRequest(req *http.Request) (*http.Response, error) {
	return doRequestWithRetry(req, 1)
}

func doRequestWithRetry(req *http.Request, retries int) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && retries > 0 {
		resp.Body.Close()
		fmt.Println("Access token expired. Refetching one...")
		// Refetch an access token
		token = GetAccessToken(*etpRt)
		req.Header.Set("Authorization", "Bearer "+token)
		// and retry the request (with decremented retry count)
		return doRequestWithRetry(req, retries-1)
	}

	return resp, err
}
