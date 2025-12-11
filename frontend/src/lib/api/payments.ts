import apiClient from './api-client';
import { PaymentMethodInfo, PaymentRequest, PaymentResponse } from '@/types';

// Payment-related API calls
export class PaymentService {
  // Get available payment methods
  async getPaymentMethods() {
    return apiClient.get<PaymentMethodInfo[]>('/api/v1/payments/methods');
  }

  // Create a new payment
  async createPayment(paymentRequest: PaymentRequest) {
    return apiClient.post<PaymentResponse>('/api/v1/payments', paymentRequest);
  }

  // Get payment status by ID
  async getPaymentStatus(paymentId: string) {
    return apiClient.get<PaymentResponse>(`/api/v1/payments/${paymentId}`);
  }

  // Update payment status (for webhooks/admin)
  async updatePaymentStatus(paymentId: string, status: string) {
    return apiClient.put<PaymentResponse>(`/api/v1/payments/${paymentId}/status`, {
      status,
    });
  }

  // Refund a payment
  async refundPayment(paymentId: string, reason?: string, amount?: number) {
    return apiClient.post<{
      refundId: string;
      originalPaymentId: string;
      amount: number;
      status: string;
      createdAt: string;
    }>(`/api/v1/payments/${paymentId}/refund`, {
      reason,
      amount,
    });
  }

  // Get payment history for current user
  async getMyPayments(filter?: {
    page?: number;
    limit?: number;
    status?: string;
    startDate?: string;
    endDate?: string;
  }) {
    const params = new URLSearchParams();
    
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());
    if (filter?.status) params.append('status', filter.status);
    if (filter?.startDate) params.append('startDate', filter.startDate);
    if (filter?.endDate) params.append('endDate', filter.endDate);

    const queryString = params.toString();
    const endpoint = `/api/v1/payments/my${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      payments: PaymentResponse[];
      total: number;
      page: number;
      limit: number;
      hasNext: boolean;
    }>(endpoint);
  }

  // Get refund history for current user
  async getMyRefunds(filter?: {
    page?: number;
    limit?: number;
    status?: string;
  }) {
    const params = new URLSearchParams();
    
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.limit) params.append('limit', filter.limit.toString());
    if (filter?.status) params.append('status', filter.status);

    const queryString = params.toString();
    const endpoint = `/api/v1/payments/refunds/my${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<{
      refunds: {
        refundId: string;
        originalPaymentId: string;
        amount: number;
        status: string;
        reason: string;
        createdAt: string;
        processedAt?: string;
      }[];
      total: number;
      page: number;
      limit: number;
      hasNext: boolean;
    }>(endpoint);
  }

  // Fiuu-specific methods
  async getFiuuSeamlessConfig(orderData: {
    orderId: string;
    amount: number;
    billName: string;
    billEmail: string;
    billMobile: string;
    billDesc: string;
    channel?: string;
  }) {
    const params = new URLSearchParams({
      order_id: orderData.orderId,
      amount: orderData.amount.toString(),
      bill_name: orderData.billName,
      bill_email: orderData.billEmail,
      bill_mobile: orderData.billMobile,
      bill_desc: orderData.billDesc,
      channel: orderData.channel || 'FPX',
    });

    return apiClient.get<{
      version: string;
      actionType: string;
      merchantID: string;
      paymentMethod: string;
      orderNumber: string;
      amount: number;
      currency: string;
      productDescription: string;
      userName: string;
      userEmail: string;
      userContact: string;
      remark: string;
      lang: string;
      vcode: string;
      callbackURL: string;
      returnURL: string;
      backgroundUrl: string;
      sandbox: boolean;
      scriptUrl: string;
    }>(`/api/v1/payments/fiuu/seamless/config?${params}`);
  }

  // Verify Fiuu payment callback
  async verifyFiuuCallback(callbackData: Record<string, any>) {
    return apiClient.post<{
      valid: boolean;
      paymentId?: string;
      status?: string;
      amount?: number;
    }>('/api/v1/payments/fiuu/verify', callbackData);
  }

  // Stripe-specific methods (if needed)
  async createStripePaymentIntent(paymentRequest: {
    amount: number;
    currency: string;
    orderId: string;
    description?: string;
  }) {
    return apiClient.post<{
      clientSecret: string;
      paymentIntentId: string;
    }>('/api/v1/payments/stripe/intent', paymentRequest);
  }

  // Confirm Stripe payment
  async confirmStripePayment(paymentIntentId: string, paymentMethodId?: string) {
    return apiClient.post<PaymentResponse>(`/api/v1/payments/stripe/confirm/${paymentIntentId}`, {
      paymentMethodId,
    });
  }

  // Get payment statistics for current user
  async getMyPaymentStats() {
    return apiClient.get<{
      totalPaid: number;
      totalRefunded: number;
      successfulPayments: number;
      pendingPayments: number;
      failedPayments: number;
      refunds: number;
    }>('/api/v1/payments/my/stats');
  }
}

// Export singleton instance
export const paymentService = new PaymentService();