#!/usr/bin/env python3
import hmac
import hashlib
import time
import requests
from urllib.parse import urlencode

# Binance Testnet credentials
API_KEY = "CfJsnKKOqXKzQBXca8Wii6rBW9sCSmSaK9Skn0JGG6ooAdaUSSMgMGbudTa6dnwz"
API_SECRET = "bqQBmHfL0qKjUd8Vj7Y1GpLfA6RVMNq8eoLtHO0Fu6PLwNv4n2X19uzWaJsBbJH9"
BASE_URL = "https://testnet.binance.vision"

def get_signature(params, secret):
    """Generate HMAC SHA256 signature"""
    query_string = urlencode(params)
    signature = hmac.new(
        secret.encode('utf-8'),
        query_string.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    return signature

def get_account_info():
    """Get account information from Binance Testnet"""
    endpoint = "/api/v3/account"
    
    # Prepare parameters
    params = {
        'timestamp': int(time.time() * 1000)
    }
    
    # Generate signature
    params['signature'] = get_signature(params, API_SECRET)
    
    # Make request
    headers = {
        'X-MBX-APIKEY': API_KEY
    }
    
    url = BASE_URL + endpoint
    response = requests.get(url, params=params, headers=headers)
    
    if response.status_code == 200:
        data = response.json()
        print("✅ Account Info Retrieved Successfully!\n")
        print("=" * 60)
        print(f"Can Trade: {data.get('canTrade')}")
        print(f"Can Withdraw: {data.get('canWithdraw')}")
        print(f"Can Deposit: {data.get('canDeposit')}")
        print("=" * 60)
        print("\n💰 BALANCES:\n")
        
        balances = data.get('balances', [])
        has_balance = False
        
        for balance in balances:
            free = float(balance['free'])
            locked = float(balance['locked'])
            total = free + locked
            
            if total > 0:
                has_balance = True
                print(f"  {balance['asset']:8s}: {free:15.8f} (free) + {locked:15.8f} (locked) = {total:15.8f}")
        
        if not has_balance:
            print("  ❌ NO BALANCE FOUND!")
            print("\n  ⚠️  You need to get test funds from:")
            print("  📝 Method 1: Login to https://testnet.binance.vision/")
            print("                and use the Faucet feature")
            print("  📝 Method 2: Contact testnet@binance.com")
            print("  📝 Method 3: Switch to Futures Testnet:")
            print("                https://testnet.binancefuture.com/")
            print("                (Has automatic faucet!)")
        
        print("\n" + "=" * 60)
        
    else:
        print(f"❌ Error: {response.status_code}")
        print(response.text)

if __name__ == "__main__":
    print("🔍 Checking Binance Testnet Account Balance...\n")
    get_account_info()
