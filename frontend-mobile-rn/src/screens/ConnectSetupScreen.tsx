// src/screens/ConnectSetupScreen.tsx - Stripe Connect Seller Onboarding
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
import { axios } from 'axios';

// API client for Stripe Connect
const connectApiClient = axios.create({
  baseURL: process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8090',
  timeout: 15000,
});

// Connect account types
type AccountType = 'express' | 'standard' | 'custom';
type BusinessType = 'individual' | 'company';

interface ConnectAccountData {
  email: string;
  businessType: BusinessType;
  accountType: AccountType;
  country: string;
  businessProfile: {
    name: string;
    website?: string;
    phone?: string;
    description?: string;
  };
}

export const ConnectSetupScreen: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const [accountData, setAccountData] = useState<ConnectAccountData>({
    email: '',
    businessType: 'individual',
    accountType: 'express',
    country: 'MY', // Malaysia
    businessProfile: {
      name: '',
      website: '',
      phone: '',
      description: '',
    },
  });
  const [accountLink, setAccountLink] = useState<string>('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Create Connect account
  const createConnectAccount = async () => {
    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      
      const response = await connectApiClient.post('/api/v1/connect/accounts', {
        email: accountData.email,
        business_type: accountData.businessType,
        type: accountData.accountType,
        country: accountData.country,
        business_profile: accountData.businessProfile,
      });

      const { account_id } = response.data.data;
      console.log('✅ Connect Account Created:', account_id);

      // Create account link for onboarding
      await createAccountLink(account_id);
      
    } catch (error: any) {
      console.error('❌ Connect Account Creation Failed:', error);
      setErrors({ 
        general: error.response?.data?.message || 'Failed to create Connect account' 
      });
    } finally {
      setLoading(false);
    }
  };

  // Create account link for onboarding
  const createAccountLink = async (accountId: string) => {
    try {
      const response = await connectApiClient.post('/api/v1/connect/account-links', {
        account_id: accountId,
        refresh_url: 'https://blytz.app/connect/refresh',
        return_url: 'https://blytz.app/connect/return',
        type: 'account_onboarding',
      });

      const { url } = response.data.data;
      setAccountLink(url);
      
      Alert.alert(
        '🎯 Connect Account Created!',
        'Please complete your seller onboarding by clicking the link below.',
        [
          { text: 'Cancel', style: 'cancel' },
          { text: 'Open Onboarding', onPress: () => openAccountLink(url) }
        ]
      );
      
    } catch (error: any) {
      console.error('❌ Account Link Creation Failed:', error);
      setErrors({ 
        accountLink: error.response?.data?.message || 'Failed to create onboarding link' 
      });
    }
  };

  // Open account link
  const openAccountLink = (url: string) => {
    Linking.openURL(url).catch((err) => {
      console.error('Failed to open URL:', err);
      Alert.alert('Error', 'Could not open the onboarding link');
    });
  };

  // Validate form
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Email validation
    if (!accountData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(accountData.email)) {
      newErrors.email = 'Please enter a valid email';
    }

    // Business name validation
    if (!accountData.businessProfile.name.trim()) {
      newErrors.name = 'Business name is required';
    } else if (accountData.businessProfile.name.length < 2) {
      newErrors.name = 'Business name must be at least 2 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle input changes
  const handleInputChange = (field: keyof ConnectAccountData, value: any) => {
    setAccountData(prev => ({
      ...prev,
      [field]: value,
    }));
    
    // Clear error when user types
    if (errors[field]) {
      setErrors(prev => ({
        ...prev,
        [field]: '',
      }));
    }
  };

  const handleProfileChange = (field: string, value: string) => {
    setAccountData(prev => ({
      ...prev,
      businessProfile: {
        ...prev.businessProfile,
        [field]: value,
      },
    }));
    
    // Clear error when user types
    if (errors[field]) {
      setErrors(prev => ({
        ...prev,
        [field]: '',
      }));
    }
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
            <Text style={styles.title}>Stripe Connect Setup</Text>
            <Text style={styles.subtitle}>
              Create your seller account to start accepting payments
            </Text>
          </View>

          {/* Form */}
          <View style={styles.formSection}>
            {/* Email */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Email Address</Text>
              <TextInput
                style={[
                  styles.input,
                  errors.email ? styles.inputError : null,
                ]}
                value={accountData.email}
                onChangeText={(value) => handleInputChange('email', value)}
                placeholder="seller@example.com"
                keyboardType="email-address"
                autoCapitalize="none"
                autoCorrect={false}
                editable={!loading}
              />
              {errors.email && (
                <Text style={styles.errorText}>{errors.email}</Text>
              )}
            </View>

            {/* Business Type */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Business Type</Text>
              <View style={styles.radioButtonGroup}>
                {(['individual', 'company'] as BusinessType[]).map((type) => (
                  <TouchableOpacity
                    key={type}
                    style={[
                      styles.radioButton,
                      accountData.businessType === type && styles.radioButtonSelected,
                    ]}
                    onPress={() => handleInputChange('businessType', type)}
                    disabled={loading}
                  >
                    <Text style={[
                      styles.radioButtonText,
                      accountData.businessType === type && styles.radioButtonTextSelected,
                    ]}>
                      {type.charAt(0).toUpperCase() + type.slice(1)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            {/* Account Type */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Account Type</Text>
              <Text style={styles.description}>
                Express: Quick setup, Stripe handles compliance
              </Text>
              <Text style={styles.description}>
                Standard: More control, requires full integration
              </Text>
              <Text style={styles.description}>
                Custom: Full control, requires full compliance handling
              </Text>
              <View style={styles.radioButtonGroup}>
                {(['express', 'standard', 'custom'] as AccountType[]).map((type) => (
                  <TouchableOpacity
                    key={type}
                    style={[
                      styles.radioButton,
                      accountData.accountType === type && styles.radioButtonSelected,
                    ]}
                    onPress={() => handleInputChange('accountType', type)}
                    disabled={loading}
                  >
                    <Text style={[
                      styles.radioButtonText,
                      accountData.accountType === type && styles.radioButtonTextSelected,
                    ]}>
                      {type.charAt(0).toUpperCase() + type.slice(1)}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            {/* Country */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Country</Text>
              <TextInput
                style={styles.input}
                value={accountData.country}
                onChangeText={(value) => handleInputChange('country', value)}
                placeholder="MY"
                autoCapitalize="characters"
                maxLength={2}
                editable={!loading}
              />
              <Text style={styles.description}>
                Currently only Malaysia (MY) is supported
              </Text>
            </View>

            {/* Business Profile */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Business Name</Text>
              <TextInput
                style={[
                  styles.input,
                  errors.name ? styles.inputError : null,
                ]}
                value={accountData.businessProfile.name}
                onChangeText={(value) => handleProfileChange('name', value)}
                placeholder="Your Business Name"
                editable={!loading}
              />
              {errors.name && (
                <Text style={styles.errorText}>{errors.name}</Text>
              )}
            </View>

            {/* Website (Optional) */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Website (Optional)</Text>
              <TextInput
                style={styles.input}
                value={accountData.businessProfile.website}
                onChangeText={(value) => handleProfileChange('website', value)}
                placeholder="https://yourbusiness.com"
                keyboardType="url"
                autoCapitalize="none"
                autoCorrect={false}
                editable={!loading}
              />
            </View>

            {/* Phone (Optional) */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Phone (Optional)</Text>
              <TextInput
                style={styles.input}
                value={accountData.businessProfile.phone}
                onChangeText={(value) => handleProfileChange('phone', value)}
                placeholder="+60123456789"
                keyboardType="phone"
                editable={!loading}
              />
            </View>

            {/* Description (Optional) */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>Business Description (Optional)</Text>
              <TextInput
                style={[styles.input, styles.textArea]}
                value={accountData.businessProfile.description}
                onChangeText={(value) => handleProfileChange('description', value)}
                placeholder="Describe your business..."
                multiline
                numberOfLines={4}
                textAlignVertical="top"
                editable={!loading}
              />
            </View>
          </View>

          {/* Error Display */}
          {errors.general && (
            <View style={styles.errorSection}>
              <Text style={styles.errorText}>{errors.general}</Text>
            </View>
          )}

          {/* Create Account Button */}
          <TouchableOpacity
            style={[
              styles.createButton,
              loading && styles.createButtonDisabled,
            ]}
            onPress={createConnectAccount}
            disabled={loading}
          >
            {loading ? (
              <ActivityIndicator color="#fff" size="small" />
            ) : (
              <Text style={styles.createButtonText}>
                Create Stripe Connect Account
              </Text>
            )}
          </TouchableOpacity>

          {/* Account Link Display */}
          {accountLink && (
            <View style={styles.linkSection}>
              <Text style={styles.linkLabel}>🎯 Onboarding Link</Text>
              <TouchableOpacity
                style={styles.linkButton}
                onPress={() => openAccountLink(accountLink)}
              >
                <Text style={styles.linkButtonText}>
                  Complete Onboarding
                </Text>
              </TouchableOpacity>
              <Text style={styles.linkDescription}>
                Complete your seller setup by filling out your business information
              </Text>
            </View>
          )}

          {/* Test Information */}
          {__DEV__ && (
            <View style={styles.testSection}>
              <Text style={styles.testTitle}>🧪 Connect Test Information</Text>
              <Text style={styles.testText}>
                Account Type: {accountData.accountType}
              </Text>
              <Text style={styles.testText}>
                Business Type: {accountData.businessType}
              </Text>
              <Text style={styles.testText}>
                Country: {accountData.country}
              </Text>
              <Text style={styles.testText}>
                Email: {accountData.email}
              </Text>
              <Text style={styles.testText}>
                Business Name: {accountData.businessProfile.name}
              </Text>
              {accountLink && (
                <Text style={styles.testText}>
                  Account Link: Ready ✅
                </Text>
              )}
            </View>
          )}

          {/* API Information */}
          <View style={styles.apiSection}>
            <Text style={styles.apiTitle}>🔗 Stripe Connect API</Text>
            <Text style={styles.apiText}>Endpoint: /api/v1/connect/accounts</Text>
            <Text style styles={styles.apiText}>Method: POST</Text>
            <Text style={styles.apiText}>Platform: Blytz Marketplace</Text>
            <Text style={styles.apiText}>Provider: Stripe Connect</Text>
            <Text style={styles.apiText}>Features: Payouts, Dashboard, Onboarding</Text>
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
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 16,
    color: '#6c757d',
    textAlign: 'center',
    lineHeight: 24,
  },
  formSection: {
    marginBottom: 30,
  },
  inputGroup: {
    marginBottom: 20,
  },
  label: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  description: {
    fontSize: 14,
    color: '#6c757d',
    marginBottom: 8,
  },
  input: {
    borderWidth: 1,
    borderColor: '#dee2e6',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 16,
    color: '#1a1a1a',
    backgroundColor: '#fff',
  },
  inputError: {
    borderColor: '#dc3545',
    borderWidth: 2,
  },
  textArea: {
    height: 80,
    textAlignVertical: 'top',
  },
  errorText: {
    color: '#dc3545',
    fontSize: 14,
    marginTop: 4,
  },
  radioButtonGroup: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  radioButton: {
    borderWidth: 1,
    borderColor: '#dee2e6',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#fff',
  },
  radioButtonSelected: {
    backgroundColor: '#635bff',
    borderColor: '#635bff',
  },
  radioButtonText: {
    fontSize: 14,
    color: '#495057',
  },
  radioButtonTextSelected: {
    color: '#fff',
    fontWeight: '600',
  },
  errorSection: {
    backgroundColor: '#f8d7da',
    borderRadius: 8,
    padding: 16,
    marginBottom: 20,
    borderLeftWidth: 4,
    borderLeftColor: '#dc3545',
  },
  createButton: {
    backgroundColor: '#635bff',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginBottom: 20,
  },
  createButtonDisabled: {
    backgroundColor: '#6c757d',
  },
  createButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  linkSection: {
    backgroundColor: '#d4edda',
    borderRadius: 8,
    padding: 16,
    marginBottom: 20,
    borderLeftWidth: 4,
    borderLeftColor: '#28a745',
  },
  linkLabel: {
    fontSize: 16,
    fontWeight: '600',
    color: '#155724',
    marginBottom: 8,
  },
  linkButton: {
    backgroundColor: '#28a745',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    alignItems: 'center',
    marginBottom: 8,
  },
  linkButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  linkDescription: {
    fontSize: 14,
    color: '#155724',
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

export default ConnectSetupScreen;