import 'package:blytz_flutter_app/core/constants/app_constants.dart';
import 'package:blytz_flutter_app/core/utils/extensions.dart';
import 'package:blytz_flutter_app/data/models/auction_model.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class AuctionCard extends ConsumerWidget {

  const AuctionCard({
    required this.auction, super.key,
    this.onTap,
    this.onWatch,
    this.showWatchButton = true,
  });
  final AuctionModel auction;
  final VoidCallback? onTap;
  final VoidCallback? onWatch;
  final bool showWatchButton;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Semantics(
      button: true,
      label: 'Auction: ${auction.title}, Current bid: ${auction.currentBidAmount.currencyFormat}, ${auction.totalBids} bids',
      child: Card(
      elevation: 4,
      shadowColor: Colors.black.withOpacity(0.1),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppConstants.radius12)),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(AppConstants.radius12),
        child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          // Image Section
          Flexible(
            flex: 3,
            child: Stack(
              children: [
                // Product Image
                Container(
                  width: double.infinity,
                  decoration: const BoxDecoration(
                      borderRadius: BorderRadius.vertical(
                        top: Radius.circular(AppConstants.radius12),
                      ),
                  ),
                  child: ClipRRect(
                    borderRadius: const BorderRadius.vertical(
                      top: Radius.circular(12),
                    ),
                    child: auction.images.isNotEmpty
                        ? CachedNetworkImage(
                            imageUrl: auction.images.first,
                            fit: BoxFit.cover,
                            placeholder: (context, url) => Container(
                              color: Colors.grey[300],
                              child: const Center(
                                child: Icon(
                                  Icons.image,
                                  color: Colors.grey,
                                  size: 40,
                                ),
                              ),
                            ),
                            errorWidget: (context, url, error) => Container(
                              color: Colors.grey[300],
                              child: const Center(
                                child: Icon(
                                  Icons.broken_image,
                                  color: Colors.grey,
                                  size: 40,
                                ),
                              ),
                            ),
                          )
                        : Container(
                            color: Colors.grey[300],
                            child: const Center(
                              child: Icon(
                                Icons.image,
                                color: Colors.grey,
                                size: 40,
                              ),
                            ),
                          ),
                  ),
                ),
                
                // Status Badge
                Positioned(
                  top: 8,
                  left: 8,
                  child: _buildStatusBadge(),
                ),
                
                // Watch Button
                if (showWatchButton)
                  Positioned(
                    top: 8,
                    right: 8,
                    child: IconButton(
                      icon: const Icon(Icons.favorite_border, color: Colors.white),
                      onPressed: onWatch,
                      style: IconButton.styleFrom(
                        backgroundColor: Colors.black.withOpacity(0.5),
                        minimumSize: const Size(32, 32),
                      ),
                      tooltip: 'Add to favorites',
                    ),
                  ),
                
                // Live Badge for Active Auctions
                if (auction.isActive)
                  Positioned(
                    bottom: 8,
                    right: 8,
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: Colors.red,
                        borderRadius: BorderRadius.circular(AppConstants.radius12),
                      ),
                      child: const Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.fiber_manual_record,
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
              ],
            ),
          ),
          
          // Content Section
          Flexible(
            flex: 2,
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Title
                  Text(
                    auction.title,
                    style: const TextStyle(
                      fontWeight: FontWeight.w600,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),

                   const SizedBox(height: AppConstants.spacing4),

                  // Category
                  if (auction.categories.isNotEmpty)
                    Text(
                      auction.categories.first,
                      style: TextStyle(
                        color: Theme.of(context).primaryColor,
                        fontSize: 14,
                      ),
                    ),
                   
                  const SizedBox(height: 8),
                   
                  // Current Price
                  Row(
                    children: [
                      Text(
                        'Current Bid',
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.onSurface.withOpacity(0.7),
                          fontSize: 14,
                        ),
                      ),
                      const Spacer(),
                      Text(
                        auction.currentBidAmount.currencyFormat,
                        style: TextStyle(
                          color: Theme.of(context).primaryColor,
                          fontWeight: FontWeight.w600,
                          fontSize: 16,
                        ),
                      ),
                    ],
                  ),
                  
                   const SizedBox(height: AppConstants.spacing4),
                  
                  // Bid Count and Time Left
                  Row(
                    children: [
                      Row(
                        children: [
                          Icon(
                            Icons.gavel,
                            size: 16,
                            color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                          ),
                           const SizedBox(width: AppConstants.spacing4),
                          Text(
                            '${auction.totalBids} bids',
                            style: TextStyle(
                              color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                              fontSize: 14,
                            ),
                          ),
                        ],
                      ),
                      const Spacer(),
                      Row(
                        children: [
                          Icon(
                            Icons.access_time,
                            size: 16,
                            color: auction.isEndingSoon 
                                ? Colors.orange 
                                : Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                          ),
                           const SizedBox(width: AppConstants.spacing4),
                          Text(
                            auction.timeLeft,
                            style: TextStyle(
                              color: auction.isEndingSoon
                                  ? Colors.orange
                                  : Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatusBadge() {
    Color backgroundColor;
    Color textColor;
    String text;

    switch (auction.status) {
      case 'active':
        backgroundColor = Colors.green;
        textColor = Colors.white;
        text = 'ACTIVE';
      case 'ending_soon':
        backgroundColor = Colors.orange;
        textColor = Colors.white;
        text = 'ENDING SOON';
      case 'ended':
        backgroundColor = Colors.grey;
        textColor = Colors.white;
        text = 'ENDED';
      case 'scheduled':
        backgroundColor = Colors.blue;
        textColor = Colors.white;
        text = 'SCHEDULED';
      default:
        backgroundColor = Colors.grey;
        textColor = Colors.white;
        text = auction.status.toUpperCase();
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(AppConstants.radius12),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: textColor,
          fontSize: 10,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}