package handler

import (
	"github.com/mq/api/internal/port/inbound"
)

type OnboardingHandler struct {
	onboarding inbound.OnboardingService
}

func NewOnboardingHandler(onboarding inbound.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{onboarding: onboarding}
}
