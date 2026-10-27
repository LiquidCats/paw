package inbound

import (
	"github.com/LiquidCats/paw/internal/kernel/ports"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/app/port/outbound"
	"github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"
)

type APIRepositroyRegistry ports.Registry[values.ProviderName, outbound.RateAPIRepository]
