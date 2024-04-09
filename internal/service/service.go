package service

import (
	"admin-webrtc-go/internal/repository"
	"admin-webrtc-go/pkg/jwt"
	"admin-webrtc-go/pkg/log"
	"admin-webrtc-go/pkg/sid"
)

type Service struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
}

func NewService(tm repository.Transaction, logger *log.Logger, sid *sid.Sid, jwt *jwt.JWT) *Service {
	return &Service{
		logger: logger,
		sid:    sid,
		jwt:    jwt,
		tm:     tm,
	}
}
