package quote

import (
	"context"
	"errors"
	"log"
	"math"
	"sort"
	"time"
)

var ErrNoQuotes = errors.New("no carriers returned a quote")

// Shipment describes a parcel to be rated.
type Shipment struct {
	OriginZip string  `json:"originZip"`
	DestZip   string  `json:"destZip"`
	WeightKg  float64 `json:"weightKg"`
	LengthCm  float64 `json:"lengthCm"`
	WidthCm   float64 `json:"widthCm"`
	HeightCm  float64 `json:"heightCm"`
	Service   string  `json:"service"` // "ground" | "express"
}

// Rate is a single carrier quote, normalized to integer cents.
type Rate struct {
	Carrier     string    `json:"carrier"`
	Service     string    `json:"service"`
	AmountCents int64     `json:"amountCents"`
	Currency    string    `json:"currency"`
	TransitDays int       `json:"transitDays"`
	QuotedAt    time.Time `json:"quotedAt"`
}

// Carrier is implemented by each adapter in internal/carriers.
type Carrier interface {
	Name() string
	Quote(ctx context.Context, s Shipment) (Rate, error)
}

// BillableWeightKg returns the greater of scale weight and dimensional
// weight, using the industry-standard 5000 cm³/kg divisor.
func BillableWeightKg(s Shipment) float64 {
	dim := s.LengthCm * s.WidthCm * s.HeightCm / 5000
	return math.Max(s.WeightKg, dim)
}

type Engine struct {
	carriers []Carrier
	cache    *Cache
}

func NewEngine(cache *Cache, carriers ...Carrier) *Engine {
	return &Engine{cache: cache, carriers: carriers}
}

// BestRate fans out to every registered carrier and returns the cheapest
// quote plus the full set of alternatives. A carrier failure degrades the
// comparison rather than failing the request.
func (e *Engine) BestRate(ctx context.Context, s Shipment) (Rate, []Rate, error) {
	if e.cache != nil {
		if best, alts, ok := e.cache.Get(s); ok {
			return best, alts, nil
		}
	}

	var rates []Rate
	for _, c := range e.carriers {
		rate, err := c.Quote(ctx, s)
		if err != nil {
			log.Printf("quote: carrier %s failed: %v", c.Name(), err)
			continue
		}
		rates = append(rates, rate)
	}
	if len(rates) == 0 {
		return Rate{}, nil, ErrNoQuotes
	}

	sort.Slice(rates, func(i, j int) bool {
		return rates[i].AmountCents < rates[j].AmountCents
	})

	if e.cache != nil {
		e.cache.Set(s, rates[0], rates)
	}

	return rates[0], rates, nil
}
