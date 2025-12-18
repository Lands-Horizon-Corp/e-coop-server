package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aymerick/raymond"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type PDFOptions[T any] struct {
	Template  string
	Name      string
	Height    float64
	Width     float64
	Unit      string
	Landscape bool
}

func (p PDFOptions[T]) convertToInches() (width, height float64, err error) {
	if p.Unit == "" {
		return 0, 0, fmt.Errorf("unit is required")
	}
	if p.Width <= 0 {
		return 0, 0, fmt.Errorf("invalid width: %v (must be > 0)", p.Width)
	}
	if p.Height <= 0 {
		return 0, 0, fmt.Errorf("invalid height: %v (must be > 0)", p.Height)
	}

	w, err := UnitToInches(p.Width, p.Unit)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert width to inches: %w", err)
	}
	h, err := UnitToInches(p.Height, p.Unit)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert height to inches: %w", err)
	}
	if p.Landscape {
		return h, w, nil
	}
	return w, h, nil
}

func (p PDFOptions[T]) processTemplate(data T) (string, error) {
	if p.Template == "" {
		return "", fmt.Errorf("template is empty")
	}
	out, err := raymond.Render(p.Template, data)
	if err != nil {
		return "", fmt.Errorf("render error: %w", err)
	}
	return out, nil
}

func (p PDFOptions[T]) saveHTMLToPDFBytesWithSize(parentContext context.Context, data T) ([]byte, error) {
	if parentContext == nil {
		return nil, fmt.Errorf("parent context is nil")
	}

	template, err := p.processTemplate(data)
	if err != nil {
		return nil, err
	}

	ctx, cancel := chromedp.NewContext(parentContext)
	defer cancel()

	width, height, err := p.convertToInches()
	if err != nil {
		return nil, err
	}

	dataURL := "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString([]byte(template))
	var pdfBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(width).
				WithPaperHeight(height).
				WithLandscape(p.Landscape).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuf = buf
			return nil
		}),
	); err != nil {
		return nil, err
	}

	if len(pdfBuf) == 0 {
		return nil, fmt.Errorf("generated PDF is empty")
	}

	return pdfBuf, nil
}

func (p PDFOptions[T]) Generate(ctx context.Context, data T) ([]byte, error) {
	return p.saveHTMLToPDFBytesWithSize(ctx, data)
}
