import apiClient from './api-client';
import { Product, ProductFilter } from '@/types';

// Product-related API calls
export class ProductService {
  // Get products with optional filtering
  async getProducts(filter?: ProductFilter) {
    const params = new URLSearchParams();
    
    if (filter?.search) params.append('search', filter.search);
    if (filter?.category) params.append('category', filter.category);
    if (filter?.subcategory) params.append('subcategory', filter.subcategory);
    if (filter?.minPrice) params.append('minPrice', filter.minPrice.toString());
    if (filter?.maxPrice) params.append('maxPrice', filter.maxPrice.toString());
    if (filter?.sellerId) params.append('sellerId', filter.sellerId);
    if (filter?.status) params.append('status', filter.status);
    if (filter?.sortBy) params.append('sortBy', filter.sortBy);
    if (filter?.sortOrder) params.append('sortOrder', filter.sortOrder);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());

    const queryString = params.toString();
    const endpoint = `/api/v1/products${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      products: Product[];
      total: number;
      page: number;
      pageSize: number;
      hasNext: boolean;
    }>(endpoint);
  }

  // Get single product by ID
  async getProduct(id: string) {
    return apiClient.get<Product>(`/api/v1/products/${id}`);
  }

  // Get featured products
  async getFeaturedProducts() {
    return apiClient.get<Product[]>('/api/v1/products/featured');
  }

  // Create new product (requires authentication)
  async createProduct(productData: {
    name: string;
    description: string;
    price: number;
    currency: string;
    imageUrl: string;
    images?: string[];
    stock: number;
    category: string;
    subcategory?: string;
    tags?: string[];
  }) {
    return apiClient.post<Product>('/api/v1/products', productData);
  }

  // Update product (requires authentication)
  async updateProduct(id: string, productData: Partial<{
    name: string;
    description: string;
    price: number;
    currency: string;
    imageUrl: string;
    images: string[];
    stock: number;
    category: string;
    subcategory: string;
    tags: string[];
    status: string;
    isFeatured: boolean;
  }>) {
    return apiClient.put<Product>(`/api/v1/products/${id}`, productData);
  }

  // Delete product (requires authentication)
  async deleteProduct(id: string) {
    return apiClient.delete<void>(`/api/v1/products/${id}`);
  }

  // Update product inventory (requires authentication)
  async updateInventory(id: string, stock: number, reserved?: number) {
    return apiClient.put<Product>(`/api/v1/products/${id}/inventory`, {
      stock,
      reserved,
    });
  }

  // Get products for current user (requires authentication)
  async getMyProducts(filter?: {
    page?: number;
    limit?: number;
    status?: string;
  }) {
    const params = new URLSearchParams();
    
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());
    if (filter?.status) params.append('status', filter.status);

    const queryString = params.toString();
    const endpoint = `/api/v1/products/my${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      products: Product[];
      total: number;
      page: number;
      pageSize: number;
      hasNext: boolean;
    }>(endpoint);
  }
}

// Export singleton instance
export const productService = new ProductService();