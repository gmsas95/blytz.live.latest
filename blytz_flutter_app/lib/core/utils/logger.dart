import 'package:logger/logger.dart';

class AppLogger {
  final Logger _logger = Logger(
    printer: PrettyPrinter(
      dateTimeFormat: DateTimeFormat.onlyTimeAndSinceStart,
    ),
  );

  AppLogger._(); // Private constructor

  static final AppLogger _instance = AppLogger._();

  static AppLogger get instance => _instance;

  factory AppLogger() => _instance;

  void debug(String message, {dynamic error, StackTrace? stackTrace}) {
    _logger.d(message, error: error, stackTrace: stackTrace);
  }

  void info(String message, {dynamic error, StackTrace? stackTrace}) {
    _logger.i(message, error: error, stackTrace: stackTrace);
  }

  void warning(String message, {dynamic error, StackTrace? stackTrace}) {
    _logger.w(message, error: error, stackTrace: stackTrace);
  }

  void error(String message, {dynamic error, StackTrace? stackTrace}) {
    _logger.e(message, error: error, stackTrace: stackTrace);
  }

  void log(String message, {dynamic error, StackTrace? stackTrace}) {
    _logger.d(message, error: error, stackTrace: stackTrace);
  }

  void logApiRequest(String method, String url, {dynamic data}) {
    _logger.d('API Request: $method $url', error: data);
  }

  void logApiResponse(String method, String url, int statusCode, {dynamic data}) {
    _logger.i('API Response: $method $url - $statusCode', error: data);
  }

  void logApiError(String method, String url, dynamic error) {
    _logger.e('API Error: $method $url', error: error);
  }

  void logWebSocketEvent(String event, {dynamic data}) {
    _logger.d('WebSocket: $event', error: data);
  }

  void logUserAction(String action, {Map<String, dynamic>? params}) {
    _logger.i('User Action: $action', error: params);
  }

  void logPerformance(String operation, Duration duration, {Map<String, dynamic>? params}) {
    _logger.i('Performance: $operation took ${duration.inMilliseconds}ms', error: params);
  }
}