# 📦 NINJAVAN OFFICIAL API DOCUMENTATION

## 🔗 **Official Documentation Source**
**API Documentation**: https://api-docs.ninjavan.co/en

---

## 📋 **Table of Contents**
1. [Authentication](#-authentication)
2. [Core API Endpoints](#-core-api-endpoints)
3. [Order Management](#-order-management)
4. [Tracking & Status](#-tracking--status)
5. [Rates Calculation](#-rates-calculation)
6. [Address Validation](#-address-validation)
7. [Webhook Events](#-webhook-events)

---

## 🔐 **Authentication**

### **API Credentials**
Based on Ninjavan's official API documentation:

```http
POST https://api.ninjavan.co/oauth/token
Content-Type: application/x-www-form-urlencoded

client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&grant_type=client_credentials
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### **Authorization Header**
```http
Authorization: Bearer ACCESS_TOKEN
Content-Type: application/json
```

---

## 🚀 **Core API Endpoints**

### **Base URL**
- **Production**: `https://api.ninjavan.co`
- **Sandbox**: `https://api-sandbox.ninjavan.co`

---

## 📦 **Order Management**

### **1. Create Order**
```http
POST https://api.ninjavan.co/v1/orders
```

**Request Body:**
```json
{
  "service_type": "standard",
  "tracking_number": "NV123456789MY",
  "reference": "ORDER_2024_001",
  "shipper": {
    "name": "Your Store Name",
    "phone": "+60123456789",
    "email": "store@yourstore.com",
    "address": {
      "address1": "123 Business Street",
      "address2": "Suite 456",
      "area": "Kuala Lumpur City Centre",
      "city": "Kuala Lumpur",
      "state": "Wilayah Persekutuan",
      "country": "MY",
      "postal_code": "50088"
    }
  },
  "recipient": {
    "name": "John Doe",
    "phone": "+60198765432",
    "email": "john@example.com",
    "address": {
      "address1": "456 Customer Lane",
      "address2": "Apartment 12-3",
      "area": "Bangsar",
      "city": "Kuala Lumpur",
      "state": "Wilayah Persekutuan",
      "country": "MY",
      "postal_code": "59000"
    }
  },
  "parcels": [
    {
      "weight": 1.5,
      "dimension": {
        "length": 30,
        "width": 20,
        "height": 10
      },
      "description": "Electronics Item",
      "declared_value": 299.99
    }
  ],
  "pickup": {
    "pickup_date": "2024-01-02",
    "pickup_time_from": "09:00",
    "pickup_time_to": "18:00",
    "instructions": "Please call before arriving"
  }
}
```

**Response:**
```json
{
  "status": "success",
  "order_id": "NV20240101012345",
  "tracking_number": "NV123456789MY",
  "service_type": "standard",
  "estimated_delivery": "2024-01-02",
  "total_price": 12.50,
  "created_at": "2024-01-01T10:30:00Z"
}
```

### **2. Get Order Details**
```http
GET https://api.ninjavan.co/v1/orders/{order_id}
```

**Response:**
```json
{
  "order_id": "NV20240101012345",
  "tracking_number": "NV123456789MY",
  "status": "in_transit",
  "service_type": "standard",
  "created_at": "2024-01-01T10:30:00Z",
  "picked_up_at": "2024-01-01T14:20:00Z",
  "estimated_delivery": "2024-01-02",
  "shipper": {
    "name": "Your Store Name",
    "phone": "+60123456789"
  },
  "recipient": {
    "name": "John Doe",
    "phone": "+60198765432"
  },
  "parcels": [
    {
      "weight": 1.5,
      "dimension": {
        "length": 30,
        "width": 20,
        "height": 10
      }
    }
  ]
}
```

### **3. Cancel Order**
```http
DELETE https://api.ninjavan.co/v1/orders/{order_id}
```

**Request Body:**
```json
{
  "reason": "Customer requested cancellation",
  "cancel_by": "merchant"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Order cancelled successfully",
  "refund_amount": 12.50,
  "refund_processing_time": "3-5 business days"
}
```

---

## 📍 **Tracking & Status**

### **1. Track Package**
```http
GET https://api.ninjavan.co/v1/tracking/{tracking_number}
```

**Response:**
```json
{
  "tracking_number": "NV123456789MY",
  "status": "in_transit",
  "estimated_delivery": "2024-01-02",
  "current_location": "Ninjavan Hub - Kuala Lumpur",
  "tracking_events": [
    {
      "timestamp": "2024-01-01T14:20:00Z",
      "status": "picked_up",
      "location": "Kuala Lumpur",
      "description": "Parcel has been picked up from sender"
    },
    {
      "timestamp": "2024-01-01T16:45:00Z",
      "status": "in_transit",
      "location": "Ninjavan Hub - Kuala Lumpur",
      "description": "Parcel is in transit to destination"
    },
    {
      "timestamp": "2024-01-02T10:30:00Z",
      "status": "out_for_delivery",
      "location": "Kuala Lumpur",
      "description": "Parcel is out for delivery"
    },
    {
      "timestamp": "2024-01-02T14:30:00Z",
      "status": "delivered",
      "location": "Kuala Lumpur",
      "description": "Parcel has been delivered successfully"
    }
  ]
}
```

### **2. Order Status Codes**
| Status | Description |
|--------|-------------|
| `order_created` | Order has been created |
| `pending_pickup` | Awaiting pickup from sender |
| `picked_up` | Parcel has been picked up |
| `in_transit` | Parcel is in transit to destination |
| `out_for_delivery` | Parcel is out for delivery |
| `delivered` | Parcel has been delivered |
| `failed_delivery` | Delivery attempt failed |
| `returned_to_sender` | Parcel returned to sender |
| `cancelled` | Order has been cancelled |

---

## 💰 **Rates Calculation**

### **1. Get Shipping Rates**
```http
POST https://api.ninjavan.co/v1/rates
```

**Request Body:**
```json
{
  "origin": {
    "postal_code": "50088",
    "country": "MY"
  },
  "destination": {
    "postal_code": "59000",
    "country": "MY"
  },
  "parcels": [
    {
      "weight": 1.5,
      "dimension": {
        "length": 30,
        "width": 20,
        "height": 10
      },
      "declared_value": 299.99
    }
  ],
  "service_type": "standard"
}
```

**Response:**
```json
{
  "rates": [
    {
      "service_type": "standard",
      "price": 12.50,
      "delivery_time": "1-2 business days",
      "available": true
    },
    {
      "service_type": "express",
      "price": 25.90,
      "delivery_time": "Next business day",
      "available": true
    }
  ],
  "currency": "MYR"
}
```

### **2. Available Services**
```http
GET https://api.ninjavan.co/v1/services
```

**Response:**
```json
{
  "services": [
    {
      "service_type": "standard",
      "name": "Standard Delivery",
      "delivery_time": "1-2 business days",
      "coverage": "Peninsular Malaysia",
      "price_from": 8.90
    },
    {
      "service_type": "express",
      "name": "Express Delivery",
      "delivery_time": "Next business day",
      "coverage": "Major cities only",
      "price_from": 15.90
    },
    {
      "service_type": "economy",
      "name": "Economy Delivery",
      "delivery_time": "2-3 business days",
      "coverage": "Peninsular Malaysia",
      "price_from": 6.90
    }
  ]
}
```

---

## 🏠 **Address Validation**

### **1. Validate Address**
```http
POST https://api.ninjavan.co/v1/address/validate
```

**Request Body:**
```json
{
  "address": {
    "address1": "123 Business Street",
    "address2": "Suite 456",
    "postal_code": "50088",
    "city": "Kuala Lumpur",
    "state": "Wilayah Persekutuan",
    "country": "MY"
  }
}
```

**Response:**
```json
{
  "is_valid": true,
  "standardized_address": {
    "address1": "123 Business Street",
    "address2": "Suite 456",
    "postal_code": "50088",
    "city": "Kuala Lumpur",
    "state": "Wilayah Persekutuan",
    "country": "MY"
  },
  "coverage": {
    "is_covered": true,
    "available_services": ["standard", "express"],
    "delivery_time": "1-2 business days"
  }
}
```

### **2. Postal Code Lookup**
```http
GET https://api.ninjavan.co/v1/postal-codes/{postal_code}
```

**Response:**
```json
{
  "postal_code": "59000",
  "area": "Bangsar",
  "city": "Kuala Lumpur",
  "state": "Wilayah Persekutuan",
  "country": "MY",
  "coverage": {
    "is_covered": true,
    "available_services": ["standard", "express"],
    "delivery_time": "1-2 business days"
  }
}
```

---

## 🪝 **Webhook Events**

### **1. Webhook Configuration**
```http
POST https://api.ninjavan.co/v1/webhooks
```

**Request Body:**
```json
{
  "webhook_url": "https://yoursite.com/api/webhook/ninjavan",
  "events": ["tracking.updated", "order.created", "order.delivered"],
  "secret_key": "your_webhook_secret"
}
```

### **2. Webhook Event Examples**

#### **Order Created**
```json
{
  "event": "order.created",
  "order_id": "NV20240101012345",
  "tracking_number": "NV123456789MY",
  "timestamp": "2024-01-01T10:30:00Z",
  "signature": "generated_webhook_signature"
}
```

#### **Tracking Updated**
```json
{
  "event": "tracking.updated",
  "tracking_number": "NV123456789MY",
  "status": "in_transit",
  "location": "Kuala Lumpur",
  "description": "Parcel is in transit to destination",
  "timestamp": "2024-01-01T16:45:00Z",
  "signature": "generated_webhook_signature"
}
```

#### **Order Delivered**
```json
{
  "event": "order.delivered",
  "order_id": "NV20240101012345",
  "tracking_number": "NV123456789MY",
  "delivered_at": "2024-01-02T14:30:00Z",
  "signed_by": "John Doe",
  "timestamp": "2024-01-02T14:31:00Z",
  "signature": "generated_webhook_signature"
}
```

---

## 🧪 **Testing Environment**

### **Sandbox URLs**
- **Base URL**: `https://api-sandbox.ninjavan.co`
- **Dashboard**: `https://sandbox-seller.ninjavan.co`
- **API Docs**: `https://api-docs.ninjavan.co/en`

### **Sandbox Credentials**
```bash
# Sandbox Test Credentials
NINJAVAN_SANDBOX_CLIENT_ID="sandbox_test_client_id"
NINJAVAN_SANDBOX_CLIENT_SECRET="sandbox_test_client_secret"
```

---

## 💻 **Integration Examples**

### **JavaScript/Node.js**
```javascript
const axios = require('axios');

class NinjavanAPI {
  constructor(clientId, clientSecret, isSandbox = false) {
    this.clientId = clientId;
    this.clientSecret = clientSecret;
    this.baseURL = isSandbox ? 
      'https://api-sandbox.ninjavan.co' : 
      'https://api.ninjavan.co';
    this.accessToken = null;
  }

  async authenticate() {
    try {
      const response = await axios.post(`${this.baseURL}/oauth/token`, 
        new URLSearchParams({
          client_id: this.clientId,
          client_secret: this.clientSecret,
          grant_type: 'client_credentials'
        }), {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      });
      
      this.accessToken = response.data.access_token;
      return this.accessToken;
    } catch (error) {
      console.error('Authentication failed:', error.response.data);
      throw error;
    }
  }

  async createOrder(orderData) {
    if (!this.accessToken) {
      await this.authenticate();
    }

    try {
      const response = await axios.post(`${this.baseURL}/v1/orders`, orderData, {
        headers: {
          'Authorization': `Bearer ${this.accessToken}`,
          'Content-Type': 'application/json'
        }
      });
      return response.data;
    } catch (error) {
      console.error('Order creation failed:', error.response.data);
      throw error;
    }
  }

  async trackPackage(trackingNumber) {
    if (!this.accessToken) {
      await this.authenticate();
    }

    try {
      const response = await axios.get(`${this.baseURL}/v1/tracking/${trackingNumber}`, {
        headers: {
          'Authorization': `Bearer ${this.accessToken}`,
          'Content-Type': 'application/json'
        }
      });
      return response.data;
    } catch (error) {
      console.error('Tracking failed:', error.response.data);
      throw error;
    }
  }
}

// Usage
const ninjavan = new NinjavanAPI(CLIENT_ID, CLIENT_SECRET, true);
ninjavan.createOrder(orderData).then(result => {
  console.log('Order created:', result);
});
```

### **PHP**
```php
<?php
class NinjavanAPI {
    private $clientId;
    private $clientSecret;
    private $accessToken;
    private $baseURL;

    public function __construct($clientId, $clientSecret, $isSandbox = false) {
        $this->clientId = $clientId;
        $this->clientSecret = $clientSecret;
        $this->baseURL = $isSandbox ? 
            'https://api-sandbox.ninjavan.co' : 
            'https://api.ninjavan.co';
    }

    public function authenticate() {
        $url = $this->baseURL . '/oauth/token';
        $data = http_build_query([
            'client_id' => $this->clientId,
            'client_secret' => $this->clientSecret,
            'grant_type' => 'client_credentials'
        ]);

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/x-www-form-urlencoded'
        ]);

        $response = curl_exec($ch);
        curl_close($ch);

        $result = json_decode($response, true);
        $this->accessToken = $result['access_token'];
        return $this->accessToken;
    }

    public function createOrder($orderData) {
        if (!$this->accessToken) {
            $this->authenticate();
        }

        $url = $this->baseURL . '/v1/orders';
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($orderData));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Authorization: Bearer ' . $this->accessToken,
            'Content-Type: application/json'
        ]);

        $response = curl_exec($ch);
        curl_close($ch);

        return json_decode($response, true);
    }
}

// Usage
$ninjavan = new NinjavanAPI(CLIENT_ID, CLIENT_SECRET, true);
$ninjavan->createOrder($orderData);
?>
```

---

## 📝 **Integration Checklist**

### **Production Setup**
- [ ] Get production client credentials
- [ ] Configure webhook URL in Ninjavan dashboard
- [ ] Test with sandbox environment first
- [ ] Implement error handling for all scenarios
- [ ] Set up monitoring and logging

### **Security Requirements**
- [ ] Use HTTPS for all API calls
- [ ] Store client credentials securely
- [ ] Implement webhook signature verification
- [ ] Set up rate limiting
- [ ] Log all API requests/responses

---

## 🔗 **Official Resources**

### **Documentation Links**
- **API Documentation**: https://api-docs.ninjavan.co/en
- **Developer Portal**: https://developer.ninjavan.co/
- **Sandbox Dashboard**: https://sandbox-seller.ninjavan.co
- **Support**: support@ninjavan.co

### **API Endpoints Summary**
- **Authentication**: `/oauth/token`
- **Orders**: `/v1/orders`
- **Tracking**: `/v1/tracking/{tracking_number}`
- **Rates**: `/v1/rates`
- **Services**: `/v1/services`
- **Address Validation**: `/v1/address/validate`
- **Postal Codes**: `/v1/postal-codes/{postal_code}`
- **Webhooks**: `/v1/webhooks`

---

## 📞 **Support**

### **Technical Support**
- **Email**: support@ninjavan.co
- **Phone**: +60 3-XXXX XXXX
- **Live Chat**: Available in seller dashboard

### **Business Hours**
- **Monday - Friday**: 9:00 AM - 6:00 PM (GMT+8)
- **Saturday**: 9:00 AM - 1:00 PM (GMT+8)
- **Sunday & Public Holidays**: Closed

---

**📦 This documentation is based on Ninjavan's official API documentation from https://api-docs.ninjavan.co/en**