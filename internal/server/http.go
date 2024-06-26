package server

import (
	apiV1 "admin-webrtc-go/api/v1"
	"admin-webrtc-go/docs"
	"admin-webrtc-go/internal/handler"
	"admin-webrtc-go/internal/middleware"
	"admin-webrtc-go/internal/service"
	"admin-webrtc-go/pkg/jwt"
	"admin-webrtc-go/pkg/log"
	"admin-webrtc-go/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHTTPServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	userHandler *handler.UserHandler,
	userService service.UserService,
) *http.Server {
	gin.SetMode(gin.DebugMode)
	s := http.NewServer(
		gin.Default(),
		logger,
		http.WithServerHost(conf.GetString("http.host")),
		http.WithServerPort(conf.GetInt("http.port")),
	)

	// swagger doc
	docs.SwaggerInfo.BasePath = "/v1"
	s.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		//ginSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", conf.GetInt("app.http.port"))),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	s.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		// middleware.SignMiddleware(log),
	)
	s.GET("/", func(ctx *gin.Context) {
		logger.WithContext(ctx).Info("hello")
		apiV1.HandleSuccess(ctx, map[string]interface{}{
			":)": "Thank you for using nunu!",
		})
	})

	v1 := s.Group("/v1")
	{
		// No route group has permission
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/register", userHandler.Register)
			noAuthRouter.POST("/login", userHandler.Login)
		}
		// Non-strict permission routing group
		noStrictAuthRouter := v1.Group("/").Use(middleware.NoStrictAuth(jwt, logger))
		{
			noStrictAuthRouter.GET("/user", userHandler.GetProfile)
			noStrictAuthRouter.GET("/getMenuTree", userHandler.GetMenuTree)
		}
		// 需要非严格校验Api权限的分组
		noStrictApiAuthRouter := v1.Group("/:api").Use(middleware.NoStrictAuth(jwt, logger), middleware.RBACAuth(jwt, userService, logger))
		{
			noStrictApiAuthRouter.GET("/apiAuthTest", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "apiAuthTest Success!",
				})
			})
		}
		// Strict permission routing group
		strictAuthRouter := v1.Group("/").Use(middleware.StrictAuth(jwt, logger))
		{
			strictAuthRouter.PUT("/user", userHandler.UpdateProfile)
		}
		// 需要严格校验Api权限的分组
		strictApiAuthRouter := v1.Group("/:api").Use(middleware.StrictAuth(jwt, logger), middleware.RBACAuth(jwt, userService, logger))
		{
			strictApiAuthRouter.GET("apiStrictAuthTest")
		}
	}

	return s
}
