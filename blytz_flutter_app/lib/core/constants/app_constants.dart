class AppConstants {
  static const String appName = 'Blytz';
  static const String appVersion = '1.0.0';
  
  // Storage Keys
  static const String authTokenKey = 'auth_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userKey = 'user_data';
  static const String userIdKey = 'user_id';
  static const String themeKey = 'theme_mode';
  static const String languageKey = 'language_code';
  
  // Pagination
  static const int defaultPageSize = 20;
  static const int maxPageSize = 100;
  
  // Bid limits
  static const double minBidAmount = 1;
  static const double maxBidAmount = 10000;
  static const double minBidIncrement = 1;
  
  // Image constraints
  static const int maxImageSize = 5 * 1024 * 1024; // 5MB
  static const List<String> supportedImageFormats = ['jpg', 'jpeg', 'png', 'webp'];
  
  // Rate limiting
  static const int maxBidAttemptsPerMinute = 10;
  static const int maxChatMessagesPerMinute = 30;
  
  // Cache durations
  static const Duration auctionCacheDuration = Duration(minutes: 5);
  static const Duration userCacheDuration = Duration(hours: 1);
  static const Duration imageCacheDuration = Duration(days: 7);
  
  // Design Tokens - Spacing
  static const double spacing2 = 2;
  static const double spacing4 = 4;
  static const double spacing8 = 8;
  static const double spacing12 = 12;
  static const double spacing16 = 16;
  static const double spacing20 = 20;
  static const double spacing24 = 24;
  static const double spacing32 = 32;
  static const double spacing48 = 48;
  static const double spacing64 = 64;
  
  // Design Tokens - Border Radius
  static const double radius4 = 4;
  static const double radius8 = 8;
  static const double radius12 = 12;
  static const double radius16 = 16;
  static const double radius20 = 20;
  static const double radius24 = 24;
  static const double radius32 = 32;
  
  // Design Tokens - Elevation
  static const double elevation0 = 0;
  static const double elevation1 = 1;
  static const double elevation2 = 2;
  static const double elevation4 = 4;
  static const double elevation8 = 8;
  static const double elevation12 = 12;
  static const double elevation16 = 16;
}