package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const BASE_URL = "http://localhost:8082"
const API_KEY = "test_key_123"

var HEADERS = map[string]string{
	"Authorization": "Bearer " + API_KEY,
	"Content-Type":  "application/json",
}

func main() {
	fmt.Println("🚀 Testing ACP Gateway Go Implementation")
	fmt.Println("========================================")
	
	// Test health endpoint
	fmt.Println("\n1. Testing Health Check...")
	testHealth()
	
	// Test products endpoint
	fmt.Println("\n2. Testing Products Endpoint...")
	testProducts()
	
	// Test intent endpoint
	fmt.Println("\n3. Testing Intent Endpoint...")
	testIntent()
	
	// Test checkout endpoint
	fmt.Println("\n4. Testing Checkout Endpoint...")
	testCheckout()
	
	// Test merchant registration
	fmt.Println("\n5. Testing Merchant Registration...")
	testMerchantRegister()
	
	// Test merchant feed
	fmt.Println("\n6. Testing Merchant Feed...")
	testMerchantFeed()
	
	fmt.Println("\n✅ All tests completed!")
}

func testHealth() {
	resp, err := http.Get(BASE_URL + "/healthz")
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Health check: %s (Status: %d)\n", string(body), resp.StatusCode)
}

func testProducts() {
	resp, err := http.Get(BASE_URL + "/acp/v1/products")
	if err != nil {
		fmt.Printf("❌ Products request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Products: %s\n", string(body))
}

func testIntent() {
	intent := map[string]interface{}{
		"type":  "acp.intent",
		"actor": "llm",
		"payload": map[string]interface{}{
			"action": "add_to_cart",
			"items": []map[string]interface{}{
				{
					"id":       "sku_123",
					"quantity": 1,
				},
			},
			"context": map[string]interface{}{
				"session_id": "session_123",
			},
		},
	}
	
	jsonData, _ := json.Marshal(intent)
	req, _ := http.NewRequest("POST", BASE_URL+"/acp/v1/intent", bytes.NewBuffer(jsonData))
	
	for key, value := range HEADERS {
		req.Header.Set(key, value)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Intent request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Intent: %s\n", string(body))
}

func testCheckout() {
	checkout := map[string]interface{}{
		"cart_id": "session_123",
		"buyer": map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john@example.com",
		},
		"shipping_address": map[string]interface{}{
			"name":        "John Doe",
			"line_one":    "123 Main St",
			"city":        "Anytown",
			"state":       "CA",
			"country":     "US",
			"postal_code": "12345",
		},
		"payment": map[string]interface{}{
			"method": "card",
			"token":  "tok_123",
		},
	}
	
	jsonData, _ := json.Marshal(checkout)
	req, _ := http.NewRequest("POST", BASE_URL+"/acp/v1/checkout", bytes.NewBuffer(jsonData))
	
	for key, value := range HEADERS {
		req.Header.Set(key, value)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Checkout request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Checkout: %s\n", string(body))
}

func testMerchantRegister() {
	merchant := map[string]interface{}{
		"name":   "Test Merchant",
		"domain": "testmerchant.com",
		"metadata": map[string]interface{}{
			"description": "A test merchant for ACP Gateway",
		},
	}
	
	jsonData, _ := json.Marshal(merchant)
	req, _ := http.NewRequest("POST", BASE_URL+"/merchant/register", bytes.NewBuffer(jsonData))
	
	for key, value := range HEADERS {
		req.Header.Set(key, value)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Merchant registration failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Merchant registration: %s\n", string(body))
}

func testMerchantFeed() {
	feed := map[string]interface{}{
		"products": []map[string]interface{}{
			{
				"id":          "new_sku_789",
				"title":       "New Product",
				"description": "A new product from the feed",
				"price": map[string]interface{}{
					"value":    99.99,
					"currency": "USD",
				},
				"available": true,
			},
		},
	}
	
	jsonData, _ := json.Marshal(feed)
	req, _ := http.NewRequest("POST", BASE_URL+"/merchant/feed", bytes.NewBuffer(jsonData))
	
	for key, value := range HEADERS {
		req.Header.Set(key, value)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Merchant feed failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Merchant feed: %s\n", string(body))
}
