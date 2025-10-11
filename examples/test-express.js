#!/usr/bin/env node
/**
 * Test script for ACP Gateway Express implementation
 */
const axios = require('axios');

const BASE_URL = 'http://localhost:8080';
const API_KEY = 'test_key_123';
const HEADERS = { 'Authorization': `Bearer ${API_KEY}` };

async function testHealth() {
  console.log('Testing health endpoint...');
  try {
    const response = await axios.get(`${BASE_URL}/healthz`);
    console.log(`Health check: ${response.status} - ${JSON.stringify(response.data)}`);
    return response.status === 200;
  } catch (error) {
    console.log(`Health check failed: ${error.message}`);
    return false;
  }
}

async function testProducts() {
  console.log('\nTesting products endpoint...');
  try {
    const response = await axios.get(`${BASE_URL}/acp/v1/products`, { headers: HEADERS });
    console.log(`Products: ${response.status}`);
    if (response.status === 200) {
      const data = response.data;
      console.log(`Found ${data.total} products`);
      data.items.forEach(product => {
        console.log(`  - ${product.id}: ${product.title} ($${product.price.value})`);
      });
    }
    return response.status === 200;
  } catch (error) {
    console.log(`Products test failed: ${error.message}`);
    return false;
  }
}

async function testSearchProducts() {
  console.log('\nTesting product search...');
  try {
    const response = await axios.get(`${BASE_URL}/acp/v1/products?q=boots`, { headers: HEADERS });
    console.log(`Search 'boots': ${response.status}`);
    if (response.status === 200) {
      const data = response.data;
      console.log(`Found ${data.total} matching products`);
    }
    return response.status === 200;
  } catch (error) {
    console.log(`Search test failed: ${error.message}`);
    return false;
  }
}

async function testAddToCart() {
  console.log('\nTesting add to cart intent...');
  const intent = {
    type: 'acp.intent',
    actor: 'llm',
    payload: {
      action: 'add_to_cart',
      items: [
        {
          product_id: 'sku_123',
          variant_id: 'var_9',
          quantity: 2
        }
      ],
      context: {
        session_id: 'test_session_123',
        locale: 'en-US'
      },
      notes: 'User wants size 9 work boots'
    }
  };
  
  try {
    const response = await axios.post(`${BASE_URL}/acp/v1/intent`, intent, { headers: HEADERS });
    console.log(`Add to cart: ${response.status}`);
    if (response.status === 200) {
      const data = response.data;
      console.log(`Cart ID: ${data.cart.id}`);
      console.log(`Items: ${data.cart.items.length}`);
      console.log(`Subtotal: $${data.cart.subtotal.value}`);
      return data.cart.id;
    }
    return null;
  } catch (error) {
    console.log(`Add to cart failed: ${error.message}`);
    return null;
  }
}

async function testCheckout(cartId) {
  if (!cartId) {
    console.log('\nSkipping checkout test - no cart ID');
    return false;
  }
    
  console.log(`\nTesting checkout with cart ${cartId}...`);
  const checkoutData = {
    cart_id: cartId,
    email: 'test@example.com',
    shipping_address: {
      country: 'US',
      city: 'San Francisco',
      postal_code: '94102'
    },
    payment: {
      method: 'card',
      token: 'tok_test_123'
    }
  };
  
  try {
    const response = await axios.post(`${BASE_URL}/acp/v1/checkout`, checkoutData, { headers: HEADERS });
    console.log(`Checkout: ${response.status}`);
    if (response.status === 200) {
      const data = response.data;
      console.log(`Order ID: ${data.order_id}`);
      console.log(`Status: ${data.payment_intent_status}`);
      console.log(`Redirect URL: ${data.redirect_url}`);
      return true;
    }
    return false;
  } catch (error) {
    console.log(`Checkout failed: ${error.message}`);
    return false;
  }
}

async function testSandboxMode() {
  console.log('\nTesting sandbox mode...');
  const intent = {
    type: 'acp.intent',
    actor: 'llm',
    payload: {
      action: 'add_to_cart',
      items: [
        {
          product_id: 'sku_456',
          quantity: 1
        }
      ]
    }
  };
  
  try {
    // Test with sandbox header
    const sandboxHeaders = { ...HEADERS, 'X-ACP-Sandbox': 'true' };
    const response = await axios.post(`${BASE_URL}/acp/v1/intent`, intent, { headers: sandboxHeaders });
    console.log(`Sandbox mode (header): ${response.status}`);
    if (response.status === 200) {
      const data = response.data;
      console.log(`Sandbox flag: ${data.sandbox}`);
      return true;
    }
    return false;
  } catch (error) {
    console.log(`Sandbox test failed: ${error.message}`);
    return false;
  }
}

async function testMerchantRegistration() {
  console.log('\nTesting merchant registration...');
  const merchantData = {
    name: 'Test Store',
    domain: 'teststore.example.com',
    api_base: 'https://teststore.example.com/api',
    metadata: {
      category: 'retail',
      region: 'US'
    }
  };
  
  try {
    const response = await axios.post(`${BASE_URL}/merchant/register`, merchantData, { headers: HEADERS });
    console.log(`Merchant registration: ${response.status}`);
    if (response.status === 200) {
      const data = response.data;
      console.log(`Merchant ID: ${data.merchant_id}`);
      return true;
    }
    return false;
  } catch (error) {
    console.log(`Merchant registration failed: ${error.message}`);
    return false;
  }
}

async function main() {
  console.log('ACP Gateway Express Test Suite');
  console.log('='.repeat(40));
  
  const tests = [
    ['Health Check', testHealth],
    ['List Products', testProducts],
    ['Search Products', testSearchProducts],
    ['Merchant Registration', testMerchantRegistration],
    ['Sandbox Mode', testSandboxMode],
  ];
  
  const results = [];
  for (const [testName, testFunc] of tests) {
    try {
      const result = await testFunc();
      results.push([testName, result]);
    } catch (error) {
      console.log(`Error in ${testName}: ${error.message}`);
      results.push([testName, false]);
    }
  }
  
  // Test cart flow
  try {
    const cartId = await testAddToCart();
    if (cartId) {
      results.push(['Add to Cart', true]);
      const checkoutResult = await testCheckout(cartId);
      results.push(['Checkout', checkoutResult]);
    } else {
      results.push(['Add to Cart', false]);
      results.push(['Checkout', false]);
    }
  } catch (error) {
    console.log(`Error in cart flow: ${error.message}`);
    results.push(['Add to Cart', false]);
    results.push(['Checkout', false]);
  }
  
  // Summary
  console.log('\n' + '='.repeat(40));
  console.log('Test Results:');
  let passed = 0;
  for (const [testName, result] of results) {
    const status = result ? 'PASS' : 'FAIL';
    console.log(`  ${testName}: ${status}`);
    if (result) passed++;
  }
  
  console.log(`\nPassed: ${passed}/${results.length}`);
  return passed === results.length;
}

if (require.main === module) {
  main().then(success => {
    process.exit(success ? 0 : 1);
  }).catch(error => {
    console.error('Test suite failed:', error);
    process.exit(1);
  });
}

module.exports = {
  testHealth,
  testProducts,
  testSearchProducts,
  testAddToCart,
  testCheckout,
  testSandboxMode,
  testMerchantRegistration
};
