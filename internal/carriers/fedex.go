package carriers

import (
	"context"
	"math"
	"time"

	"github.com/shipit-ai-demo-org/shipping-rates-service/internal/quote"
)

// FedEx pricing mirrors the 2026 list rates with CargoCloud's volume
// discount applied, plus the published fuel surcharge.
type FedEx struct {
	baseCents        int64
	perKgCents       int64
	fuelSurchargePct float64
}

func NewFedEx() *FedEx {
	return &FedEx{baseCents: 815, perKgCents: 198, fuelSurchargePct: 0.065}
}

func (f *FedEx) Name() string { return "fedex" }

func (f *FedEx) Quote(_ context.Context, s quote.Shipment) (quote.Rate, error) {
	billable := quote.BillableWeightKg(s)
	amount := f.baseCents + int64(math.Ceil(billable))*f.perKgCents
	transit := 4

	if s.Service == "express" {
		amount = amount*2 + 350
		transit = 1
	}

	amount += int64(math.Round(float64(amount) * f.fuelSurchargePct))

	return quote.Rate{
		Carrier:     "fedex",
		Service:     s.Service,
		AmountCents: amount,
		Currency:    "USD",
		TransitDays: transit,
		QuotedAt:    time.Now().UTC(),
	}, nil
}
