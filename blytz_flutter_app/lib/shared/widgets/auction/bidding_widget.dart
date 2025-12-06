import 'package:blytz_flutter_app/core/constants/app_constants.dart';
import 'package:blytz_flutter_app/core/utils/extensions.dart';
import 'package:blytz_flutter_app/data/models/auction_model.dart';
import 'package:blytz_flutter_app/features/auctions/presentation/providers/auction_providers.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class BiddingWidget extends ConsumerStatefulWidget {

  const BiddingWidget({
    required this.auction, super.key,
  });
  final AuctionModel auction;

  @override
  ConsumerState<BiddingWidget> createState() => _BiddingWidgetState();
}

class _BiddingWidgetState extends ConsumerState<BiddingWidget> {
  final TextEditingController _bidController = TextEditingController();
  final FocusNode _bidFocusNode = FocusNode();
  bool _isPlacingBid = false;

  @override
  void initState() {
    super.initState();
    // Set minimum bid as default
    _bidController.text = (widget.auction.currentBidAmount + 1.0).toStringAsFixed(2);
  }

  @override
  void dispose() {
    _bidController.dispose();
    _bidFocusNode.dispose();
    super.dispose();
  }

  double get _minimumBid => widget.auction.currentBidAmount + 1.0;
  double get _currentBid => double.tryParse(_bidController.text) ?? 0.0;

  bool _isValidBid() {
    return _currentBid >= _minimumBid;
  }

  void _onQuickBid(double amount) {
    setState(() {
      _bidController.text = amount.toStringAsFixed(2);
    });
    _bidFocusNode.requestFocus();
  }

  Future<void> _placeBid() async {
    if (!_isValidBid()) {
      _showError('Bid must be at least ${_minimumBid.currencyFormat}');
      return;
    }

    setState(() {
      _isPlacingBid = true;
    });

    try {
      final bidData = {
        'auctionId': widget.auction.id,
        'amount': _currentBid,
      };

      await ref.read(bidPlacementProvider(bidData).future);

      if (mounted) {
        _showSuccess('Bid placed successfully!');
        // Refresh auction data
        ref.invalidate(auctionDetailProvider(widget.auction.id));
        Navigator.of(context).pop();
      }
    } catch (e) {
      if (mounted) {
        _showError(e.toString());
      }
    } finally {
      if (mounted) {
        setState(() {
          _isPlacingBid = false;
        });
      }
    }
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: Colors.red,
      ),
    );
  }

  void _showSuccess(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: Colors.green,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.zero,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(AppConstants.radius16)),
      ),
      elevation: AppConstants.elevation8,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
          // Header
          Semantics(
            label: 'Place your bid',
            child: Row(
              children: [
                Icon(
                  Icons.gavel,
                  color: Theme.of(context).primaryColor,
                ),
                const SizedBox(width: AppConstants.spacing8),
                Text(
                  'Place Your Bid',
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
          TextFormField(
            controller: _bidController,
            focusNode: _bidFocusNode,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              prefixText: r'$',
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(AppConstants.radius8),
              ),
              errorText: _isValidBid() ? null : 'Minimum bid is ${_minimumBid.currencyFormat}',
            ),
            onChanged: (value) {
              setState(() {}); // Trigger rebuild to update validation
            },
          ),

          const SizedBox(height: AppConstants.spacing16),

          // Bid Button
          Semantics(
            button: true,
            label: _isPlacingBid ? 'Placing bid' : 'Place bid',
            child: FilledButton(
              onPressed: (_isPlacingBid || !_isValidBid()) ? null : _placeBid,
              style: FilledButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.primary,
                foregroundColor: Theme.of(context).colorScheme.onPrimary,
                minimumSize: const Size(double.infinity, 56),
                padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(AppConstants.radius16),
                ),
              ),
              child: _isPlacingBid
                  ? Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation<Color>(
                              Theme.of(context).colorScheme.onPrimary,
                            ),
                          ),
                        ),
                        const SizedBox(width: AppConstants.spacing8),
                        const Text('Placing Bid...'),
                      ],
                    )
                  : const Text('Place Bid'),
            ),
          ),

          const SizedBox(height: AppConstants.spacing8),

          // Terms
          Semantics(
            label: 'Bid agreement terms',
            child: Text(
              'By placing a bid, you commit to purchase this item if you win the auction.',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Colors.grey[600],
              ),
              textAlign: TextAlign.center,
            ),
          ),
        ],
        ),
      ),
    );
  }

  Widget _buildQuickBidButton(double amount) {
    return Expanded(
      child: OutlinedButton(
        onPressed: () => _onQuickBid(amount),
        style: OutlinedButton.styleFrom(
          backgroundColor: Colors.grey[200],
          foregroundColor: Colors.black87,
        ),
        child: Text(
          '\$${amount.toStringAsFixed(2)}',
          style: const TextStyle(
            fontWeight: FontWeight.w600,
            color: Colors.black87,
          ),
        ),
      ),
    );
  }
}