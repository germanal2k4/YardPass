package qr

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

// legacyPrefix was previously embedded in guest pass QR codes; plain hardware
// scanners send it verbatim, breaking uuid parsing, so new codes contain only
// the UUID. The prefix is still stripped when parsing for old codes.
const legacyPrefix = "yardpass://pass/"

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateQR(ctx context.Context, passID uuid.UUID) ([]byte, error) {
	return g.GenerateRawQR(ctx, passID.String())
}

func (g *Generator) GenerateRawQR(ctx context.Context, payload string) ([]byte, error) {
	png, err := qrcode.Encode(payload, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("generate QR code: %w", err)
	}

	return png, nil
}

func (g *Generator) ParseQR(ctx context.Context, qrData string) (uuid.UUID, error) {
	uuidStr := strings.TrimPrefix(strings.TrimSpace(qrData), legacyPrefix)

	passID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid QR code format: %w", err)
	}

	return passID, nil
}
