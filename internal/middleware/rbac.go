package middleware

import (
	v1 "admin-webrtc-go/api/v1"
	"admin-webrtc-go/internal/service"
	"admin-webrtc-go/pkg/jwt"
	"admin-webrtc-go/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func RBACAuth(j *jwt.JWT, us service.UserService, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.Request.Header.Get("Authorization")
		if tokenStr == "" {
			logger.WithContext(ctx).Error("token empty", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}))
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		claims, err := j.ParseToken(tokenStr)
		if err != nil {
			logger.WithContext(ctx).Error("token parse error", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}), zap.Error(err))
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		// 需要取出路径中的/:api参数用于permission表查询是否拥有请求权限
		apiPath := ctx.Param("api")

		flag, err := us.CheckAPIAuthPermission(ctx, claims.UserId, apiPath)
		// 查询报错
		if err != nil {
			logger.WithContext(ctx).Error("CheckAPIAuthPermission method error", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}), zap.Error(err))
			v1.HandleError(ctx, http.StatusNotFound, err, nil)
			ctx.Abort()
			return
		}
		// 无权限
		if !flag {
			logger.WithContext(ctx).Error(v1.ErrUnauthorized.Error(), zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}))
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
