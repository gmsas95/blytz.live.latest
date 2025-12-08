-- Database Schema Verification Script
-- This script checks if our production database schema is properly set up

\c blytzwork

-- Check if UUID extension is available
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Show current database tables
\dt

-- Show table structures for core tables
\d users
\d products
\d auctions
\d orders
\d bids
\d payments

-- Check indexes
\di

-- Sample data queries
SELECT 'Users table count' as table_name, COUNT(*) as count FROM users;
SELECT 'Products table count' as table_name, COUNT(*) as count FROM products;
SELECT 'Auctions table count' as table_name, COUNT(*) as count FROM auctions;
SELECT 'Orders table count' as table_name, COUNT(*) as count FROM orders;

-- Show sample users (if any)
SELECT id, email, name, role, created_at 
FROM users 
LIMIT 5;

-- Show database version
SELECT version();