-- Migration to add margin_mode and callback_rate fields to trading_configs table
-- Run this migration using: psql -U your_user -d tradercoin_db -f migrations/add_margin_callback_fields.sql

-- Add margin_mode column (ISOLATED or CROSSED for futures trading)
ALTER TABLE trading_configs 
ADD COLUMN IF NOT EXISTS margin_mode VARCHAR(20) DEFAULT 'ISOLATED';

-- Add callback_rate column (for trailing stop, 0.1-5%)
ALTER TABLE trading_configs 
ADD COLUMN IF NOT EXISTS callback_rate DECIMAL(10,2) DEFAULT 1.0;

-- Add comment to document the fields
COMMENT ON COLUMN trading_configs.margin_mode IS 'Margin mode for futures trading: ISOLATED or CROSSED';
COMMENT ON COLUMN trading_configs.callback_rate IS 'Callback rate for trailing stop (0.1-5%)';

-- Show the updated table structure
\d trading_configs;
