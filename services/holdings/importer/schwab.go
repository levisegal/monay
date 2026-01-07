package importer

import (
	"context"
	"fmt"
	"io"
)

type SchwabParser struct{}

func (p *SchwabParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
	return nil, fmt.Errorf("schwab parser not implemented: need sample CSV to build parser")
}


