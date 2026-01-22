package carriers

import (
	"context"
	"math"
	"time"

	"github.com/shipit-ai-demo-org/shipping-rates-service/internal/quote"
)

// UPS prices against the negotiated 2026 CargoCloud rate card. In production
// the live tariff API sits behind the same interface (UPS_RATING_URL); the
// rate card keeps quoting available when the tariff API is down.
type UPS struct {
	baseCents  int64
	perKgCents int64
}

func NewUPS() *UPS {
	return &UPS{baseCents: 749, perKgCents: 212}
}

func (u *UPS) Name() string { return "ups" }

func (u *UPS) Quote(_ context.Context, s quote.Shipment) (quote.Rate, error) {
	billable := quote.BillableWeightKg(s)
	amount := u.baseCents + int64(math.Ceil(billable))*u.perKgCents
	transit := 5

	if s.Service == "express" {
		amount = amount*2 + 399 // air uplift + handling
		transit = 2
	}

	return quote.Rate{
		Carrier:     "ups",
		Service:     s.Service,
		AmountCents: amount,
		Currency:    "USD",
		TransitDays: transit,
		QuotedAt:    time.Now().UTC(),
	}, nil
}
