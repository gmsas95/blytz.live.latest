// src/screens/StripePaymentScreen.tsx - Complete Stripe Mobile Payment
import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Linking,
} from 'react-native';
import {
  useStripe,
  CardField,
  ApplePayButton,
  GooglePayButton,
  useConfirmPayment,
} from '@stripe/stripe-expo';
import { axios } from 'axios';

// API client for Stripe
const stripeApiClient = axios.create({
  baseURL: process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8090',
  timeout: 10000,
});

// Payment types
interface PaymentIntent {
  id: string;
  client_secret: string;
  amount: number;
  currency: string;
  status: string;
  created: number;
}

export const StripePaymentScreen: React.FC = () => {
  const { confirmPayment } = useStripe();
  const [loading, setLoading] = useState<boolean>(false);
  const [paymentIntent, setPaymentIntent] = useState<PaymentIntent | null>(null);
  const [amount, setAmount] = useState<string>('10.00'); // Default $10
  const [currency, setCurrency] = useState<string>('usd');
  const [description, setDescription] = useState<string>('Blytz Auction Item');
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Create payment intent on mount
  useEffect(() => {
    createPaymentIntent();
  }, []);

  // Create payment intent from backend
  const createPaymentIntent = async () => {
    try {
      setLoading(true);
      
      const amountInCents = Math.round(parseFloat(amount) * 100);
      
      const response = await stripeApiClient.post('/api/v1/payments/create-intent', {
        amount: amountInCents,
        currency: currency,
        description: description,
        metadata: {
          platform: 'blytz-mobile',
          timestamp: new Date().toISOString(),
        },
        // Connect destination would be added for marketplace
        // destination_account_id: 'acct_seller123',
        // application_fee_amount: 500, // $5.00 platform fee
      });

      setPaymentIntent(response.data.data);
      console.log('✅ Payment Intent Created:', response.data.data);
    } catch (error: any) {
      console.error('❌ Payment Intent Creation Failed:', error);
      setErrors({ 
        general: 'Failed to create payment intent. Please try again.' 
      });
    } finally {
      setLoading(false);
    }
  };

  // Handle payment confirmation
  const handlePayment = async () => {
    if (!paymentIntent) {
      setErrors({ general: 'No payment intent available' });
      return;
    }

    try {
      setLoading(true);
      
      const { error, paymentIntent: confirmedIntent } = await confirmPayment(
        paymentIntent.client_secret,
        {
          paymentMethodType: 'Card',
          // For mobile, Stripe handles card collection
        }
      );

      if (error) {
        Alert.alert('💳 Payment Failed', error.message);
        setErrors({ payment: error.message });
      } else if (confirmedIntent) {
        Alert.alert('✅ Payment Successful', 
          `Payment of $${(confirmedIntent.amount / 100).toFixed(2)} completed!`);
        console.log('✅ Payment Confirmed:', confirmedIntent);
        // Reset for new payment
        createPaymentIntent();
      }
    } catch (error: any) {
      console.error('❌ Payment Confirmation Failed:', error);
      Alert.alert('💳 Payment Error', error.message);
      setErrors({ payment: error.message });
    } finally {
      setLoading(false);
    }
  };

  // Handle Apple Pay
  const handleApplePay = async () => {
    if (!paymentIntent) return;

    try {
      setLoading(true);
      
      // Apple Pay implementation would go here
      Alert.alert('🍎 Apple Pay', 'Apple Pay integration coming soon!');
      
    } catch (error: any) {
      console.error('❌ Apple Pay Failed:', error);
      Alert.alert('🍎 Apple Pay Error', error.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle Google Pay
  const handleGooglePay = async () => {
    if (!paymentIntent) return;

    try {
      setLoading(true);
      
      // Google Pay implementation would go here
      Alert.alert('🤖 Google Pay', 'Google Pay integration coming soon!');
      
    } catch (error: any) {
      console.error('❌ Google Pay Failed:', error);
      Alert.alert('🤖 Google Pay Error', error.message);
    } finally {
      setLoading(false);
    }
  };

  // Refresh payment intent
  const refreshPaymentIntent = () => {
    createPaymentIntent();
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scrollContainer}>
        <View style={styles.content}>
          {/* Header */}
          <View style={styles.header}>
            <Text style={styles.title}>Stripe Payment</Text>
            <Text style={styles.subtitle}>Secure mobile payment with Stripe</Text>
          </View>

          {/* Payment Info */}
          <View style={styles.paymentInfo}>
            <Text style={styles.infoLabel}>Payment Amount</Text>
            <View style={styles.amountContainer}>
              <Text style={styles.currencySymbol}>$</Text>
              <TextInput
                style={styles.amountInput}
                value={amount}
                onChangeText={setAmount}
                keyboardType="numeric"
                placeholder="10.00"
                editable={!loading}
              />
              <Text style={styles.currencyText}>{currency.toUpperCase()}</Text>
            </View>

            <Text style={styles.infoLabel}>Description</Text>
            <TextInput
              style={styles.descriptionInput}
              value={description}
              onChangeText={setDescription}
              placeholder="Payment description"
              multiline
              editable={!loading}
            />
          </View>

          {/* Stripe Card Field */}
          <View style={styles.cardSection}>
            <Text style={styles.sectionTitle}>Payment Details</Text>
            <CardField
              postalCodeEnabled={false}
              placeholder={{
                number: '4242 4242 4242 4242',
                expMonth: 'MM',
                expYear: 'YY',
                cvc: 'CVV',
              }}
              cardStyle={styles.cardStyle}
              style={styles.cardField}
              onCardChange={(cardDetails) => {
                console.log('Card details:', cardDetails);
              }}
            />
          </View>

          {/* Digital Wallet Buttons */}
          <View style={styles.walletSection}>
            <Text style={styles.sectionTitle}>Digital Wallets</Text>
            
            <View style={styles.walletButtons}>
              <ApplePayButton
                onPress={handleApplePay}
                style={styles.walletButton}
                type="plain"
              />
              
              <GooglePayButton
                onPress={handleGooglePay}
                style={styles.walletButton}
                type="plain"
              />
            </View>
          </View>

          {/* Error Display */}
          {errors.general && (
            <View style={styles.errorSection}>
              <Text style={styles.errorText}>{errors.general}</Text>
            </View>
          )}

          {errors.payment && (
            <View style={styles.errorSection}>
              <Text style={styles.errorText}>{errors.payment}</Text>
            </View>
          )}

          {/* Payment Button */}
          <TouchableOpacity
            style={[
              styles.paymentButton,
              loading && styles.paymentButtonDisabled,
            ]}
            onPress={handlePayment}
            disabled={loading || !paymentIntent}
          >
            {loading ? (
              <ActivityIndicator color="#fff" size="small" />
            ) : (
              <Text style={styles.paymentButtonText}>
                Pay ${amount} {currency.toUpperCase()}
              </Text>
            )}
          </TouchableOpacity>

          {/* Refresh Button */}
          <TouchableOpacity
            style={styles.refreshButton}
            onPress={refreshPaymentIntent}
            disabled={loading}
          >
            <Text style={styles.refreshButtonText}>Refresh Payment Intent</Text>
          </TouchableOpacity>

          {/* Test Information */}
          {__DEV__ && (
            <View style={styles.testSection}>
              <Text style={styles.testTitle}>🧪 Test Information</Text>
              <Text style={styles.testText}>
                Card: 4242 4242 4242 4242
              </Text>
              <Text style={styles.testText}>
                Expiry: Any future date
              </Text>
              <Text style={styles.testText}>
                CVC: Any 3 digits
              </Text>
              <Text style={styles.testText}>
                Payment Intent ID: {paymentIntent?.id}
              </Text>
              <Text style={styles.testText}>
                Amount: {(paymentIntent?.amount || 0) / 100} {currency.toUpperCase()}
              </Text>
              <Text style={styles.testText}>
                Status: {paymentIntent?.status}
              </Text>
            </View>
          )}

          {/* API Test Info */}
          <View style={styles.apiSection}>
            <Text style={styles.apiTitle}>💳 Stripe API Integration</Text>
            <Text style={styles.apiText}>Endpoint: /api/v1/payments/create-intent</Text>
            <Text style={styles.apiText}>Method: POST</Text>
            <Text style={styles.apiText}>Platform: Blytz Mobile</Text>
            <Text style={styles.apiText}>Provider: Stripe Connect</Text>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  scrollContainer: {
    flexGrow: 1,
  },
  content: {
    flex: 1,
    padding: 20,
  },
  header: {
    alignItems: 'center',
    marginBottom: 30,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#6c757d',
    textAlign: 'center',
  },
  paymentInfo: {
    marginBottom: 30,
  },
  infoLabel: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  amountContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 20,
  },
  currencySymbol: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#635bff',
    marginRight: 8,
  },
  amountInput: {
    flex: 1,
    borderWidth: 2,
    borderColor: '#635bff',
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginRight: 8,
  },
  currencyText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#6c757d',
  },
  descriptionInput: {
    borderWidth: 1,
    borderColor: '#dee2e6',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1a1a1a',
    minHeight: 60,
    textAlignVertical: 'top',
  },
  cardSection: {
    marginBottom: 30,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 16,
  },
  cardStyle: {
    backgroundColor: '#ffffff',
    textColor: '#1a1a1a',
    borderColor: '#dee2e6',
    borderWidth: 1,
    borderRadius: 8,
    fontSize: 16,
    placeholderColor: '#6c757d',
  },
  cardField: {
    width: '100%',
    height: 50,
  },
  walletSection: {
    marginBottom: 30,
  },
  walletButtons: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 16,
  },
  walletButton: {
    flex: 1,
    height: 50,
  },
  errorSection: {
    backgroundColor: '#f8d7da',
    borderRadius: 8,
    padding: 12,
    marginBottom: 20,
    borderLeftWidth: 4,
    borderLeftColor: '#dc3545',
  },
  errorText: {
    color: '#721c24',
    fontSize: 14,
  },
  paymentButton: {
    backgroundColor: '#635bff',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginBottom: 16,
  },
  paymentButtonDisabled: {
    backgroundColor: '#6c757d',
  },
  paymentButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  refreshButton: {
    backgroundColor: '#28a745',
    borderRadius: 8,
    paddingVertical: 12,
    alignItems: 'center',
    marginBottom: 30,
  },
  refreshButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  testSection: {
    backgroundColor: '#d1ecf1',
    borderRadius: 8,
    padding: 16,
    marginBottom: 20,
    borderLeftWidth: 4,
    borderLeftColor: '#17a2b8',
  },
  testTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#0c5460',
    marginBottom: 8,
  },
  testText: {
    fontSize: 14,
    color: '#0c5460',
    marginBottom: 4,
  },
  apiSection: {
    backgroundColor: '#e2e3e5',
    borderRadius: 8,
    padding: 16,
    borderLeftWidth: 4,
    borderLeftColor: '#6c757d',
  },
  apiTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#495057',
    marginBottom: 8,
  },
  apiText: {
    fontSize: 14,
    color: '#495057',
    marginBottom: 4,
  },
});

export default StripePaymentScreen;