import 'package:flutter/material.dart';

class OrdersPage extends StatefulWidget {
  const OrdersPage({super.key});

  @override
  State<OrdersPage> createState() => _OrdersPageState();
}

class _OrdersPageState extends State<OrdersPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 200,
            floating: false,
            pinned: true,
            backgroundColor: Colors.transparent,
            flexibleSpace: FlexibleSpaceBar(
              background: Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [
                      Colors.purple.shade400,
                      Colors.blue.shade600,
                    ],
                  ),
                ),
                child: SafeArea(
                  child: Padding(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'My Orders',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 28,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 20),
                        Row(
                          children: [
                            _buildStatCard('Active', '3', Colors.orange),
                            const SizedBox(width: 15),
                            _buildStatCard('Completed', '24', Colors.green),
                            const SizedBox(width: 15),
                            _buildStatCard('Total', '27', Colors.blue),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
            bottom: TabBar(
              controller: _tabController,
              indicatorColor: Colors.white,
              labelColor: Colors.white,
              unselectedLabelColor: Colors.white70,
              tabs: const [
                Tab(text: 'Active'),
                Tab(text: 'Processing'),
                Tab(text: 'Completed'),
                Tab(text: 'Cancelled'),
              ],
            ),
          ),
          SliverFillRemaining(
            child: TabBarView(
              controller: _tabController,
              children: [
                _buildOrdersList('active'),
                _buildOrdersList('processing'),
                _buildOrdersList('completed'),
                _buildOrdersList('cancelled'),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatCard(String label, String value, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.2),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withOpacity(0.3)),
      ),
      child: Column(
        children: [
          Text(
            value,
            style: TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          Text(
            label,
            style: TextStyle(
              color: Colors.white.withOpacity(0.9),
              fontSize: 12,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildOrdersList(String status) {
    final orders = _getOrdersForStatus(status);
    
    if (orders.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.shopping_bag_outlined,
              size: 80,
              color: Colors.grey.shade400,
            ),
            const SizedBox(height: 16),
            Text(
              'No ${status} orders',
              style: TextStyle(
                fontSize: 18,
                color: Colors.grey.shade600,
                fontWeight: FontWeight.w500,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Your ${status} orders will appear here',
              style: TextStyle(
                fontSize: 14,
                color: Colors.grey.shade500,
              ),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: orders.length,
      itemBuilder: (context, index) {
        final order = orders[index];
        return Container(
          margin: const EdgeInsets.only(bottom: 16),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.05),
                blurRadius: 10,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              borderRadius: BorderRadius.circular(16),
              onTap: () {
                // Navigate to order details
              },
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Order #${order['id']}',
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: _getStatusColor(order['status']?.toString() ?? '').withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Text(
                            order['status']?.toString().toUpperCase() ?? '',
                            style: TextStyle(
                              color: _getStatusColor(order['status']?.toString() ?? ''),
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        ClipRRect(
                          borderRadius: BorderRadius.circular(8),
                          child: Container(
                            width: 60,
                            height: 60,
                            color: Colors.grey.shade200,
                            child: Icon(
                              Icons.image,
                              color: Colors.grey.shade400,
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                order['itemName']?.toString() ?? '',
                                style: const TextStyle(
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600,
                                ),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                              const SizedBox(height: 4),
                              Text(
                                'Quantity: ${order['quantity']}',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: Colors.grey.shade600,
                                ),
                              ),
                              Text(
                                '\$${order['price']}',
                                style: const TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.blue,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Placed on ${order['date']}',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade500,
                          ),
                        ),
                        Row(
                          children: [
                            if (order['status'] == 'active') ...[
                              TextButton(
                                onPressed: () {
                                  // Track order
                                },
                                child: const Text('Track'),
                              ),
                              TextButton(
                                onPressed: () {
                                  // Cancel order
                                },
                                style: TextButton.styleFrom(
                                  foregroundColor: Colors.red,
                                ),
                                child: const Text('Cancel'),
                              ),
                            ] else if (order['status'] == 'completed') ...[
                              TextButton(
                                onPressed: () {
                                  // Buy again
                                },
                                child: const Text('Buy Again'),
                              ),
                              TextButton(
                                onPressed: () {
                                  // Review
                                },
                                child: const Text('Review'),
                              ),
                            ] else if (order['status'] == 'processing') ...[
                              TextButton(
                                onPressed: () {
                                  // View details
                                },
                                child: const Text('Details'),
                              ),
                            ],
                          ],
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'active':
        return Colors.orange;
      case 'processing':
        return Colors.blue;
      case 'completed':
        return Colors.green;
      case 'cancelled':
        return Colors.red;
      default:
        return Colors.grey;
    }
  }

  List<Map<String, dynamic>> _getOrdersForStatus(String status) {
    // Mock data - in real app, this would come from API
    switch (status) {
      case 'active':
        return [
          {
            'id': 'ORD001',
            'itemName': 'Vintage Camera',
            'quantity': 1,
            'price': '450.00',
            'status': 'active',
            'date': '2024-01-15',
          },
          {
            'id': 'ORD002',
            'itemName': 'Designer Watch',
            'quantity': 1,
            'price': '1,200.00',
            'status': 'active',
            'date': '2024-01-14',
          },
          {
            'id': 'ORD003',
            'itemName': 'Art Painting',
            'quantity': 2,
            'price': '800.00',
            'status': 'active',
            'date': '2024-01-13',
          },
        ];
      case 'processing':
        return [
          {
            'id': 'ORD004',
            'itemName': 'Rare Book Collection',
            'quantity': 3,
            'price': '350.00',
            'status': 'processing',
            'date': '2024-01-12',
          },
        ];
      case 'completed':
        return [
          {
            'id': 'ORD005',
            'itemName': 'Antique Vase',
            'quantity': 1,
            'price': '650.00',
            'status': 'completed',
            'date': '2024-01-10',
          },
          {
            'id': 'ORD006',
            'itemName': 'Jewelry Set',
            'quantity': 1,
            'price': '2,100.00',
            'status': 'completed',
            'date': '2024-01-08',
          },
        ];
      case 'cancelled':
        return [
          {
            'id': 'ORD007',
            'itemName': 'Electronic Gadget',
            'quantity': 1,
            'price': '180.00',
            'status': 'cancelled',
            'date': '2024-01-05',
          },
        ];
      default:
        return [];
    }
  }
}