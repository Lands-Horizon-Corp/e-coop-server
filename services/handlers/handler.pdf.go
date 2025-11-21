package handlers

import (
	"context"
	"encoding/base64"
	"log"
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
	w, err := UnitToInches(p.Width, p.Unit)
	if err != nil {
		return 0, 0, err
	}
	h, err := UnitToInches(p.Height, p.Unit)
	if err != nil {
		return 0, 0, err
	}
	if p.Landscape {
		return h, w, nil
	}
	return w, h, nil
}

func (p PDFOptions[T]) processTemplate(data T) string {
	out, err := raymond.Render(p.Template, data)
	if err != nil {
		log.Fatalf("render error: %v", err)
	}
	return out
}

func (p PDFOptions[T]) saveHTMLToPDFBytesWithSize(parentContext context.Context, data T) ([]byte, error) {
	template := p.processTemplate(data)
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
	return pdfBuf, nil
}

func (p PDFOptions[T]) Generate(cotnext context.Context, data T) ([]byte, error) {

	return p.saveHTMLToPDFBytesWithSize(cotnext, data)
}
