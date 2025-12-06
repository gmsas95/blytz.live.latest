import 'package:blytz_flutter_app/core/utils/logger.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Dio provider for HTTP requests
final dioProvider = Provider<Dio>((ref) {
  return Dio(BaseOptions(
    baseUrl: 'http://localhost:8080/api/v1', // Default API base URL
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 10),
  ));
});

// Mock providers for bidding functionality - these would connect to actual API
final bidPlacementProvider = FutureProvider.autoDispose.family<void, Map<String, dynamic>>((ref, bidData) async {
  final dio = ref.watch(dioProvider);

  try {
    await dio.post('/bids', data: bidData);
    AppLogger().info('Bid placed successfully: $bidData');
  } catch (e) {
    AppLogger().error('Failed to place bid: $e');
    rethrow;
  }
});

final auctionDetailProvider = FutureProvider.autoDispose.family<Map<String, dynamic>, String>((ref, auctionId) async {
  final dio = ref.watch(dioProvider);

  try {
    final response = await dio.get('/auctions/$auctionId');
    return response.data as Map<String, dynamic>;
  } catch (e) {
    AppLogger().error('Failed to fetch auction details: $e');
    rethrow;
  }
});