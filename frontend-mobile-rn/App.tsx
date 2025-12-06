// App.tsx - Complete React Native with Stripe Connect
import 'react-native-gesture-handler';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { NavigationContainer } from '@react-navigation/native';
import { ThemeProvider } from './src/context/ThemeProvider';
import { AuthProvider } from './src/contexts/blytz/auth-context';
import { StripeProvider } from '@stripe/stripe-expo';
import React, { useEffect, useState } from 'react';
import { LogBox, AppRegistry, Platform, StyleSheet } from 'react-native';
import './src/config/i18n';

// Initialize environment
import { developmentUtils } from './src/api/config/environment';

// Test screens for Blytz API with Stripe
import { BlytzLoginScreen } from './src/screens/BlytzLoginScreen';
import { BlytzApiTestScreen } from './src/screens/BlytzApiTestScreen';
import { StripePaymentScreen } from './src/screens/StripePaymentScreen';
import { ConnectSetupScreen } from './src/screens/ConnectSetupScreen';

// Test Navigator for API testing
import { createStackNavigator } from '@react-navigation/stack';
const Stack = createStackNavigator();

const TestNavigator = () => {
  return (
    <Stack.Navigator initialRouteName="StripePayment">
      <Stack.Screen
        name="StripePayment"
        component={StripePaymentScreen}
        options={{
          title: 'Stripe Payment',
          headerStyle: {
            backgroundColor: '#635bff',
          },
          headerTintColor: '#fff',
          headerTitleStyle: {
            fontWeight: 'bold',
          },
        }}
      />
      <Stack.Screen
        name="ConnectSetup"
        component={ConnectSetupScreen}
        options={{
          title: 'Stripe Connect Setup',
          headerStyle: {
            backgroundColor: '#635bff',
          },
          headerTintColor: '#fff',
          headerTitleStyle: {
            fontWeight: 'bold',
          },
        }}
      />
      <Stack.Screen
        name="ApiTest"
        component={BlytzApiTestScreen}
        options={{
          title: 'API Test',
          headerStyle: {
            backgroundColor: '#007bff',
          },
          headerTintColor: '#fff',
          headerTitleStyle: {
            fontWeight: 'bold',
          },
        }}
      />
      <Stack.Screen
        name="BlytzLogin"
        component={BlytzLoginScreen}
        options={{
          title: 'Blytz Login',
          headerStyle: {
            backgroundColor: '#28a745',
          },
          headerTintColor: '#fff',
          headerTitleStyle: {
            fontWeight: 'bold',
          },
        }}
      />
    </Stack.Navigator>
  );
};

// Main Blytz App with Stripe Connect
const BlytzApp = () => {
  const [stripePublishableKey, setStripePublishableKey] = useState<string>('');
  const [isReady, setIsReady] = useState<boolean>(false);

  useEffect(() => {
    // Initialize environment logging
    developmentUtils.logEnvironment();
    
    // Validate configuration
    if (!developmentUtils.validateConfig()) {
      console.error('❌ Blytz configuration is invalid!');
    }

    // Get Stripe publishable key
    const stripeKey = process.env.EXPO_PUBLIC_STRIPE_PUBLISHABLE_KEY || 'pk_test_51234567890abcdef';
    setStripePublishableKey(stripeKey);
    setIsReady(true);

    if (__DEV__) {
      console.log(`🚀 Running Blytz App in development mode on ${Platform.OS}`);
      console.log(`🌍 API Base URL: ${process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8090'}`);
      console.log(`🔌 WebSocket URL: ${process.env.EXPO_PUBLIC_WS_URL || 'ws://localhost:8090'}`);
      console.log(`💳 Stripe Key: ${stripeKey.substring(0, 15)}...`);
      console.log(`🧪 Development Mode: Stripe Connect Testing Enabled`);
    }

    // TODO: Request notification permissions and listen for messages
    // requestUserPermission();
    // listenForMessages();
  }, []);

  if (!isReady || !stripePublishableKey) {
    return null; // Show loading screen
  }

  // For development - show Stripe testing
  if (__DEV__) {
    return (
      <GestureHandlerRootView style={styles.container}>
        <SafeAreaProvider>
          <StripeProvider publishableKey={stripePublishableKey}>
            <AuthProvider>
              <NavigationContainer>
                <TestNavigator />
                <StatusBar style="auto" />
              </NavigationContainer>
            </AuthProvider>
          </StripeProvider>
        </SafeAreaProvider>
      </GestureHandlerRootView>
    );
  }

  // Production - full app with all providers
  return (
    <GestureHandlerRootView style={styles.container}>
      <SafeAreaProvider>
        <ThemeProvider>
          <AuthProvider>
            <StripeProvider publishableKey={stripePublishableKey}>
              <NavigationContainer>
                <Stack.Navigator>
                  <Stack.Screen name="Main" component={MainApp} />
                  <Stack.Screen name="Payment" component={StripePaymentScreen} />
                  <Stack.Screen name="Connect" component={ConnectSetupScreen} />
                </Stack.Navigator>
                <StatusBar style="auto" />
              </NavigationContainer>
            </StripeProvider>
          </AuthProvider>
        </ThemeProvider>
      </SafeAreaProvider>
    </GestureHandlerRootView>
  );
};

// Main App component
const MainApp = () => {
  return (
    <View style={styles.mainContainer}>
      <Text style={styles.mainTitle}>Blytz - Live Auction Marketplace</Text>
      <Text style={styles.mainSubtitle}>Powered by Stripe Connect</Text>
    </View>
  );
};

// Styles
const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  mainContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f8f9fa',
    padding: 20,
  },
  mainTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 10,
    textAlign: 'center',
  },
  mainSubtitle: {
    fontSize: 16,
    color: '#6c757d',
    textAlign: 'center',
  },
});

// Ignore specific warnings for development
LogBox.ignoreLogs([
  'AsyncStorage has been extracted from react-native',
  '[react-native-gesture-handler]',
  'expo-app-loading is deprecated',
  'blytzPerformance.init is not a function', // Ignore if service not ready
]);

// Register app
AppRegistry.registerComponent('main', () => BlytzApp);

export default BlytzApp;