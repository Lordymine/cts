// Package scan coordena os scanners de cada categoria.
package scan

import (
	"context"
	"errors"
	"fmt"

	"cts/internal/target"
)

// Scanner inspeciona uma categoria e devolve o que achou.
// Interface pequena de propósito: dá pra implementar e testar sem cerimônia.
type Scanner interface {
	Category() target.Category
	Scan(ctx context.Context) ([]target.Target, error)
}

// Run roda todos os scanners e junta os alvos. Erro de um scanner é
// embrulhado com a categoria e acumulado — não aborta os outros.
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
