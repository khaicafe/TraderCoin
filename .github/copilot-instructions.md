# TraderCoin Project Instructions

## Project Overview

TraderCoin is a comprehensive cryptocurrency automated trading platform with three main components:

1. **Frontend** (Next.js) - User-facing application
2. **Backend** (Golang) - API and trading engine
3. **Backoffice** (Next.js) - Admin management system

## Architecture

- Monorepo structure with separate services
- Backend: Golang with Gin framework, PostgreSQL, Redis
- Frontend/Backoffice: Next.js 14 with TypeScript and Tailwind CSS
- Docker containers for all services
- RESTful API with JWT authentication
- WebSocket for real-time updates

## Development Guidelines

- Use TypeScript for all Next.js code
- Follow Go best practices and error handling
- Implement proper validation and security measures
- Use environment variables for configuration
- Write clean, maintainable code with comments
- Follow RESTful API conventions

## Key Features

### Frontend

- User authentication and profile management
- Exchange API key configuration (Binance, Bittrex)
- Automated trading setup (stop-loss, take-profit)
- Real-time portfolio tracking
- Subscription management

### Backend

- User authentication with JWT
- Exchange API integration
- Automated trading engine
- Order monitoring and execution
- Database operations
- WebSocket server

### Backoffice

- Admin authentication
- User management (suspend, activate)
- Subscription and billing management
- Transaction history
- Analytics and reports
