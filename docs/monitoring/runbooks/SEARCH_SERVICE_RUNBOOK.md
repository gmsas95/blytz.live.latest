# Search Service Runbook

## Overview

The Search Service provides unified search capabilities across products, auctions, and users with full-text search, faceted filtering, and real-time indexing.

## Service Details

- **Service Name**: search-service
- **Port**: 8095
- **Database**: PostgreSQL with full-text search
- **Cache**: Redis
- **Max Connections**: 200
- **Performance Targets**: <100ms response time, 70% cache hit rate, 5000+ requests/second

## Architecture

### Components

1. **Search API Layer**: RESTful endpoints for search operations
2. **Search Service Layer**: Business logic and caching
3. **Search Repository**: PostgreSQL full-text search operations
4. **Indexing Service**: Real-time entity indexing
5. **Analytics Service**: Search analytics and insights

### Key Features

- **Full-Text Search**: PostgreSQL tsvector with GIN indexes
- **Fuzzy Search**: Trigram similarity for typo tolerance
- **Faceted Search**: Category, price, location, rating filters
- **Search Suggestions**: Auto-complete and query suggestions
- **Real-time Indexing**: Immediate index updates
- **Caching**: Redis-based result caching
- **Analytics**: Search query tracking and popular searches

## Health Checks

### Service Health
```bash
curl http://localhost:8095/health
```

Expected response:
```json
{
  "success": true,
  "data": {
    "service": "search-service",
    "status": "healthy",
    "cache": {
      "status": "enabled",
      "info": "redis_connection_info"
    },
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

### Database Health
Check PostgreSQL connection and search index status:
```sql
-- Check search index table
SELECT COUNT(*) FROM search_indices WHERE is_active = true;

-- Check search vector index
SELECT indexname FROM pg_indexes WHERE tablename = 'search_indices';
```

### Cache Health
Check Redis connection:
```bash
redis-cli -h localhost -p 6379 ping
```

## Performance Monitoring

### Key Metrics

1. **Response Time**: Target <100ms for 95th percentile
2. **Throughput**: Target 5000+ requests/second
3. **Cache Hit Rate**: Target 70%+
4. **Error Rate**: Target <0.1%
5. **Indexing Latency**: Target <1s for entity updates

### Monitoring Endpoints

- **Metrics**: `http://localhost:8095/metrics` (Prometheus format)
- **Stats**: `http://localhost:8095/api/v1/admin/stats`
- **Status**: `http://localhost:8095/api/v1/admin/status`

### Grafana Dashboards

- **Search Service Performance**: Response times, throughput, error rates
- **Search Analytics**: Popular queries, search patterns
- **Cache Performance**: Hit rates, memory usage

## Troubleshooting

### Common Issues

#### 1. Slow Search Performance

**Symptoms**:
- Response times >100ms
- High database query times

**Diagnosis**:
```bash
# Check slow queries
curl "http://localhost:8095/api/v1/admin/stats" | jq '.stats.database'

# Check cache hit rate
curl "http://localhost:8095/api/v1/admin/stats" | jq '.stats.cache'
```

**Solutions**:
1. Check PostgreSQL query execution plans
2. Verify GIN indexes are being used
3. Increase cache TTL for frequently searched terms
4. Optimize search vector weights

#### 2. Low Cache Hit Rate

**Symptoms**:
- Cache hit rate <70%
- High database load

**Diagnosis**:
```bash
# Check Redis memory usage
redis-cli info memory

# Check cache statistics
curl "http://localhost:8095/api/v1/admin/stats" | jq '.stats.cache'
```

**Solutions**:
1. Increase cache TTL for popular queries
2. Implement cache warming strategies
3. Check Redis memory limits
4. Optimize cache key generation

#### 3. Search Index Out of Sync

**Symptoms**:
- Missing entities in search results
- Stale search results

**Diagnosis**:
```sql
-- Check last indexed time
SELECT MAX(indexed_at) FROM search_indices;

-- Check for missing entities
SELECT COUNT(*) FROM products p 
WHERE NOT EXISTS (
  SELECT 1 FROM search_indices s 
  WHERE s.entity_type = 'product' AND s.entity_id = p.id
);
```

**Solutions**:
1. Trigger full reindex: `curl -X POST http://localhost:8095/api/v1/admin/reindex`
2. Check indexing service logs
3. Verify database triggers are working
4. Check message queue for indexing events

#### 4. High Memory Usage

**Symptoms**:
- OOM kills
- High memory consumption

**Diagnosis**:
```bash
# Check service memory
docker stats search-service

# Check PostgreSQL memory
docker exec postgres psql -U postgres -c "SELECT * FROM pg_stat_activity;"

# Check Redis memory
redis-cli info memory
```

**Solutions**:
1. Tune PostgreSQL work_mem
2. Optimize search result pagination
3. Implement result streaming for large result sets
4. Increase container memory limits

## Emergency Procedures

### Service Recovery

1. **Immediate Actions**:
   ```bash
   # Restart service
   docker-compose restart search-service
   
   # Check health
   curl http://localhost:8095/health
   ```

2. **Database Issues**:
   ```bash
   # Check PostgreSQL
   docker-compose exec postgres pg_isready -U postgres
   
   # Restart database if needed
   docker-compose restart postgres
   ```

3. **Cache Issues**:
   ```bash
   # Clear cache
   docker-compose exec redis redis-cli FLUSHALL
   
   # Restart Redis
   docker-compose restart redis
   ```

### Scaling Procedures

#### Horizontal Scaling

1. **Add Search Service Instances**:
   ```yaml
   # In docker-compose.yml
   search-service-2:
     build: ./services/search-service
     ports:
       - "8096:8095"
     environment:
       - DATABASE_URL=postgres://postgres:postgres@postgres:5432/blytz_db
       - REDIS_URL=redis://redis:6379
   ```

2. **Load Balancer Configuration**:
   ```nginx
   upstream search_backend {
       server search-service-1:8095;
       server search-service-2:8095;
   }
   
   server {
       listen 80;
       location /api/v1/search {
           proxy_pass http://search_backend;
       }
   }
   ```

#### Vertical Scaling

1. **Increase Resources**:
   ```yaml
   search-service:
     deploy:
       resources:
         limits:
           cpus: '2.0'
           memory: 4G
         reservations:
           cpus: '1.0'
           memory: 2G
   ```

2. **Database Scaling**:
   ```sql
   -- Increase connection pool
   ALTER SYSTEM SET max_connections = 400;
   
   -- Optimize for search workloads
   ALTER SYSTEM SET work_mem = '256MB';
   ALTER SYSTEM SET maintenance_work_mem = '1GB';
   ```

## Maintenance

### Regular Tasks

1. **Daily**:
   - Monitor performance metrics
   - Check error rates
   - Review search analytics

2. **Weekly**:
   - Analyze popular search queries
   - Optimize slow queries
   - Review cache performance

3. **Monthly**:
   - Full index rebuild if needed
   - Database maintenance (VACUUM, ANALYZE)
   - Performance tuning review

### Index Maintenance

```bash
# Trigger full reindex
curl -X POST http://localhost:8095/api/v1/admin/reindex

# Check index status
curl http://localhost:8095/api/v1/admin/status

# Monitor reindex progress
docker-compose logs -f search-service
```

### Cache Maintenance

```bash
# Check Redis memory usage
redis-cli info memory

# Clear expired cache entries
redis-cli --scan --pattern "search:*" | xargs redis-cli del

# Monitor cache hit rate
redis-cli info stats
```

## Security

### Access Control

1. **API Authentication**:
   - Internal services use service tokens
   - Public endpoints are rate-limited
   - Admin endpoints require authentication

2. **Database Security**:
   - Read-only replicas for search queries
   - Connection encryption
   - Row-level security for sensitive data

3. **Cache Security**:
   - Redis authentication
   - Network isolation
   - TLS encryption

### Rate Limiting

```bash
# Check rate limit status
curl -I http://localhost:8095/api/v1/search?q=test

# Rate limit headers
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Integration Points

### Data Sources

1. **Product Service**: Product indexing and updates
2. **Auction Service**: Auction indexing and updates
3. **Auth Service**: User indexing and updates

### Consumers

1. **Frontend**: Search API for user interface
2. **Gateway Service**: API gateway integration
3. **Analytics Service**: Search analytics data

### Webhooks

```bash
# Product update webhook
POST /api/v1/webhooks/product-update
{
  "event_type": "product.updated",
  "entity_id": "product-123",
  "timestamp": "2024-01-01T00:00:00Z"
}

# Auction update webhook
POST /api/v1/webhooks/auction-update
{
  "event_type": "auction.updated",
  "entity_id": "auction-456",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Testing

### Performance Testing

```bash
# Load testing with Apache Bench
ab -n 10000 -c 100 "http://localhost:8095/api/v1/search?q=laptop"

# Concurrent search test
go test -bench=BenchmarkSearchPerformance ./tests/performance/

# Memory leak testing
go test -run TestMemoryUsage ./tests/performance/
```

### Functional Testing

```bash
# Search functionality
curl "http://localhost:8095/api/v1/search?q=laptop&limit=20"

# Indexing functionality
curl -X POST http://localhost:8095/api/v1/index \
  -H "Content-Type: application/json" \
  -d '{"entity_type":"product","entity_id":"test-1","title":"Test Product"}'

# Health check
curl http://localhost:8095/health
```

## Alerts and Notifications

### Critical Alerts

1. **Service Down**: Health check failures
2. **High Error Rate**: >1% error rate
3. **Slow Response**: >500ms 95th percentile
4. **Cache Failure**: Redis connection issues
5. **Database Issues**: PostgreSQL connection problems

### Warning Alerts

1. **High Memory**: >80% memory usage
2. **Low Cache Hit**: <60% cache hit rate
3. **Index Lag**: >5 minutes indexing delay
4. **High Throughput**: >80% of max capacity

## Contact and Escalation

### Service Team
- **Primary**: Search Service Team
- **Secondary**: Platform Engineering Team
- **Escalation**: Infrastructure Team

### Communication Channels

1. **Slack**: #search-service-alerts
2. **Email**: search-team@company.com
3. **PagerDuty**: Search Service Rotation
4. **Incident**: incidents@company.com

## Documentation

- **API Documentation**: `/docs/api/search-service.md`
- **Architecture**: `/docs/architecture/search-service.md`
- **Deployment**: `/docs/deployment/search-service.md`
- **Monitoring**: `/docs/monitoring/search-service.md`

## Change Management

### Deployment Process

1. **Staging Testing**: Deploy to staging environment
2. **Performance Testing**: Load test in staging
3. **Canary Deployment**: Deploy to 10% of production
4. **Full Deployment**: Deploy to 100% of production
5. **Monitoring**: Watch for issues for 30 minutes

### Rollback Procedure

```bash
# Quick rollback to previous version
docker-compose pull search-service:previous
docker-compose up -d search-service

# Verify rollback
curl http://localhost:8095/health
curl http://localhost:8095/api/v1/search?q=test
```

## Runbook Maintenance

- **Last Updated**: 2024-01-01
- **Next Review**: 2024-02-01
- **Owner**: Search Service Team
- **Reviewers**: Platform Engineering Team

### Version History

- **v1.0.0**: Initial implementation with full-text search
- **v1.1.0**: Added fuzzy search and suggestions
- **v1.2.0**: Performance optimizations and caching improvements
- **v1.3.0**: Added analytics and monitoring enhancements