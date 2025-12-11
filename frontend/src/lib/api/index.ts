// Export all API services
export { default as apiClient } from './api-client';
export { authService } from './auth';
export { productService } from './products';
export { auctionService } from './auctions';
export { paymentService } from './payments';

// Re-export types for convenience
export type {
  ApiResponse,
  PaginatedResponse,
  Product,
  ProductFilter,
  Auction,
  AuctionFilter,
  Bid,
  User,
  PaymentMethodInfo,
  PaymentRequest,
  PaymentResponse,
} from '@/types';