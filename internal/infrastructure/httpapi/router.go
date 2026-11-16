package httpapi

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.elastic.co/apm/module/apmgin/v2"
	"go.elastic.co/apm/v2"
	"go.uber.org/zap"
)

type RouterConfig struct {
	Mode           string
	SwaggerEnabled bool
	SimulatedDelay time.Duration
}

func NewRouter(
	cfg RouterConfig,
	payments *PaymentHandler,
	tracer *apm.Tracer,
	log *zap.Logger,
) (*gin.Engine, error) {
	gin.SetMode(cfg.Mode)
	// Rejects payloads carrying fields the API does not expect (mass assignment).
	binding.EnableDecoderDisallowUnknownFields = true

	router := gin.New()
	// The API is not behind a known proxy; trusting none prevents X-Forwarded-For spoofing.
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	// Order matters: apmgin opens the transaction and stores it in the request
	// context; requestContext derives the correlated logger from it.
	router.Use(
		apmgin.Middleware(router, apmgin.WithTracer(tracer)),
		requestContext(log),
		requestLogger(),
		gin.Recovery(),
	)

	if cfg.SwaggerEnabled {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := router.Group("/", securityHeaders(), limitBody(), errorHandler())
	api.GET("/payments", payments.List)
	api.POST("/payments", payments.Create)
	api.POST("/payments/delay", simulatedLatency(cfg.SimulatedDelay), payments.CreateWithDelay)

	return router, nil
}
