# Blytz Live Auction - API Documentation Index

## Overview

This section contains comprehensive API documentation for all Blytz Live Auction platform services. Each service exposes RESTful APIs following consistent patterns for authentication, error handling, and response formats.

## Table of Contents

1. [Authentication Service API](#authentication-service-api)
2. [Product Service API](#product-service-api)
3. [Auction Service API](#auction-service-api)
4. [Order Service API](#order-service-api)
5. [Payment Service API](#payment-service-api)
6. [Chat Service API](#chat-service-api)
7. [Logistics Service API](#logistics-service-api)
8. [LiveKit Service API](#livekit-service-api)
9. [Notification Service API](#notification-service-api)
10. [Gateway Service API](#gateway-service-api)

## General API Information

### Base URLs

- **Development**: `http://localhost:8092/api/v1`
- **Staging**: `https://staging.blytz.app/api/v1`
- **Production**: `https://api.blytz.app/api/v1`

### Authentication

All API endpoints (except authentication endpoints) require a valid JWT token:

```http
Authorization: Bearer <jwt_token>
```

### Response Format

All APIs follow consistent response formats:

#### Success Response
```json
{
  "success": true,
  "data": {
    // Response data
  },
  "message": "Operation completed successfully",
  "timestamp": "2025-12-11T07:00:00Z"
}
```

#### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details"
  },
  "timestamp": "2025-12-11T07:00:00Z"
}
```

### HTTP Status Codes

| Code | Meaning | Description |
|-------|----------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required or invalid |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error occurred |

### Rate Limiting

- **Standard endpoints**: 100 requests per minute
- **Authentication endpoints**: 10 requests per minute
- **Payment endpoints**: 20 requests per minute

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

---

## Authentication Service API

### Base URL: `/auth`

### Endpoints

#### Register User
```http
POST /auth/register
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe",
  "phone": "+1234567890"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user",
      "is_active": true,
      "created_at": "2025-12-11T07:00:00Z"
    },
    "token": "jwt_token_here"
  }
}
```

#### Login User
```http
POST /auth/login
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user"
    },
    "token": "jwt_token_here",
    "expires_at": "2025-12-12T07:00:00Z"
  }
}
```

#### Refresh Token
```http
POST /auth/refresh
```

**Request Headers:**
```http
Authorization: Bearer <refresh_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "new_jwt_token_here",
    "expires_at": "2025-12-12T07:00:00Z"
  }
}
```

#### Logout User
```http
POST /auth/logout
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

#### Get User Profile
```http
GET /auth/profile
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "phone": "+1234567890",
    "role": "user",
    "is_active": true,
    "created_at": "2025-12-11T07:00:00Z",
    "updated_at": "2025-12-11T07:00:00Z"
  }
}
```

#### Update Profile
```http
PUT /auth/profile
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "name": "John Smith",
  "phone": "+1234567890",
  "bio": "Auction enthusiast"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Smith",
    "phone": "+1234567890",
    "bio": "Auction enthusiast",
    "updated_at": "2025-12-11T07:30:00Z"
  }
}
```

---

## Product Service API

### Base URL: `/products`

### Endpoints

#### List Products
```http
GET /products
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `category` (string): Filter by category
- `search` (string): Search term
- `sort` (string): Sort field (name, price, created_at)
- `order` (string): Sort order (asc, desc)

**Response:**
```json
{
  "success": true,
  "data": {
    "products": [
      {
        "id": "uuid",
        "name": "iPhone 15 Pro",
        "description": "Latest iPhone model",
        "category": "Electronics",
        "price": 999.99,
        "images": [
          "https://example.com/image1.jpg",
          "https://example.com/image2.jpg"
        ],
        "condition": "new",
        "seller_id": "uuid",
        "created_at": "2025-12-11T07:00:00Z",
        "updated_at": "2025-12-11T07:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "pages": 8
    }
  }
}
```

#### Get Product Details
```http
GET /products/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "iPhone 15 Pro",
    "description": "Latest iPhone model with advanced features",
    "category": "Electronics",
    "price": 999.99,
    "images": [
      "https://example.com/image1.jpg",
      "https://example.com/image2.jpg"
    ],
    "condition": "new",
    "seller_id": "uuid",
    "seller": {
      "id": "uuid",
      "name": "John Doe",
      "rating": 4.8
    },
    "specifications": {
      "brand": "Apple",
      "model": "iPhone 15 Pro",
      "storage": "256GB",
      "color": "Space Black"
    },
    "created_at": "2025-12-11T07:00:00Z",
    "updated_at": "2025-12-11T07:00:00Z"
  }
}
```

#### Create Product
```http
POST /products
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**Request Body:**
```
name: iPhone 15 Pro
description: Latest iPhone model
category: Electronics
price: 999.99
condition: new
images: [file1, file2]
specifications[brand]: Apple
specifications[model]: iPhone 15 Pro
specifications[storage]: 256GB
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "iPhone 15 Pro",
    "description": "Latest iPhone model",
    "category": "Electronics",
    "price": 999.99,
    "seller_id": "user_uuid",
    "created_at": "2025-12-11T07:00:00Z"
  }
}
```

#### Update Product
```http
PUT /products/{id}
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "name": "iPhone 15 Pro (Updated)",
  "price": 949.99,
  "description": "Updated description"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "iPhone 15 Pro (Updated)",
    "price": 949.99,
    "updated_at": "2025-12-11T07:30:00Z"
  }
}
```

#### Delete Product
```http
DELETE /products/{id}
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "message": "Product deleted successfully"
}
```

---

## Auction Service API

### Base URL: `/auctions`

### Endpoints

#### List Auctions
```http
GET /auctions
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `status` (string): Filter by status (active, ended, upcoming)
- `category` (string): Filter by category
- `search` (string): Search term

**Response:**
```json
{
  "success": true,
  "data": {
    "auctions": [
      {
        "id": "uuid",
        "product": {
          "id": "uuid",
          "name": "iPhone 15 Pro",
          "images": ["image1.jpg"]
        },
        "current_bid": 1050.00,
        "starting_bid": 999.99,
        "bid_increment": 25.00,
        "bid_count": 15,
        "status": "active",
        "is_live": true,
        "end_time": "2025-12-11T12:00:00Z",
        "seller_id": "uuid",
        "created_at": "2025-12-11T07:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "pages": 3
    }
  }
}
```

#### Get Auction Details
```http
GET /auctions/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "product": {
      "id": "uuid",
      "name": "iPhone 15 Pro",
      "description": "Latest iPhone model",
      "images": ["image1.jpg", "image2.jpg"],
      "specifications": {
        "brand": "Apple",
        "model": "iPhone 15 Pro"
      }
    },
    "current_bid": 1050.00,
    "starting_bid": 999.99,
    "bid_increment": 25.00,
    "bid_count": 15,
    "status": "active",
    "is_live": true,
    "end_time": "2025-12-11T12:00:00Z",
    "seller_id": "uuid",
    "seller": {
      "id": "uuid",
      "name": "John Doe",
      "rating": 4.8
    },
    "bids": [
      {
        "id": "uuid",
        "amount": 1050.00,
        "bidder_id": "uuid",
        "bidder": {
          "name": "Jane Smith"
        },
        "placed_at": "2025-12-11T07:30:00Z"
      }
    ],
    "created_at": "2025-12-11T07:00:00Z"
  }
}
```

#### Create Auction
```http
POST /auctions
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "product_id": "uuid",
  "starting_bid": 999.99,
  "bid_increment": 25.00,
  "duration_hours": 24,
  "is_live": true,
  "start_time": "2025-12-11T08:00:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "product_id": "uuid",
    "starting_bid": 999.99,
    "current_bid": 999.99,
    "status": "upcoming",
    "end_time": "2025-12-12T08:00:00Z",
    "created_at": "2025-12-11T07:00:00Z"
  }
}
```

#### Place Bid
```http
POST /auctions/{id}/bid
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "amount": 1075.00
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "bid": {
      "id": "uuid",
      "amount": 1075.00,
      "auction_id": "uuid",
      "bidder_id": "uuid",
      "placed_at": "2025-12-11T07:45:00Z"
    },
    "auction": {
      "current_bid": 1075.00,
      "bid_count": 16
    }
  }
}
```

#### Get Bidding History
```http
GET /auctions/{id}/bids
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page

**Response:**
```json
{
  "success": true,
  "data": {
    "bids": [
      {
        "id": "uuid",
        "amount": 1075.00,
        "bidder": {
          "id": "uuid",
          "name": "Jane Smith"
        },
        "placed_at": "2025-12-11T07:45:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 16,
      "pages": 1
    }
  }
}
```

---

## Order Service API

### Base URL: `/orders`

### Endpoints

#### Create Order
```http
POST /orders
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "auction_id": "uuid",
  "shipping_address": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "zip_code": "10001",
    "country": "USA"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "auction_id": "uuid",
    "buyer_id": "uuid",
    "seller_id": "uuid",
    "amount": 1075.00,
    "status": "pending_payment",
    "shipping_address": {
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "zip_code": "10001",
      "country": "USA"
    },
    "created_at": "2025-12-11T08:00:00Z"
  }
}
```

#### Get Order Details
```http
GET /orders/{id}
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "auction": {
      "id": "uuid",
      "product": {
        "name": "iPhone 15 Pro",
        "images": ["image1.jpg"]
      },
      "final_bid": 1075.00
    },
    "buyer_id": "uuid",
    "seller_id": "uuid",
    "amount": 1075.00,
    "status": "processing",
    "shipping_address": {
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "zip_code": "10001",
      "country": "USA"
    },
    "tracking_number": "1Z999AA1234567890",
    "created_at": "2025-12-11T08:00:00Z",
    "updated_at": "2025-12-11T09:00:00Z"
  }
}
```

#### List User Orders
```http
GET /orders
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `status` (string): Filter by status

**Response:**
```json
{
  "success": true,
  "data": {
    "orders": [
      {
        "id": "uuid",
        "auction": {
          "product": {
            "name": "iPhone 15 Pro",
            "images": ["image1.jpg"]
          }
        },
        "amount": 1075.00,
        "status": "delivered",
        "created_at": "2025-12-11T08:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "pages": 1
    }
  }
}
```

---

## Payment Service API

### Base URL: `/payments`

### Endpoints

#### Process Payment
```http
POST /payments/process
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "order_id": "uuid",
  "payment_method": "credit_card",
  "card_token": "tok_1234567890",
  "billing_address": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "zip_code": "10001",
    "country": "USA"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "payment_id": "pay_uuid",
    "order_id": "uuid",
    "amount": 1075.00,
    "status": "processing",
    "provider": "stripe",
    "provider_transaction_id": "ch_1234567890",
    "created_at": "2025-12-11T08:05:00Z"
  }
}
```

#### Get Payment Status
```http
GET /payments/{id}/status
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "payment_id": "pay_uuid",
    "status": "completed",
    "amount": 1075.00,
    "provider": "stripe",
    "provider_transaction_id": "ch_1234567890",
    "completed_at": "2025-12-11T08:06:00Z"
  }
}
```

#### Refund Payment
```http
POST /payments/{id}/refund
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "reason": "Item not as described",
  "amount": 1075.00
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "refund_id": "ref_uuid",
    "payment_id": "pay_uuid",
    "amount": 1075.00,
    "status": "processing",
    "created_at": "2025-12-11T09:00:00Z"
  }
}
```

---

## Chat Service API

### Base URL: `/chat`

### Endpoints

#### Get Chat Rooms
```http
GET /chat/rooms
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `auction_id` (string): Filter by auction

**Response:**
```json
{
  "success": true,
  "data": {
    "rooms": [
      {
        "id": "uuid",
        "auction_id": "uuid",
        "name": "iPhone 15 Pro Auction",
        "participant_count": 25,
        "created_at": "2025-12-11T07:00:00Z"
      }
    ]
  }
}
```

#### Get Chat Messages
```http
GET /chat/rooms/{id}/messages
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `before` (string): Get messages before this timestamp

**Response:**
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": "uuid",
        "room_id": "uuid",
        "user_id": "uuid",
        "user": {
          "name": "John Doe",
          "role": "seller"
        },
        "content": "This is a great item!",
        "type": "text",
        "created_at": "2025-12-11T07:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 100,
      "pages": 2
    }
  }
}
```

#### Send Message
```http
POST /chat/rooms/{id}/messages
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "content": "Is this item still available?",
  "type": "text"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "room_id": "uuid",
    "user_id": "uuid",
    "content": "Is this item still available?",
    "type": "text",
    "created_at": "2025-12-11T07:45:00Z"
  }
}
```

---

## LiveKit Service API

### Base URL: `/livekit`

### Endpoints

#### Get Live Token
```http
GET /livekit/token
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `room` (string): Room name (required)
- `role` (string): User role (viewer, host, broadcaster)

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "url": "wss://blytz-live-u5u72ozx.livekit.cloud",
    "room": "auction_uuid",
    "identity": "user_uuid_1704123456_789012345",
    "expires_at": "2025-12-11T13:00:00Z"
  }
}
```

#### Create Room
```http
POST /livekit/rooms
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "name": "auction_uuid",
  "empty_timeout": 300,
  "max_participants": 100,
  "record": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "room_id": "uuid",
    "name": "auction_uuid",
    "created_at": "2025-12-11T07:00:00Z"
  }
}
```

---

## Notification Service API

### Base URL: `/notifications`

### Endpoints

#### Get Notifications
```http
GET /notifications
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `type` (string): Filter by type
- `read` (boolean): Filter by read status

**Response:**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "type": "auction_won",
        "title": "You won the auction!",
        "message": "Congratulations! You won the iPhone 15 Pro auction.",
        "data": {
          "auction_id": "uuid",
          "product_name": "iPhone 15 Pro"
        },
        "read": false,
        "created_at": "2025-12-11T12:00:00Z"
      }
    ],
    "unread_count": 5,
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 25,
      "pages": 2
    }
  }
}
```

#### Mark as Read
```http
PUT /notifications/{id}/read
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "message": "Notification marked as read"
}
```

#### Update Preferences
```http
PUT /notifications/preferences
```

**Request Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "email_notifications": true,
  "push_notifications": true,
  "auction_reminders": true,
  "bid_notifications": true,
  "message_notifications": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "email_notifications": true,
    "push_notifications": true,
    "auction_reminders": true,
    "bid_notifications": true,
    "message_notifications": false
  }
}
```

---

## Gateway Service API

### Base URL: `/gateway`

### Endpoints

#### Health Check
```http
GET /gateway/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-12-11T07:00:00Z",
  "version": "1.0.0",
  "services": {
    "auth": "healthy",
    "products": "healthy",
    "auctions": "healthy",
    "orders": "healthy",
    "payments": "healthy",
    "chat": "healthy",
    "logistics": "healthy",
    "livekit": "healthy",
    "notifications": "healthy"
  }
}
```

#### API Status
```http
GET /gateway/status
```

**Response:**
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "uptime": "72h30m15s",
    "request_count": 1500000,
    "error_rate": 0.02,
    "average_response_time": 0.145,
    "services": {
      "auth": {
        "status": "healthy",
        "response_time": 0.050,
        "error_rate": 0.01
      },
      "auctions": {
        "status": "healthy",
        "response_time": 0.180,
        "error_rate": 0.03
      }
    }
  }
}
```

---

## Error Codes Reference

### Authentication Errors

| Code | Message | Description |
|-------|----------|-------------|
| AUTH_001 | Invalid credentials | Email or password incorrect |
| AUTH_002 | User not found | User does not exist |
| AUTH_003 | Token expired | JWT token has expired |
| AUTH_004 | Invalid token | JWT token is invalid |
| AUTH_005 | Account locked | User account is locked |
| AUTH_006 | Email not verified | User email not verified |

### Auction Errors

| Code | Message | Description |
|-------|----------|-------------|
| AUCTION_001 | Auction not found | Auction does not exist |
| AUCTION_002 | Auction ended | Auction has already ended |
| AUCTION_003 | Bid too low | Bid amount is below minimum |
| AUCTION_004 | Not authorized | User not authorized for this action |
| AUCTION_005 | Invalid bid increment | Bid increment is invalid |

### Payment Errors

| Code | Message | Description |
|-------|----------|-------------|
| PAYMENT_001 | Payment failed | Payment processing failed |
| PAYMENT_002 | Insufficient funds | Insufficient payment method funds |
| PAYMENT_003 | Card declined | Payment card was declined |
| PAYMENT_004 | Refund failed | Refund processing failed |

## SDKs and Libraries

### JavaScript/TypeScript
```bash
npm install @blytz/api-client
```

```typescript
import { BlytzAPI } from '@blytz/api-client';

const api = new BlytzAPI({
  baseURL: 'https://api.blytz.app/api/v1',
  token: 'your-jwt-token'
});

// Get auctions
const auctions = await api.auctions.list();

// Place bid
const bid = await api.auctions.placeBid('auction-id', {
  amount: 1000
});
```

### Go
```bash
go get github.com/gmsas95/blytz-live-sdk-go
```

```go
import "github.com/gmsas95/blytz-live-sdk-go"

client := blytz.NewClient("https://api.blytz.app/api/v1", "your-jwt-token")

// Get auctions
auctions, err := client.Auctions.List()

// Place bid
bid, err := client.Auctions.PlaceBid("auction-id", 1000.00)
```

### Python
```bash
pip install blytz-api-client
```

```python
from blytz_api_client import BlytzAPI

client = BlytzAPI(
    base_url='https://api.blytz.app/api/v1',
    token='your-jwt-token'
)

# Get auctions
auctions = client.auctions.list()

# Place bid
bid = client.auctions.place_bid('auction-id', 1000.00)
```

## Testing

### Postman Collection

Download our Postman collection:
[https://api.blytz.app/postman-collection](https://api.blytz.app/postman-collection)

### OpenAPI Specification

Full OpenAPI 3.0 specification available:
[https://api.blytz.app/openapi.json](https://api.blytz.app/openapi.json)

---

**Last Updated**: 2025-12-11  
**Version**: 1.0.0  
**Contact**: api-team@blytz.app

*This documentation will be updated as APIs evolve. Check back regularly for the latest information.*