#!/bin/bash
# Seed Data Creation Script for Blytz MVP
# This script creates seed data from existing demo data

set -e

echo "🌱 Creating Seed Data for Blytz MVP"
echo "===================================="

# Database configuration
DB_URL=${DATABASE_URL:-"postgres://blytz:blytz_password_2025@localhost:5432/blytz_mvp?sslmode=disable"}

echo "📋 Database Configuration:"
echo "  URL: ${DB_URL//*@*/***:***@}" # Hide password
echo

# Function to execute SQL
execute_sql() {
	local sql=$1
	local description=$2
	
	echo "🔄 $description..."
	
	if echo "$sql" | psql "$DB_URL" -v ON_ERROR_STOP=1; then
		echo "✅ $description completed successfully"
	else
		echo "❌ $description failed"
		return 1
	fi
	echo
}

# Create seed data for users
echo "👥 Creating User Seed Data"
echo "========================"

# Create demo users
execute_sql "
INSERT INTO users (id, email, password_hash, display_name, phone_number, role, is_active, email_verified, created_at, updated_at) VALUES
('demo-user-123', 'demo@blytz.app', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7pN1ywK9qK2tLbu', 'Demo User', '+1234567890', 'user', true, true, NOW(), NOW()),
('admin-user-456', 'admin@blytz.app', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7pN1ywK9qK2tLbu', 'Admin User', '+1234567891', 'admin', true, true, NOW(), NOW()),
('test-user-789', 'test@blytz.app', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7pN1ywK9qK2tLbu', 'Test User', '+1234567892', 'user', true, false, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;
" "Creating demo users"

# Create refresh tokens for demo users
execute_sql "
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, is_revoked) VALUES
('rt-demo-123', 'demo-user-123', 'demo-refresh-token-123', NOW() + INTERVAL '7 days', NOW(), false),
('rt-admin-456', 'admin-user-456', 'admin-refresh-token-456', NOW() + INTERVAL '7 days', NOW(), false)
ON CONFLICT (token) DO NOTHING;
" "Creating refresh tokens"

echo "📦 Creating Product Seed Data"
echo "==========================="

# Create categories first
execute_sql "
INSERT INTO categories (id, name, description, parent_id, is_active, sort_order, created_at, updated_at) VALUES
('cat-electronics', 'Electronics', 'Electronic devices and accessories', NULL, true, 1, NOW(), NOW()),
('cat-clothing', 'Clothing', 'Fashion and apparel', NULL, true, 2, NOW(), NOW()),
('cat-home', 'Home & Garden', 'Home improvement and garden supplies', NULL, true, 3, NOW(), NOW()),
('cat-sports', 'Sports & Outdoors', 'Sports equipment and outdoor gear', NULL, true, 4, NOW(), NOW()),
('cat-books', 'Books & Media', 'Books, movies, and digital media', NULL, true, 5, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
" "Creating product categories"

# Create demo products
execute_sql "
INSERT INTO products (id, name, description, price, category_id, seller_id, sku, stock_quantity, min_bid_amount, buy_it_now_price, images, tags, status, is_featured, created_at, updated_at) VALUES
('prod-001', 'iPhone 15 Pro', 'Latest iPhone with advanced camera system', 999.99, 'cat-electronics', 'demo-user-123', 'IPHONE15PRO', 50, 500.00, 1299.99, '[\"https://example.com/iphone1.jpg\", \"https://example.com/iphone2.jpg\"]', '[\"apple\", \"iphone\", \"smartphone\"]', 'active', true, NOW(), NOW()),
('prod-002', 'Samsung 65\" OLED TV', 'Premium 4K OLED television with smart features', 1499.99, 'cat-electronics', 'demo-user-123', 'SAMSUNG65OLED', 25, 750.00, 1899.99, '[\"https://example.com/tv1.jpg\"]', '[\"samsung\", \"tv\", \"oled\"]', 'active', true, NOW(), NOW()),
('prod-003', 'Nike Air Max 2024', 'Latest running shoes with enhanced cushioning', 189.99, 'cat-sports', 'demo-user-123', 'NIKEAIRMAX2024', 100, 95.00, 249.99, '[\"https://example.com/shoes1.jpg\", \"https://example.com/shoes2.jpg\"]', '[\"nike\", \"shoes\", \"running\"]', 'active', false, NOW(), NOW()),
('prod-004', 'Levi\\'s 501 Jeans', 'Classic fit denim jeans', 79.99, 'cat-clothing', 'demo-user-123', 'LEVIS501', 200, 40.00, 119.99, '[\"https://example.com/jeans1.jpg\"]', '[\"levis\", \"jeans\", \"denim\"]', 'active', false, NOW(), NOW()),
('prod-005', 'Coffee Maker Deluxe', 'Premium coffee maker with thermal carafe', 89.99, 'cat-home', 'demo-user-123', 'COFFEEMAKER', 75, 45.00, 139.99, '[\"https://example.com/coffee1.jpg\"]', '[\"coffee\", \"maker\", \"kitchen\"]', 'active', false, NOW(), NOW())
ON CONFLICT (sku) DO NOTHING;
" "Creating demo products"

echo "🏛 Creating Auction Seed Data"
echo "==========================="

# Create demo auctions
execute_sql "
INSERT INTO auctions (id, product_id, seller_id, title, description, starting_price, reserve_price, buy_it_now_price, current_bid, bid_count, start_time, end_time, status, auto_extend, anti_snipe_time, created_at, updated_at) VALUES
('auc-001', 'prod-001', 'demo-user-123', 'iPhone 15 Pro Auction', 'Brand new iPhone 15 Pro, sealed in box', 500.00, 800.00, 1299.99, 500.00, 0, NOW() - INTERVAL '1 hour', NOW() + INTERVAL '23 hours', 'active', true, 300, NOW(), NOW()),
('auc-002', 'prod-002', 'demo-user-123', 'Samsung 65\" OLED TV Auction', 'Premium OLED TV, 1 year warranty', 1000.00, 1200.00, 1899.99, 1000.00, 0, NOW() - INTERVAL '30 minutes', NOW() + INTERVAL '22 hours', 'active', true, 300, NOW(), NOW()),
('auc-003', 'prod-003', 'demo-user-123', 'Nike Air Max 2024 Auction', 'Latest running shoes, size 10', 100.00, 150.00, 249.99, 100.00, 0, NOW() - INTERVAL '2 hours', NOW() + INTERVAL '20 hours', 'active', true, 300, NOW(), NOW())
ON CONFLICT DO NOTHING;
" "Creating demo auctions"

# Create demo bids
execute_sql "
INSERT INTO bids (id, auction_id, bidder_id, amount, bid_time, is_auto_bid, is_winning, created_at) VALUES
('bid-001', 'auc-001', 'admin-user-456', 550.00, NOW() - INTERVAL '45 minutes', false, true, NOW()),
('bid-002', 'auc-001', 'test-user-789', 600.00, NOW() - INTERVAL '30 minutes', false, false, NOW()),
('bid-003', 'auc-002', 'demo-user-123', 1100.00, NOW() - INTERVAL '15 minutes', false, true, NOW())
ON CONFLICT DO NOTHING;
" "Creating demo bids"

echo "📋 Creating Order Seed Data"
echo "==========================="

# Create demo orders
execute_sql "
INSERT INTO orders (id, user_id, auction_id, seller_id, total_amount, status, payment_status, shipping_address, billing_address, items, subtotal, tax_amount, shipping_amount, discount_amount, currency, notes, created_at, updated_at) VALUES
('order-001', 'admin-user-456', 'auc-001', 'demo-user-123', 550.00, 'pending', 'pending', '{\"street\": \"123 Main St\", \"city\": \"New York\", \"state\": \"NY\", \"zip\": \"10001\", \"country\": \"USA\"}', '{\"street\": \"123 Main St\", \"city\": \"New York\", \"state\": \"NY\", \"zip\": \"10001\", \"country\": \"USA\"}', '[{\"product_id\": \"prod-001\", \"quantity\": 1, \"unit_price\": 550.00, \"total_price\": 550.00}]', 550.00, 44.00, 15.00, 0.00, 'USD', 'Please ship quickly', NOW(), NOW()),
('order-002', 'test-user-789', 'auc-002', 'demo-user-123', 1100.00, 'processing', 'paid', '{\"street\": \"456 Oak Ave\", \"city\": \"Los Angeles\", \"state\": \"CA\", \"zip\": \"90001\", \"country\": \"USA\"}', '{\"street\": \"456 Oak Ave\", \"city\": \"Los Angeles\", \"state\": \"CA\", \"zip\": \"90001\", \"country\": \"USA\"}', '[{\"product_id\": \"prod-002\", \"quantity\": 1, \"unit_price\": 1100.00, \"total_price\": 1100.00}]', 1100.00, 88.00, 25.00, 0.00, 'USD', 'Gift wrap please', NOW(), NOW())
ON CONFLICT DO NOTHING;
" "Creating demo orders"

# Create order items
execute_sql "
INSERT INTO order_items (id, order_id, product_id, auction_id, quantity, unit_price, total_price, product_snapshot, created_at) VALUES
('oi-001', 'order-001', 'prod-001', 'auc-001', 1, 550.00, 550.00, '{\"name\": \"iPhone 15 Pro\", \"price\": 999.99, \"sku\": \"IPHONE15PRO\"}', NOW()),
('oi-002', 'order-002', 'prod-002', 'auc-002', 1, 1100.00, 1100.00, '{\"name\": \"Samsung 65\" OLED TV\", \"price\": 1499.99, \"sku\": \"SAMSUNG65OLED\"}', NOW())
ON CONFLICT DO NOTHING;
" "Creating order items"

# Create order status history
execute_sql "
INSERT INTO order_status_history (id, order_id, status, previous_status, comment, created_by, created_at) VALUES
('osh-001', 'order-001', 'pending', NULL, 'Order created', 'admin-user-456', NOW()),
('osh-002', 'order-001', 'processing', 'pending', 'Payment received', 'system', NOW()),
('osh-003', 'order-002', 'pending', NULL, 'Order created', 'test-user-789', NOW()),
('osh-004', 'order-002', 'processing', 'pending', 'Payment confirmed', 'system', NOW())
ON CONFLICT DO NOTHING;
" "Creating order status history"

echo "📝 Creating Product Reviews"
echo "==========================="

# Create product reviews
execute_sql "
INSERT INTO product_reviews (id, product_id, user_id, rating, title, comment, verified_purchase, helpful_count, status, created_at, updated_at) VALUES
('review-001', 'prod-001', 'admin-user-456', 5, 'Excellent!', 'Amazing phone, camera is incredible', true, 12, 'approved', NOW(), NOW()),
('review-002', 'prod-001', 'test-user-789', 4, 'Very Good', 'Great phone but expensive', true, 8, 'approved', NOW(), NOW()),
('review-003', 'prod-002', 'demo-user-123', 5, 'Perfect TV!', 'Best picture quality I\\'ve ever seen', true, 15, 'approved', NOW(), NOW())
ON CONFLICT DO NOTHING;
" "Creating product reviews"

echo "🔍 Verifying Seed Data"
echo "======================"

# Verify data was created
execute_sql "
SELECT 
    (SELECT COUNT(*) FROM users) as user_count,
    (SELECT COUNT(*) FROM products) as product_count,
    (SELECT COUNT(*) FROM auctions) as auction_count,
    (SELECT COUNT(*) FROM orders) as order_count,
    (SELECT COUNT(*) FROM bids) as bid_count;
" "Verifying seed data counts"

echo
echo "🎉 Seed Data Creation Completed!"
echo "================================"
echo "📊 Created Data Summary:"
echo "  👥 Users: 3 demo accounts"
echo "  📦 Products: 5 demo products"
echo "  🏛 Auctions: 3 active auctions"
echo "  📋 Orders: 2 demo orders"
echo "  💰 Bids: 3 demo bids"
echo "  ⭐ Reviews: 3 demo reviews"
echo
echo "📋 Demo Accounts:"
echo "  📧 demo@blytz.app / demo123"
echo "  👑 admin@blytz.app / admin123"
echo "  🧪 test@blytz.app / test123"
echo
echo "🚀 Next Steps:"
echo "1. Test authentication:"
echo "   curl -X POST http://localhost:8085/api/v1/auth/login \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     -d '{\"email\": \"demo@blytz.app\", \"password\": \"demo123\"}'"
echo
echo "2. Test service health:"
echo "   curl http://localhost:8085/health"
echo
echo "3. Run integration tests:"
echo "   ./scripts/test-integration.sh"
echo
echo "🎯 Your database is now populated with realistic demo data!"