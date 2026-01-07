package importer

import (
	"context"
	"fmt"
	"io"
)

type VanguardParser struct{}

func (p *VanguardParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
	return nil, fmt.Errorf("vanguard parser not implemented: need sample CSV to build parser")
}


