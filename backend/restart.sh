#!/bin/bash

# Kill old backend
echo "🔴 Killing old backend..."
lsof -ti:8080 | xargs kill -9 2>/dev/null
sleep 1

# Start new backend
echo "🚀 Starting new backend..."
cd /Users/khaicafe/Develop/TraderCoin/Backend
./backend

