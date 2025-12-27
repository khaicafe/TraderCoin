# Telegram Notification API Documentation

## Overview

This API provides endpoints to configure and manage Telegram notifications for trading activities.

## Authentication

All endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

## Endpoints

### 1. Get Telegram Configuration

**GET** `/api/v1/telegram/config`

Get the current user's Telegram configuration.

**Response:**

```json
{
  "id": 1,
  "user_id": 123,
  "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
  "chat_id": "123456789",
  "bot_name": "My Trading Bot",
  "is_enabled": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**

- `200 OK` - Success
- `404 Not Found` - Configuration not found
- `401 Unauthorized` - Invalid or missing token

---

### 2. Create or Update Telegram Configuration

**POST** `/api/v1/telegram/config`

Create a new Telegram configuration or update the existing one.

**Request Body:**

```json
{
  "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
  "chat_id": "123456789",
  "bot_name": "My Trading Bot",
  "is_enabled": true
}
```

**Response:**

```json
{
  "message": "Telegram configuration created successfully",
  "config": {
    "id": 1,
    "user_id": 123,
    "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
    "chat_id": "123456789",
    "bot_name": "My Trading Bot",
    "is_enabled": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Status Codes:**

- `201 Created` - Configuration created
- `200 OK` - Configuration updated
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid or missing token

---

### 3. Delete Telegram Configuration

**DELETE** `/api/v1/telegram/config`

Delete the user's Telegram configuration.

**Response:**

```json
{
  "message": "Telegram configuration deleted successfully"
}
```

**Status Codes:**

- `200 OK` - Success
- `401 Unauthorized` - Invalid or missing token
- `500 Internal Server Error` - Failed to delete

---

### 4. Test Telegram Connection

**POST** `/api/v1/telegram/test-connection`

Test the connection to Telegram with provided credentials.

**Request Body:**

```json
{
  "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
  "chat_id": "123456789"
}
```

**Response:**

```json
{
  "message": "Telegram connection test successful",
  "user_id": 123
}
```

**Error Response:**

```json
{
  "error": "Failed to connect to Telegram",
  "details": "unauthorized: bot token is invalid"
}
```

**Status Codes:**

- `200 OK` - Connection successful
- `400 Bad Request` - Connection failed or invalid credentials
- `401 Unauthorized` - Invalid or missing token

---

### 5. Send Test Message

**POST** `/api/v1/telegram/test-message`

Send a test message to the user's configured Telegram chat.

**Response:**

```json
{
  "message": "Test message sent successfully"
}
```

**Error Response:**

```json
{
  "error": "Failed to send test message",
  "details": "telegram configuration not found or disabled for user"
}
```

**Status Codes:**

- `200 OK` - Message sent successfully
- `400 Bad Request` - Failed to send message
- `401 Unauthorized` - Invalid or missing token

---

### 6. Toggle Telegram Notifications

**PATCH** `/api/v1/telegram/toggle`

Enable or disable Telegram notifications.

**Request Body:**

```json
{
  "is_enabled": true
}
```

**Response:**

```json
{
  "message": "Telegram notifications enabled successfully",
  "config": {
    "id": 1,
    "user_id": 123,
    "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
    "chat_id": "123456789",
    "bot_name": "My Trading Bot",
    "is_enabled": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Status Codes:**

- `200 OK` - Success
- `404 Not Found` - Configuration not found
- `401 Unauthorized` - Invalid or missing token

---

## Admin Endpoints

### 7. Get All Telegram Configurations (Admin)

**GET** `/api/v1/admin/telegram`

Get all Telegram configurations with pagination (Admin only).

**Query Parameters:**

- `page` (default: 1) - Page number
- `limit` (default: 20) - Items per page

**Response:**

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 123,
      "bot_token": "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
      "chat_id": "123456789",
      "bot_name": "Trading Bot 1",
      "is_enabled": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "User": {
        "id": 123,
        "email": "user@example.com",
        "full_name": "John Doe"
      }
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

**Status Codes:**

- `200 OK` - Success
- `401 Unauthorized` - Invalid or missing admin token

---

## How to Get Telegram Bot Token and Chat ID

### Getting Bot Token:

1. Open Telegram and search for **@BotFather**
2. Send `/newbot` command
3. Follow instructions to create a new bot
4. BotFather will give you a bot token like: `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz`

### Getting Chat ID:

1. Start a conversation with your bot
2. Send any message to the bot
3. Open this URL in browser: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Look for `"chat":{"id":123456789}` in the response
5. That number is your chat ID

---

## Notification Types

The system automatically sends the following notification types when configured:

### 1. Bot Status Updates

Sent when a bot's status changes (started, stopped, paused, etc.)

### 2. Order Notifications

Sent when orders are placed, filled, or cancelled

### 3. Trade Notifications

Sent when trades are executed

### 4. Error Alerts

Sent when errors occur in trading operations

### 5. Bot Paused Alerts

Sent when a bot is automatically paused due to errors or other conditions

---

## Usage Example

```javascript
// 1. Test connection first
const testResponse = await fetch('/api/v1/telegram/test-connection', {
  method: 'POST',
  headers: {
    Authorization: 'Bearer YOUR_TOKEN',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    bot_token: '1234567890:ABCdefGHIjklMNOpqrsTUVwxyz',
    chat_id: '123456789',
  }),
});

// 2. If successful, save configuration
const saveResponse = await fetch('/api/v1/telegram/config', {
  method: 'POST',
  headers: {
    Authorization: 'Bearer YOUR_TOKEN',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    bot_token: '1234567890:ABCdefGHIjklMNOpqrsTUVwxyz',
    chat_id: '123456789',
    bot_name: 'My Trading Bot',
    is_enabled: true,
  }),
});

// 3. Send test message
const testMsg = await fetch('/api/v1/telegram/test-message', {
  method: 'POST',
  headers: {
    Authorization: 'Bearer YOUR_TOKEN',
  },
});
```

---

## Error Codes

| Code | Description                                     |
| ---- | ----------------------------------------------- |
| 400  | Bad request - Invalid input or failed operation |
| 401  | Unauthorized - Missing or invalid JWT token     |
| 404  | Not found - Configuration does not exist        |
| 500  | Internal server error - Server-side error       |

---

## Notes

- Each user can only have one Telegram configuration
- The bot must be started by the user before it can send messages
- Bot token and chat ID are required fields
- Notifications are only sent when `is_enabled` is `true`
- All messages are sent in HTML format for better formatting
