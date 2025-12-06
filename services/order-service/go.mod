module github.com/gmsas95/blytz-mvp/services/order-service

go 1.25

replace github.com/gmsas95/blytz-mvp/shared => ../../shared

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/gmsas95/blytz-mvp/shared v0.0.0
	github.com/google/uuid v1.4.0
	go.uber.org/zap v1.26.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

replace (
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid => github.com/google/uuid v1.4.0
	go.uber.org/zap => go.uber.org/zap v1.26.0
	gorm.io/driver/postgres => gorm.io/driver/postgres v1.5.4
	gorm.io/gorm => gorm.io/gorm v1.25.5
)