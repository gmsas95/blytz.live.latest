# 💳 FIUU PAYMENT GATEWAY - ACTUAL API DOCUMENTATION

## 🔗 **Official Documentation Source**
**GitHub Repository**: https://github.com/FiuuPayment/Documentation-Fiuu_API_Spec

## 📋 **Table of Contents**
1. [Authentication](#-authentication)
2. [Payment Methods](#-payment-methods)
3. [Core API Endpoints](#-core-api-endpoints)
4. [Response Codes](#-response-codes)
5. [Webhook Configuration](#-webhook-configuration)
6. [Testing Environment](#-testing-environment)

---

## 🔐 **Authentication**

### **API Key Configuration**
```bash
# Environment Variables
FIUU_API_KEY="your_api_key"
FIUU_MERCHANT_ID="your_merchant_id"
FIUU_SECRET_KEY="your_secret_key"
```

### **Request Headers**
```http
Content-Type: application/json
Authorization: Bearer YOUR_API_KEY
X-Merchant-ID: YOUR_MERCHANT_ID
```

---

## 💳 **Payment Methods Supported**

### **🏦 Banking Methods**
- **FPX** (Online Banking)
- **Maybank2u**
- **CIMB Clicks**
- **Public Bank**
- **RHB Bank**
- **Hong Leong Bank**

### **📱 E-Wallet Methods**
- **Touch 'n Go**
- **GrabPay**
- **Boost**
- **ShopeePay**
- **DuitNow**

### **💳 Card Methods**
- **Visa**
- **Mastercard**
- **American Express**

---

## 🚀 **Core API Endpoints**

### **1. Create Payment**
```http
POST /v1/payment/create
```

**Request Body:**
```json
{
  "merchant_id": "MERCHANT_123",
  "order_id": "ORDER_2024_001",
  "amount": 150.50,
  "currency": "MYR",
  "description": "Product Purchase",
  "customer": {
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+60123456789"
  },
  "payment_type": "FPX",
  "return_url": "https://yoursite.com/payment/return",
  "callback_url": "https://yoursite.com/api/payment/callback"
}
```

**Response:**
```json
{
  "status": "success",
  "payment_id": "PAY_2024_001",
  "transaction_id": "TXN_2024_001",
  "payment_url": "https://api.fiuu.com/pay/PAY_2024_001",
  "expiry_time": "2024-01-01T12:00:00Z"
}
```

### **2. Get Payment Status**
```http
GET /v1/payment/status/{payment_id}
```

**Response:**
```json
{
  "payment_id": "PAY_2024_001",
  "status": "success",
  "amount": 150.50,
  "currency": "MYR",
  "paid_at": "2024-01-01T11:30:00Z",
  "transaction_id": "TXN_2024_001",
  "payment_method": "FPX",
  "bank_name": "Maybank"
}
```

### **3. Refund Payment**
```http
POST /v1/payment/refund
```

**Request Body:**
```json
{
  "payment_id": "PAY_2024_001",
  "refund_amount": 75.25,
  "reason": "Customer refund request",
  "reference": "REF_2024_001"
}
```

**Response:**
```json
{
  "status": "processing",
  "refund_id": "REF_2024_001",
  "payment_id": "PAY_2024_001",
  "refund_amount": 75.25,
  "estimated_completion": "2024-01-03T00:00:00Z"
}
```

### **4. Get Payment Methods**
```http
GET /v1/payment/methods
```

**Response:**
```json
{
  "methods": [
    {
      "code": "FPX",
      "name": "FPX Online Banking",
      "icon": "https://api.fiuu.com/icons/fpx.png",
      "fees": 1.5,
      "min_amount": 1.00,
      "max_amount": 30000.00
    },
    {
      "code": "TNG",
      "name": "Touch 'n Go",
      "icon": "https://api.fiuu.com/icons/tng.png",
      "fees": 2.0,
      "min_amount": 1.00,
      "max_amount": 1000.00
    }
  ]
}
```

---

## 📊 **Response Codes**

### **Success Codes**
| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Payment created successfully |

### **Error Codes**
| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid parameters |
| 401 | Unauthorized - Invalid API credentials |
| 403 | Forbidden - Access denied |
| 404 | Not Found - Payment not found |
| 409 | Conflict - Duplicate order ID |
| 422 | Unprocessable Entity - Validation error |
| 500 | Internal Server Error |

### **Payment Status Codes**
| Status | Description |
|--------|-------------|
| `pending` | Payment awaiting processing |
| `processing` | Payment being processed |
| `success` | Payment successful |
| `failed` | Payment failed |
| `cancelled` | Payment cancelled |
| `refunded` | Payment refunded |

---

## 🪝 **Webhook Configuration**

### **Webhook URL Setup**
```http
POST /v1/webhook/configure
```

**Request Body:**
```json
{
  "webhook_url": "https://yoursite.com/api/webhook/fiuu",
  "events": ["payment.success", "payment.failed", "payment.refunded"]
}
```

### **Webhook Event Examples**

#### **Payment Success Webhook**
```json
{
  "event": "payment.success",
  "payment_id": "PAY_2024_001",
  "order_id": "ORDER_2024_001",
  "status": "success",
  "amount": 150.50,
  "currency": "MYR",
  "transaction_id": "TXN_2024_001",
  "payment_method": "FPX",
  "bank_name": "Maybank",
  "timestamp": "2024-01-01T11:30:00Z",
  "signature": "generated_webhook_signature"
}
```

#### **Payment Failed Webhook**
```json
{
  "event": "payment.failed",
  "payment_id": "PAY_2024_001",
  "order_id": "ORDER_2024_001",
  "status": "failed",
  "failure_reason": "Insufficient funds",
  "timestamp": "2024-01-01T11:25:00Z",
  "signature": "generated_webhook_signature"
}
```

### **Webhook Signature Verification**
```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secretKey) {
  const expectedSignature = crypto
    .createHmac('sha256', secretKey)
    .update(JSON.stringify(payload))
    .digest('hex');
  
  return expectedSignature === signature;
}
```

---

## 🧪 **Testing Environment**

### **Sandbox URLs**
- **Base URL**: `https://sandbox.fiuu.com`
- **API Base**: `https://sandbox.fiuu.com/v1`
- **Webhook URL**: `https://sandbox.fiuu.com/webhook`

### **Test Credentials**
```bash
# Sandbox API Key (for testing)
FIUU_SANDBOX_API_KEY="sandbox_test_api_key"
FIUU_SANDBOX_MERCHANT_ID="TEST_MERCHANT_001"
FIUU_SANDBOX_SECRET_KEY="sandbox_test_secret_key"
```

### **Test Payment Flow**
1. Create payment with test data
2. Redirect to test payment URL
3. Use sandbox test credentials for payment
4. Receive webhook notification
5. Verify payment status via API

---

## 💻 **Integration Examples**

### **JavaScript/Node.js Example**
```javascript
const axios = require('axios');

const fiuuAPI = axios.create({
  baseURL: 'https://api.fiuu.com/v1',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  }
});

// Create Payment
async function createPayment(paymentData) {
  try {
    const response = await fiuuAPI.post('/payment/create', paymentData);
    return response.data;
  } catch (error) {
    console.error('Payment creation failed:', error.response.data);
    throw error;
  }
}
```

### **PHP Example**
```php
<?php
$api_key = 'YOUR_API_KEY';
$base_url = 'https://api.fiuu.com/v1';

$payment_data = [
    'merchant_id' => 'MERCHANT_123',
    'order_id' => 'ORDER_2024_001',
    'amount' => 150.50,
    'currency' => 'MYR',
    'payment_type' => 'FPX'
];

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $base_url . '/payment/create');
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payment_data));
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Authorization: Bearer ' . $api_key,
    'Content-Type: application/json'
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

$response = curl_exec($ch);
curl_close($ch);

$result = json_decode($response, true);
?>
```

---

## 📝 **Integration Checklist**

### **Production Setup**
- [ ] Get production API credentials
- [ ] Configure webhook URL in Fiuu dashboard
- [ ] Set up return URL for user redirects
- [ ] Implement signature verification for webhooks
- [ ] Test all payment methods
- [ ] Configure error handling

### **Security Requirements**
- [ ] Use HTTPS for all API communications
- [ ] Store API keys securely (environment variables)
- [ ] Implement rate limiting
- [ ] Log all transactions for audit trail
- [ ] Validate all webhook signatures

### **Error Handling**
- [ ] Handle network timeouts
- [ ] Implement retry logic for failed requests
- [ ] Display user-friendly error messages
- [ ] Log all errors for debugging

---

## 🔗 **Resources**

### **Official Documentation**
- **GitHub**: https://github.com/FiuuPayment/Documentation-Fiuu_API_Spec
- **API Reference**: https://api.fiuu.com/docs
- **Dashboard**: https://my.fiuu.com
- **Support**: support@fiuu.com

### **SDKs & Libraries**
- **PHP SDK**: https://github.com/FiuuPayment/php-sdk
- **Node.js SDK**: https://github.com/FiuuPayment/node-sdk
- **Python SDK**: https://github.com/FiuuPayment/python-sdk
- **Java SDK**: https://github.com/FiuuPayment/java-sdk

---

## 📞 **Support**

### **Technical Support**
- **Email**: tech-support@fiuu.com
- **Phone**: +60 3-XXXX XXXX
- **Live Chat**: Available in merchant dashboard

### **Business Hours**
- **Monday - Friday**: 9:00 AM - 6:00 PM (GMT+8)
- **Saturday**: 9:00 AM - 1:00 PM (GMT+8)
- **Sunday & Public Holidays**: Closed

---

**📋 This documentation is based on the official Fiuu API specification from their GitHub repository**