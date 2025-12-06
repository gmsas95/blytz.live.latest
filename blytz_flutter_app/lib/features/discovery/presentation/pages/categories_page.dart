import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class CategoriesPage extends StatefulWidget {
  const CategoriesPage({super.key});

  @override
  State<CategoriesPage> createState() => _CategoriesPageState();
}

class _CategoriesPageState extends State<CategoriesPage>
    with TickerProviderStateMixin {
  late TabController _tabController;
  final TextEditingController _searchController = TextEditingController();
  String _searchQuery = '';

  final List<Map<String, dynamic>> _categories = [
    {
      'name': 'Electronics',
      'icon': Icons.devices_outlined,
      'color': Colors.blue,
      'description': 'Gadgets, computers, smartphones, and more',
      'streamCount': 234,
      'trending': true,
    },
    {
      'name': 'Fashion',
      'icon': Icons.checkroom_outlined,
      'color': Colors.pink,
      'description': 'Clothing, accessories, designer items',
      'streamCount': 456,
      'trending': true,
    },
    {
      'name': 'Collectibles',
      'icon': Icons.collections_outlined,
      'color': Colors.amber,
      'description': 'Rare items, memorabilia, unique finds',
      'streamCount': 189,
      'trending': false,
    },
    {
      'name': 'Home & Garden',
      'icon': Icons.home_outlined,
      'color': Colors.green,
      'description': 'Furniture, decor, outdoor equipment',
      'streamCount': 167,
      'trending': false,
    },
    {
      'name': 'Sports',
      'icon': Icons.sports_soccer_outlined,
      'color': Colors.orange,
      'description': 'Sporting goods, fitness equipment',
      'streamCount': 123,
      'trending': false,
    },
    {
      'name': 'Toys & Games',
      'icon': Icons.toys_outlined,
      'color': Colors.purple,
      'description': 'Collectible toys, board games, video games',
      'streamCount': 98,
      'trending': false,
    },
    {
      'name': 'Books',
      'icon': Icons.book_outlined,
      'color': Colors.brown,
      'description': 'Rare books, comics, literature',
      'streamCount': 76,
      'trending': false,
    },
    {
      'name': 'Art',
      'icon': Icons.palette_outlined,
      'color': Colors.indigo,
      'description': 'Paintings, sculptures, digital art',
      'streamCount': 145,
      'trending': true,
    },
    {
      'name': 'Jewelry',
      'icon': Icons.diamond_outlined,
      'color': Colors.teal,
      'description': 'Fine jewelry, watches, accessories',
      'streamCount': 89,
      'trending': false,
    },
    {
      'name': 'Vintage',
      'icon': Icons.watch_outlined,
      'color': Colors.grey,
      'description': 'Antiques, retro items, historical pieces',
      'streamCount': 234,
      'trending': true,
    },
    {
      'name': 'Automotive',
      'icon': Icons.directions_car_outlined,
      'color': Colors.red,
      'description': 'Car parts, accessories, collectible vehicles',
      'streamCount': 67,
      'trending': false,
    },
    {
      'name': 'Musical Instruments',
      'icon': Icons.music_note_outlined,
      'color': Colors.deepPurple,
      'description': 'Guitars, keyboards, vintage instruments',
      'streamCount': 112,
      'trending': false,
    },
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    
    return Scaffold(
      backgroundColor: colorScheme.surface,
      body: CustomScrollView(
        slivers: [
          // Modern Glass Header
          SliverAppBar(
            expandedHeight: 240,
            floating: false,
            pinned: true,
            backgroundColor: Colors.transparent,
            leading: IconButton(
              icon: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: Colors.white.withOpacity(0.3),
                    width: 1,
                  ),
                ),
                child: Icon(
                  Icons.arrow_back_ios_new,
                  color: colorScheme.onPrimary,
                  size: 18,
                ),
              ),
              onPressed: () => Navigator.of(context).pop(),
            ),
            actions: [
              IconButton(
                icon: Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: Colors.white.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: Colors.white.withOpacity(0.3),
                      width: 1,
                    ),
                  ),
                  child: Icon(
                    Icons.filter_list_outlined,
                    color: colorScheme.onPrimary,
                    size: 20,
                  ),
                ),
                onPressed: _showFilterBottomSheet,
              ),
              const SizedBox(width: 8),
            ],
            flexibleSpace: FlexibleSpaceBar(
              background: Stack(
                children: [
                  // Gradient Background
                  Container(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          colorScheme.primary,
                          colorScheme.primary.withOpacity(0.8),
                          colorScheme.secondary,
                        ],
                      ),
                    ),
                  ),
                  // Glass Pattern Overlay
                  Positioned.fill(
                    child: Container(
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.1),
                        borderRadius: const BorderRadius.only(
                          bottomLeft: Radius.circular(30),
                          bottomRight: Radius.circular(30),
                        ),
                      ),
                    ),
                  ),
                  // Content
                  SafeArea(
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(24, 80, 24, 24),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Categories',
                            style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                              color: colorScheme.onPrimary,
                              fontWeight: FontWeight.bold,
                              fontSize: 32,
                            ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Explore our curated collection of categories',
                            style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                              color: colorScheme.onPrimary.withOpacity(0.9),
                            ),
                          ),
                          const SizedBox(height: 20),
                          
                          // Modern Search Bar
                          SearchBar(
                            controller: _searchController,
                            onChanged: (value) {
                              setState(() {
                                _searchQuery = value;
                              });
                            },
                            hintText: 'Search categories...',
                            hintStyle: WidgetStatePropertyAll(
                              TextStyle(
                                color: colorScheme.onPrimary.withOpacity(0.7),
                              ),
                            ),
                            leading: Icon(
                              Icons.search_outlined,
                              color: colorScheme.onPrimary.withOpacity(0.7),
                            ),
                            backgroundColor: WidgetStatePropertyAll(
                              Colors.white.withOpacity(0.2),
                            ),
                            side: WidgetStatePropertyAll(
                              BorderSide(
                                color: Colors.white.withOpacity(0.3),
                                width: 1,
                              ),
                            ),
                            shape: WidgetStatePropertyAll(
                              RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(16),
                              ),
                            ),
                            elevation: const WidgetStatePropertyAll(0),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          
          // Tab Bar
          SliverPersistentHeader(
            delegate: _SliverTabBarDelegate(
              TabBar(
                controller: _tabController,
                indicatorColor: colorScheme.primary,
                labelColor: colorScheme.primary,
                unselectedLabelColor: colorScheme.onSurface.withOpacity(0.6),
                indicatorWeight: 3,
                indicatorSize: TabBarIndicatorSize.tab,
                tabs: const [
                  Tab(text: 'All', icon: Icon(Icons.grid_view_outlined)),
                  Tab(text: 'Trending', icon: Icon(Icons.trending_up_outlined)),
                  Tab(text: 'Featured', icon: Icon(Icons.star_outline)),
                ],
              ),
            ),
            pinned: true,
          ),
          
          // Tab Content
          SliverFillRemaining(
            child: TabBarView(
              controller: _tabController,
              children: [
                _buildCategoriesGrid('all'),
                _buildCategoriesGrid('trending'),
                _buildFeaturedCategories(),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCategoriesGrid(String filter) {
    List<Map<String, dynamic>> filteredCategories = _categories;
    
    if (_searchQuery.isNotEmpty) {
      filteredCategories = _categories
          .where((category) => category['name']
              .toString()
              .toLowerCase()
              .contains(_searchQuery.toLowerCase()))
          .toList();
    }
    
    if (filter == 'trending') {
      filteredCategories = filteredCategories
          .where((category) => category['trending'] == true)
          .toList();
    }
    
    return Padding(
      padding: const EdgeInsets.all(16),
      child: GridView.builder(
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 2,
          childAspectRatio: 0.85,
          crossAxisSpacing: 16,
          mainAxisSpacing: 16,
        ),
        itemCount: filteredCategories.length,
        itemBuilder: (context, index) {
          final category = filteredCategories[index];
          return _buildCategoryCard(category, index);
        },
      ),
    );
  }

  Widget _buildCategoryCard(Map<String, dynamic> category, int index) {
    final colorScheme = Theme.of(context).colorScheme;
    
    return Container(
      decoration: BoxDecoration(
        color: colorScheme.surface,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.08),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: () {
              // TODO: Navigate to category details
            },
            borderRadius: BorderRadius.circular(20),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Icon Container
                  Container(
                    width: double.infinity,
                    height: 80,
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          (category['color'] as Color?)?.withOpacity(0.8) ?? Colors.blue.withOpacity(0.8),
                          (category['color'] as Color?)?.withOpacity(0.6) ?? Colors.blue.withOpacity(0.6),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Stack(
                      children: [
                        Center(
                          child: Icon(
                            category['icon'] as IconData? ?? Icons.category,
                            size: 40,
                            color: Colors.white,
                          ),
                        ),
                        if (category['trending'] == true)
                          Positioned(
                            top: 8,
                            right: 8,
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 6,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.red,
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: const Text(
                                'HOT',
                                style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 8,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ),
                      ],
                    ),
                  ),
                  
                  const SizedBox(height: 12),
                  
                  // Category Name
                  Text(
                    category['name'] as String? ?? 'Category',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  
                  const SizedBox(height: 4),
                  
                  // Description
                  Text(
                    category['description'] as String? ?? 'Category description',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: colorScheme.onSurface.withOpacity(0.6),
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  
                  const Spacer(),
                  
                  // Stream Count
                  Row(
                    children: [
                      Icon(
                        Icons.live_tv_outlined,
                        size: 14,
                        color: colorScheme.primary,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${category['streamCount']} live',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: colorScheme.primary,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const Spacer(),
                      Icon(
                        Icons.arrow_forward_ios,
                        size: 12,
                        color: colorScheme.onSurface.withOpacity(0.4),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFeaturedCategories() {
    final featuredCategories = _categories.take(6).toList();
    
    return Padding(
      padding: const EdgeInsets.all(16),
      child: ListView.builder(
        itemCount: featuredCategories.length,
        itemBuilder: (context, index) {
          final category = featuredCategories[index];
          return _buildFeaturedCategoryCard(category, index);
        },
      ),
    );
  }

  Widget _buildFeaturedCategoryCard(Map<String, dynamic> category, int index) {
    final colorScheme = Theme.of(context).colorScheme;
    
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: colorScheme.surface,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.08),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: () {
              // TODO: Navigate to category details
            },
            borderRadius: BorderRadius.circular(20),
            child: Row(
              children: [
                // Icon Container
                Container(
                  width: 100,
                  height: 100,
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        (category['color'] as Color?)?.withOpacity(0.8) ?? Colors.blue.withOpacity(0.8),
                        (category['color'] as Color?)?.withOpacity(0.6) ?? Colors.blue.withOpacity(0.6),
                      ],
                    ),
                  ),
                  child: Stack(
                    children: [
                      Center(
                        child: Icon(
                          category['icon'] as IconData? ?? Icons.category,
                          size: 40,
                          color: Colors.white,
                        ),
                      ),
                      if (category['trending'] == true)
                        Positioned(
                          top: 8,
                          right: 8,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 6,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: Colors.red,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Text(
                              'HOT',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 8,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
                
                // Content
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          category['name'] as String? ?? 'Category',
                          style: Theme.of(context).textTheme.titleLarge?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          category['description'] as String? ?? 'Category description',
                          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: colorScheme.onSurface.withOpacity(0.7),
                          ),
                        ),
                        const SizedBox(height: 12),
                        Row(
                          children: [
                            Icon(
                              Icons.live_tv_outlined,
                              size: 16,
                              color: colorScheme.primary,
                            ),
                            const SizedBox(width: 4),
                            Text(
                              '${category['streamCount']} live streams',
                              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: colorScheme.primary,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const Spacer(),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 6,
                              ),
                              decoration: BoxDecoration(
                                color: colorScheme.primary.withOpacity(0.1),
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                'Explore',
                                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                  color: colorScheme.primary,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _showFilterBottomSheet() {
    final colorScheme = Theme.of(context).colorScheme;
    
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.5,
        decoration: BoxDecoration(
          color: colorScheme.surface,
          borderRadius: const BorderRadius.only(
            topLeft: Radius.circular(30),
            topRight: Radius.circular(30),
          ),
        ),
        child: Column(
          children: [
            Container(
              width: 40,
              height: 4,
              margin: const EdgeInsets.symmetric(vertical: 12),
              decoration: BoxDecoration(
                color: colorScheme.onSurface.withOpacity(0.2),
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Filter Categories',
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 24),
                  // TODO: Add filter options
                  const Expanded(
                    child: Center(
                      child: Text('Filter options coming soon...'),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _SliverTabBarDelegate extends SliverPersistentHeaderDelegate {
  final TabBar _tabBar;

  _SliverTabBarDelegate(this._tabBar);

  @override
  double get minExtent => _tabBar.preferredSize.height;

  @override
  double get maxExtent => _tabBar.preferredSize.height;

  @override
  Widget build(
    BuildContext context,
    double shrinkOffset,
    bool overlapsContent,
  ) {
    return Container(
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: _tabBar,
    );
  }

  @override
  bool shouldRebuild(_SliverTabBarDelegate oldDelegate) {
    return false;
  }
}