package importer

import (
	"context"
	"fmt"
	"io"
)

type FidelityParser struct{}

func (p *FidelityParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
	return nil, fmt.Errorf("fidelity parser not implemented: need sample CSV to build parser")
}



