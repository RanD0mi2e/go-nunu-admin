package middleware

import (
	v1 "admin-webrtc-go/api/v1"
	"admin-webrtc-go/internal/service"
	"admin-webrtc-go/pkg/jwt"
	"admin-webrtc-go/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"regexp"
)

func RBACAuth(j *jwt.JWT, us service.UserService, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, exists := ctx.Get("claims")
		if !exists {
			logger.WithContext(ctx).Error("token empty", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}), zap.Error(v1.ErrUnauthorized))
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		tokenStr, ok := token.(string)
		if !ok {
			logger.WithContext(ctx).Error("token invalid", zap.Any("data", map[string]interface{}{
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

		path := ctx.Request.URL.Path
		re := regexp.MustCompile(`auth_(\d+)`)
		match := re.FindStringSubmatch(path)

		flag, err := us.CheckAPIAuthPermission(ctx, claims.UserId, match[0])
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
			v1.HandleError(ctx, http.StatusUnauthorized, err, nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
