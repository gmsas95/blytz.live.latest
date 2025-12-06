import 'package:blytz_flutter_app/core/constants/app_constants.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class DiscoveryPage extends StatefulWidget {
  const DiscoveryPage({super.key});

  @override
  State<DiscoveryPage> createState() => _DiscoveryPageState();
}

class _DiscoveryPageState extends State<DiscoveryPage>
    with TickerProviderStateMixin {
  late TabController _tabController;
  late PageController _pageController;
  String _selectedCategory = 'All';
  final TextEditingController _searchController = TextEditingController();
  final ScrollController _scrollController = ScrollController();

  final List<String> _categories = [
    'All', 'Electronics', 'Fashion', 'Collectibles', 
    'Home & Garden', 'Sports', 'Toys & Games', 'Books', 
    'Art', 'Jewelry', 'Vintage',
  ];

  final List<Map<String, dynamic>> _trendingStreams = [
    {
      'title': 'Vintage Electronics Auction',
      'seller': 'TechCollector',
      'viewers': 1234,
      'category': 'Electronics',
      'isLive': true,
      'rating': 4.8,
      'bids': 45,
    },
    {
      'title': 'Designer Fashion Showcase',
      'seller': 'StyleGuru',
      'viewers': 856,
      'category': 'Fashion',
      'isLive': true,
      'rating': 4.9,
      'bids': 67,
    },
    {
      'title': 'Rare Coins & Stamps',
      'seller': 'HistoryBuff',
      'viewers': 445,
      'category': 'Collectibles',
      'isLive': true,
      'rating': 4.7,
      'bids': 23,
    },
    {
      'title': 'Modern Art Gallery',
      'seller': 'ArtistLife',
      'viewers': 678,
      'category': 'Art',
      'isLive': true,
      'rating': 4.9,
      'bids': 89,
    },
    {
      'title': 'Sports Memorabilia',
      'seller': 'SportsFan',
      'viewers': 334,
      'category': 'Sports',
      'isLive': true,
      'rating': 4.6,
      'bids': 34,
    },
    {
      'title': 'Luxury Watches',
      'seller': 'WatchExpert',
      'viewers': 923,
      'category': 'Jewelry',
      'isLive': true,
      'rating': 4.8,
      'bids': 56,
    },
  ];

  final List<Map<String, dynamic>> _upcomingStreams = [
    {
      'title': 'Antique Furniture Sale',
      'seller': 'FurniturePro',
      'startTime': '2:00 PM',
      'category': 'Home & Garden',
      'thumbnail': 'assets/images/furniture.jpg',
    },
    {
      'title': 'Vintage Comics Collection',
      'seller': 'ComicCollector',
      'startTime': '4:30 PM',
      'category': 'Collectibles',
      'thumbnail': 'assets/images/comics.jpg',
    },
    {
      'title': 'Designer Handbags',
      'seller': 'Fashionista',
      'startTime': '6:00 PM',
      'category': 'Fashion',
      'thumbnail': 'assets/images/handbags.jpg',
    },
  ];

  final List<String> _trendingTags = [
    '#vintage', '#electronics', '#fashion', '#rare', 
    '#collectibles', '#limited', '#handmade', '#designer',
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _pageController = PageController();
  }

  @override
  void dispose() {
    _tabController.dispose();
    _pageController.dispose();
    _searchController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    
    return Scaffold(
      backgroundColor: colorScheme.surface,
      body: NestedScrollView(
        controller: _scrollController,
        headerSliverBuilder: (context, innerBoxIsScrolled) {
          return [
            // Modern Glass Header
            SliverAppBar(
              expandedHeight: 280,
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
                      Icons.notifications_outlined,
                      color: colorScheme.onPrimary,
                      size: 20,
                    ),
                  ),
                  onPressed: () {
                    // TODO: Show notifications
                  },
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
                              'Discover',
                              style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                                color: colorScheme.onPrimary,
                                fontWeight: FontWeight.bold,
                                fontSize: 32,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              'Find amazing live auctions and exclusive items',
                              style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                                color: colorScheme.onPrimary.withOpacity(0.9),
                              ),
                            ),
                            const SizedBox(height: AppConstants.spacing20),
                            
                            // Modern Search Bar
                            Semantics(
                              label: 'Search streams, items, and sellers',
                              textField: true,
                              child: SearchBar(
                                controller: _searchController,
                                hintText: 'Search streams, items, sellers...',
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
                                    borderRadius: BorderRadius.circular(AppConstants.radius16),
                                  ),
                                ),
                                elevation: const WidgetStatePropertyAll(0),
                              ),
                            ),
                            
                            const SizedBox(height: AppConstants.spacing16),
                            
                            // Quick Stats
                            Row(
                              children: [
                                _buildQuickStat('Live Now', '24', Icons.live_tv),
                                const SizedBox(width: AppConstants.spacing12),
                                _buildQuickStat('Trending', '156', Icons.trending_up),
                                const SizedBox(width: AppConstants.spacing12),
                                _buildQuickStat('New Today', '89', Icons.new_releases),
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
            
            // Category Pills
            SliverToBoxAdapter(
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: AppConstants.spacing16, vertical: AppConstants.spacing8),
                child: Semantics(
                  label: 'Category filter: $_selectedCategory',
                  child: SegmentedButton<String>(
                    style: SegmentedButton.styleFrom(
                      backgroundColor: Colors.white.withOpacity(0.8),
                      selectedBackgroundColor: colorScheme.primary.withOpacity(0.2),
                      foregroundColor: colorScheme.onSurface,
                      selectedForegroundColor: colorScheme.primary,
                      side: BorderSide(
                        color: Colors.grey.withOpacity(0.3),
                      ),
                    ),
                    segments: _categories.map((category) {
                      return ButtonSegment<String>(
                        value: category,
                        label: Text(category),
                      );
                    }).toList(),
                    selected: {_selectedCategory},
                    onSelectionChanged: (Set<String> selection) {
                      setState(() {
                        _selectedCategory = selection.first;
                      });
                    },
                  ),
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
                    Tab(text: 'Live', icon: Icon(Icons.live_tv_outlined)),
                    Tab(text: 'Upcoming', icon: Icon(Icons.schedule_outlined)),
                    Tab(text: 'Trending', icon: Icon(Icons.trending_up_outlined)),
                  ],
                ),
              ),
              pinned: true,
            ),
          ];
        },
        body: TabBarView(
          controller: _tabController,
          children: [
            _buildLiveStreamsTab(),
            _buildUpcomingStreamsTab(),
            _buildTrendingTab(),
          ],
        ),
      ),
    );
  }

  Widget _buildQuickStat(String label, String value, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: Colors.white.withOpacity(0.2),
          width: 1,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            color: Colors.white,
            size: 16,
          ),
          const SizedBox(width: 6),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                value,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                label,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 10,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildLiveStreamsTab() {
    return CustomScrollView(
      slivers: [
        // Trending Tags
        SliverToBoxAdapter(
          child: Container(
            height: 50,
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: ListView.builder(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 16),
              itemCount: _trendingTags.length,
              itemBuilder: (context, index) {
                final tag = _trendingTags[index];
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: [
                          Theme.of(context).colorScheme.primary.withOpacity(0.1),
                          Theme.of(context).colorScheme.secondary.withOpacity(0.1),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: Theme.of(context).colorScheme.primary.withOpacity(0.3),
                      ),
                    ),
                     child: Text(
                       tag,
                       style: Theme.of(context).textTheme.bodySmall?.copyWith(
                         color: Theme.of(context).colorScheme.primary,
                         fontWeight: FontWeight.w600,
                       ),
                     ),
                   ),
                 );
               },
            ),
          ),
        ),
        
        // Live Streams Grid
        SliverPadding(
          padding: const EdgeInsets.all(16),
          sliver: SliverGrid(
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              childAspectRatio: 0.8,
              crossAxisSpacing: 16,
              mainAxisSpacing: 16,
            ),
            delegate: SliverChildBuilderDelegate(
              (context, index) {
                final stream = _trendingStreams[index];
                return _buildStreamCard(stream, index, isLive: true);
              },
              childCount: _trendingStreams.length,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildUpcomingStreamsTab() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: ListView.builder(
        itemCount: _upcomingStreams.length,
        itemBuilder: (context, index) {
          final stream = _upcomingStreams[index];
          return _buildUpcomingStreamCard(stream, index);
        },
      ),
    );
  }

  Widget _buildTrendingTab() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: GridView.builder(
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 2,
          childAspectRatio: 0.75,
          crossAxisSpacing: 16,
          mainAxisSpacing: 16,
        ),
        itemCount: _trendingStreams.length,
        itemBuilder: (context, index) {
          final stream = _trendingStreams[index];
          return _buildStreamCard(stream, index, isLive: false);
        },
      ),
    );
  }

  Widget _buildStreamCard(Map<String, dynamic> stream, int index, {required bool isLive}) {
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
              // TODO: Navigate to stream
            },
            borderRadius: BorderRadius.circular(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Thumbnail Container
                Stack(
                  children: [
                    Container(
                      height: 120,
                      width: double.infinity,
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                          colors: [
                            colorScheme.primary.withOpacity(0.3),
                            colorScheme.secondary.withOpacity(0.3),
                          ],
                        ),
                      ),
                      child: Icon(
                        Icons.live_tv_outlined,
                        size: 40,
                        color: colorScheme.onSurface.withOpacity(0.4),
                      ),
                    ),
                    if (isLive)
                      Positioned(
                        top: 8,
                        left: 8,
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(
                            color: colorScheme.error,
                            borderRadius: BorderRadius.circular(12),
                            boxShadow: [
                              BoxShadow(
                                color: colorScheme.error.withOpacity(0.3),
                                blurRadius: 8,
                                offset: const Offset(0, 2),
                              ),
                            ],
                          ),
                          child: const Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                Icons.circle,
                                color: Colors.white,
                                size: 8,
                              ),
                              SizedBox(width: 4),
                              Text(
                                'LIVE',
                                style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 10,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    Positioned(
                      top: 8,
                      right: 8,
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(
                          color: Colors.black.withOpacity(0.7),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(
                              Icons.visibility,
                              color: Colors.white,
                              size: 12,
                            ),
                            const SizedBox(width: 4),
                            Text(
                              '${stream['viewers']}',
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 10,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
                
                // Content
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          stream['title']?.toString() ?? '',
                          style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          stream['seller']?.toString() ?? '',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: colorScheme.onSurface.withOpacity(0.6),
                          ),
                        ),
                        const Spacer(),
                        Row(
                          children: [
                            if (stream['rating'] != null) ...[
                              Row(
                                children: [
                                  const Icon(
                                    Icons.star,
                                    size: 12,
                                    color: Colors.amber,
                                  ),
                                  const SizedBox(width: 2),
                                  Text(
                                    stream['rating'].toString(),
                                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                ],
                              ),
                            ],
                            const SizedBox(width: 8),
                            if (stream['bids'] != null) ...[
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                decoration: BoxDecoration(
                                  color: colorScheme.primaryContainer,
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Text(
                                  '${stream['bids']} bids',
                                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: colorScheme.onPrimaryContainer,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              ),
                            ],
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
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildUpcomingStreamCard(Map<String, dynamic> stream, int index) {
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
              // TODO: Navigate to stream
            },
            borderRadius: BorderRadius.circular(20),
            child: Row(
              children: [
                // Thumbnail
                Container(
                  width: 100,
                  height: 100,
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        colorScheme.primary.withOpacity(0.3),
                        colorScheme.secondary.withOpacity(0.3),
                      ],
                    ),
                  ),
                  child: Stack(
                    children: [
                      Center(
                        child: Icon(
                          Icons.schedule_outlined,
                          size: 30,
                          color: colorScheme.onSurface.withOpacity(0.4),
                        ),
                      ),
                      Positioned(
                        bottom: 8,
                        left: 8,
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.7),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            stream['startTime']?.toString() ?? '',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 10,
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
                          stream['title']?.toString() ?? '',
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'by ${stream['seller']}',
                          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: colorScheme.onSurface.withOpacity(0.7),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(
                            color: colorScheme.primaryContainer,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Text(
                            stream['category']?.toString() ?? '',
                            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: colorScheme.onPrimaryContainer,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                        const Spacer(),
                        Row(
                          children: [
                            Text(
                              'Starts at ${stream['startTime']}',
                              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: colorScheme.primary,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const Spacer(),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                              decoration: BoxDecoration(
                                color: colorScheme.primary,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                'Set Reminder',
                                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                  color: colorScheme.onPrimary,
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