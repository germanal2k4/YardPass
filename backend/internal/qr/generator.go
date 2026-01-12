package qr

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateQR(ctx context.Context, passID uuid.UUID) ([]byte, error) {
	qrData := fmt.Sprintf("yardpass://pass/%s", passID.String())

	png, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return png, nil
}

func (g *Generator) ParseQR(ctx context.Context, qrData string) (uuid.UUID, error) {
	var uuidStr string
	if len(qrData) > 18 && qrData[:18] == "yardpass://pass/" {
		uuidStr = qrData[18:]
	} else {
		uuidStr = qrData
	}

	passID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid QR code format: %w", err)
	}

	return passID, nil
}

