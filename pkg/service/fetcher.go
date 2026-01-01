package service

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// resty is a library for making HTTP requests in Go. It is a wrapper around the net/http package.
// same as axios in javascript.
var client = resty.New().
	SetTimeout(3 * time.Second). // timeout after 3 seconds.
	SetRetryCount(2)             // retry 2 times if the request fails.

// function to call api to fetch user data, from another service.
func FetchUser(userID string) (interface{}, error) {
	resp, err := client.R().
		SetResult(map[string]interface{}{}).
		Get("http://localhost:9090/mock/user/" + userID)

	if err != nil {
		return nil, err
	}
	return resp.Result(), nil
}

// function to call api to fetch orders data, from another service.
func FetchOrders(userID string) (interface{}, error) {
	resp, err := client.R().
		SetResult(map[string]interface{}{}).
		Get("http://localhost:9090/mock/orders/" + userID)

	if err != nil {
		return nil, err
	}
	return resp.Result(), nil
}

// function to call api to fetch notifications data, from another service.
func FetchNotifications(userID string) (interface{}, error) {
	resp, err := client.R().
		SetResult(map[string]interface{}{}).
		Get("http://localhost:9090/mock/notifications/" + userID)

	if err != nil {
		return nil, err
	}
	return resp.Result(), nil
}
