#!/usr/bin/env python3
"""
Test script for ACP Gateway FastAPI implementation
"""
import requests
import json
import time

BASE_URL = "http://localhost:8080"
API_KEY = "test_key_123"
HEADERS = {"Authorization": f"Bearer {API_KEY}"}

def test_health():
    """Test health endpoint"""
    print("Testing health endpoint...")
    response = requests.get(f"{BASE_URL}/healthz")
    print(f"Health check: {response.status_code} - {response.json()}")
    return response.status_code == 200

def test_products():
    """Test products listing"""
    print("\nTesting products endpoint...")
    response = requests.get(f"{BASE_URL}/acp/v1/products", headers=HEADERS)
    print(f"Products: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Found {data['total']} products")
        for product in data['items']:
            print(f"  - {product['id']}: {product['title']} (${product['price']['value']})")
    return response.status_code == 200

def test_search_products():
    """Test product search"""
    print("\nTesting product search...")
    response = requests.get(f"{BASE_URL}/acp/v1/products?q=boots", headers=HEADERS)
    print(f"Search 'boots': {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Found {data['total']} matching products")
    return response.status_code == 200

def test_add_to_cart():
    """Test adding items to cart"""
    print("\nTesting add to cart intent...")
    intent = {
        "type": "acp.intent",
        "actor": "llm",
        "payload": {
            "action": "add_to_cart",
            "items": [
                {
                    "product_id": "sku_123",
                    "variant_id": "var_9",
                    "quantity": 2
                }
            ],
            "context": {
                "session_id": "test_session_123",
                "locale": "en-US"
            },
            "notes": "User wants size 9 work boots"
        }
    }
    
    response = requests.post(
        f"{BASE_URL}/acp/v1/intent",
        headers=HEADERS,
        json=intent
    )
    print(f"Add to cart: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Cart ID: {data['cart']['id']}")
        print(f"Items: {len(data['cart']['items'])}")
        print(f"Subtotal: ${data['cart']['subtotal']['value']}")
        return data['cart']['id']
    return None

def test_checkout(cart_id):
    """Test checkout process"""
    if not cart_id:
        print("\nSkipping checkout test - no cart ID")
        return False
        
    print(f"\nTesting checkout with cart {cart_id}...")
    checkout_data = {
        "cart_id": cart_id,
        "email": "test@example.com",
        "shipping_address": {
            "country": "US",
            "city": "San Francisco",
            "postal_code": "94102"
        },
        "payment": {
            "method": "card",
            "token": "tok_test_123"
        }
    }
    
    response = requests.post(
        f"{BASE_URL}/acp/v1/checkout",
        headers=HEADERS,
        json=checkout_data
    )
    print(f"Checkout: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Order ID: {data['order_id']}")
        print(f"Status: {data['payment_intent_status']}")
        print(f"Redirect URL: {data['redirect_url']}")
        return True
    return False

def test_sandbox_mode():
    """Test sandbox mode"""
    print("\nTesting sandbox mode...")
    intent = {
        "type": "acp.intent",
        "actor": "llm",
        "payload": {
            "action": "add_to_cart",
            "items": [
                {
                    "product_id": "sku_456",
                    "quantity": 1
                }
            ]
        }
    }
    
    # Test with sandbox header
    sandbox_headers = {**HEADERS, "X-ACP-Sandbox": "true"}
    response = requests.post(
        f"{BASE_URL}/acp/v1/intent",
        headers=sandbox_headers,
        json=intent
    )
    print(f"Sandbox mode (header): {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Sandbox flag: {data['sandbox']}")
        return True
    return False

def test_merchant_registration():
    """Test merchant registration"""
    print("\nTesting merchant registration...")
    merchant_data = {
        "name": "Test Store",
        "domain": "teststore.example.com",
        "api_base": "https://teststore.example.com/api",
        "metadata": {
            "category": "retail",
            "region": "US"
        }
    }
    
    response = requests.post(
        f"{BASE_URL}/merchant/register",
        headers=HEADERS,
        json=merchant_data
    )
    print(f"Merchant registration: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print(f"Merchant ID: {data['merchant_id']}")
        return True
    return False

def main():
    """Run all tests"""
    print("ACP Gateway FastAPI Test Suite")
    print("=" * 40)
    
    tests = [
        ("Health Check", test_health),
        ("List Products", test_products),
        ("Search Products", test_search_products),
        ("Merchant Registration", test_merchant_registration),
        ("Sandbox Mode", test_sandbox_mode),
    ]
    
    results = []
    for test_name, test_func in tests:
        try:
            result = test_func()
            results.append((test_name, result))
        except Exception as e:
            print(f"Error in {test_name}: {e}")
            results.append((test_name, False))
    
    # Test cart flow
    try:
        cart_id = test_add_to_cart()
        if cart_id:
            results.append(("Add to Cart", True))
            checkout_result = test_checkout(cart_id)
            results.append(("Checkout", checkout_result))
        else:
            results.append(("Add to Cart", False))
            results.append(("Checkout", False))
    except Exception as e:
        print(f"Error in cart flow: {e}")
        results.append(("Add to Cart", False))
        results.append(("Checkout", False))
    
    # Summary
    print("\n" + "=" * 40)
    print("Test Results:")
    passed = 0
    for test_name, result in results:
        status = "PASS" if result else "FAIL"
        print(f"  {test_name}: {status}")
        if result:
            passed += 1
    
    print(f"\nPassed: {passed}/{len(results)}")
    return passed == len(results)

if __name__ == "__main__":
    success = main()
    exit(0 if success else 1)
