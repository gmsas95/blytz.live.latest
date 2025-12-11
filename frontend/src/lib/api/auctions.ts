import apiClient from './api-client';
import { Auction, Bid, AuctionFilter } from '@/types';

// Auction-related API calls
export class AuctionService {
  // Get auctions with optional filtering
  async getAuctions(filter?: AuctionFilter) {
    const params = new URLSearchParams();
    
    if (filter?.status) params.append('status', filter.status);
    if (filter?.categoryId) params.append('categoryId', filter.categoryId);
    if (filter?.sellerId) params.append('sellerId', filter.sellerId);
    if (filter?.minPrice) params.append('minPrice', filter.minPrice.toString());
    if (filter?.maxPrice) params.append('maxPrice', filter.maxPrice.toString());
    if (filter?.search) params.append('search', filter.search);
    if (filter?.sortBy) params.append('sortBy', filter.sortBy);
    if (filter?.sortOrder) params.append('sortOrder', filter.sortOrder);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());

    const queryString = params.toString();
    const endpoint = `/api/v1/auctions${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      auctions: (Auction & { timeRemaining: number; bidCount: number; canBid: boolean; minNextBid: number })[];
      total: number;
      page: number;
      limit: number;
      hasNext: boolean;
    }>(endpoint);
  }

  // Get single auction by ID
  async getAuction(id: string) {
    return apiClient.get<Auction & { 
      timeRemaining: number; 
      bidCount: number; 
      isWatched: boolean; 
      canBid: boolean; 
      minNextBid: number;
    }>(`/api/v1/auctions/${id}`);
  }

  // Get active auctions
  async getActiveAuctions() {
    return apiClient.get<Auction[]>('/api/v1/auctions/active');
  }

  // Create new auction (requires authentication)
  async createAuction(auctionData: {
    productId: string;
    startingPrice: number;
    reservePrice?: number;
    startTime: string;
    endTime: string;
    minBidIncrement?: number;
    description?: string;
  }) {
    return apiClient.post<Auction>('/api/v1/auctions', auctionData);
  }

  // Update auction (requires authentication)
  async updateAuction(id: string, auctionData: Partial<{
    startTime: string;
    endTime: string;
    reservePrice: number;
    minBidIncrement: number;
    description: string;
  }>) {
    return apiClient.put<Auction>(`/api/v1/auctions/${id}`, auctionData);
  }

  // Delete auction (requires authentication)
  async deleteAuction(id: string) {
    return apiClient.delete<void>(`/api/v1/auctions/${id}`);
  }

  // Start auction (requires authentication)
  async startAuction(id: string) {
    return apiClient.post<void>(`/api/v1/auctions/${id}/start`);
  }

  // End auction (requires authentication)
  async endAuction(id: string) {
    return apiClient.post<void>(`/api/v1/auctions/${id}/end`);
  }

  // Place a bid on an auction (requires authentication)
  async placeBid(auctionId: string, amount: number) {
    return apiClient.post<Bid>(`/api/v1/auctions/${auctionId}/bids`, {
      amount,
    });
  }

  // Get bids for an auction
  async getAuctionBids(auctionId: string, filter?: {
    page?: number;
    limit?: number;
  }) {
    const params = new URLSearchParams();
    
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());

    const queryString = params.toString();
    const endpoint = `/api/v1/auctions/${auctionId}/bids${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      bids: Bid[];
      total: number;
      page: number;
      limit: number;
      hasNext: boolean;
    }>(endpoint);
  }

  // Get auctions for current user (requires authentication)
  async getMyAuctions(filter?: {
    page?: number;
    limit?: number;
    status?: string;
  }) {
    const params = new URLSearchParams();
    
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());
    if (filter?.status) params.append('status', filter.status);

    const queryString = params.toString();
    const endpoint = `/api/v1/auctions/my${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      auctions: (Auction & { timeRemaining: number; bidCount: number; canBid: boolean; minNextBid: number })[];
      total: number;
      page: number;
      limit: number;
      hasNext: boolean;
    }>(endpoint);
  }

  // Get auctions for a specific seller
  async getSellerAuctions(sellerId: string, filter?: {
    page?: number;
    limit?: number;
    status?: string;
  }) {
    const params = new URLSearchParams();
    
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());
    if (filter?.status) params.append('status', filter.status);

    const queryString = params.toString();
    const endpoint = `/api/v1/auctions/seller/${sellerId}${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      auctions: (Auction & { timeRemaining: number; bidCount: number; canBid: boolean; minNextBid: number })[];
      total: number;
      page: number;
      limit: number;
      hasNext: boolean;
    }>(endpoint);
  }

  // Watch/unwatch auction (requires authentication)
  async toggleWatchAuction(auctionId: string) {
    return apiClient.post<{ isWatched: boolean }>(`/api/v1/auctions/${auctionId}/watch`);
  }

  // Get user's bid history (requires authentication)
  async getMyBidHistory(filter?: {
    page?: number;
    limit?: number;
    auctionId?: string;
  }) {
    const params = new URLSearchParams();
    
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());
    if (filter?.auctionId) params.append('auctionId', filter.auctionId);

    const queryString = params.toString();
    const endpoint = `/api/v1/auctions/bids/my${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      bids: Bid[];
      total: number;
      page: number;
      limit: number;
      hasNext: boolean;
    }>(endpoint);
  }
}

// Export singleton instance
export const auctionService = new AuctionService();