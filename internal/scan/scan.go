// Package scan coordinates the per-category scanners.
package scan

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lordymine/cts/internal/target"
)

// Scanner inspects one category and returns what it found.
// Deliberately small: easy to implement and test without ceremony.
type Scanner interface {
	Category() target.Category
	Scan(ctx context.Context) ([]target.Target, error)
}

// Run runs every scanner and merges the targets. A scanner error is wrapped with
// its category and accumulated — it does not abort the others.
func Run(ctx context.Context, scanners ...Scanner) ([]target.Target, error) {
	var all []target.Target
	var errs []error
	for _, s := range scanners {
		found, err := s.Scan(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("scan %s: %w", s.Category(), err))
			continue
		}
		all = append(all, found...)
	}
	return all, errors.Join(errs...)
}
