package service

import (
	"context"
	"fmt"
	"github.com/el-bulk/backend/models"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"io"
	"net/http"
	"strings"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GenerateOrderReceipt(ctx context.Context, detail models.OrderDetail, settings models.Settings) ([]byte, error) {
	m := maroto.New()

	// Header
	headerRow := row.New(20)
	
	// Add logo if available
	logoCol := col.New(2)
	if settings.StoreLogoURL != "" {
		imgBytes, ext, err := s.fetchImage(settings.StoreLogoURL)
		if err == nil {
			logoCol.Add(
				image.NewFromBytes(imgBytes, ext, props.Rect{
					Center:  false,
					Percent: 100,
					Top:     0,
				}),
			)
		}
	}
	headerRow.Add(logoCol)

	// Always add Store Name
	headerRow.Add(
		col.New(6).Add(
			text.New("EL BULK", props.Text{
				Top:   5,
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Left,
			}),
		),
	)

	headerRow.Add(
		col.New(4).Add(
			text.New("RECEIPT", props.Text{
				Top:   5,
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Right,
			}),
		),
	)
	
	m.AddRows(headerRow)

	// Order Info
	m.AddRows(
		row.New(15).Add(
			col.New(6).Add(
				text.New(fmt.Sprintf("Order #: %s", detail.Order.OrderNumber), props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
				text.New(fmt.Sprintf("Date: %s", detail.Order.CreatedAt.Format("Jan 02, 2006")), props.Text{
					Top:  5,
					Size: 10,
				}),
			),
			col.New(6).Add(
				text.New("SHIPPING TO:", props.Text{
					Size:  8,
					Style: fontstyle.Bold,
					Align: align.Right,
				}),
				text.New(fmt.Sprintf("%s %s", detail.Customer.FirstName, detail.Customer.LastName), props.Text{
					Top:   5,
					Size:  10,
					Align: align.Right,
				}),
				text.New(derefString(detail.Customer.Address, "No Address Provided"), props.Text{
					Top:   10,
					Size:  10,
					Align: align.Right,
				}),
			),
		),
	)

	// Spacer
	m.AddRows(row.New(10))

	// Items Table Header
	m.AddRows(
		row.New(10).Add(
			col.New(6).Add(text.New("Item Description", props.Text{Size: 9, Style: fontstyle.Bold})),
			col.New(2).Add(text.New("Qty", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center})),
			col.New(2).Add(text.New("Unit Price", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Right})),
			col.New(2).Add(text.New("Total", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Right})),
		),
	)

	// Items Table Rows
	for _, item := range detail.Items {
		name := item.ProductName
		if item.ProductSet != nil {
			name = fmt.Sprintf("%s (%s)", name, *item.ProductSet)
		}
		
		m.AddRows(
			row.New(8).Add(
				col.New(6).Add(text.New(name, props.Text{Size: 8})),
				col.New(2).Add(text.New(fmt.Sprintf("%d", item.Quantity), props.Text{Size: 8, Align: align.Center})),
				col.New(2).Add(text.New(formatCOP(item.UnitPriceCOP), props.Text{Size: 8, Align: align.Right})),
				col.New(2).Add(text.New(formatCOP(item.UnitPriceCOP*float64(item.Quantity)), props.Text{Size: 8, Align: align.Right})),
			),
		)
	}

	// Spacer
	m.AddRows(row.New(10))

	// Totals
	m.AddRows(
		row.New(6).Add(
			col.New(8).Add(text.New("", props.Text{})),
			col.New(2).Add(text.New("Subtotal:", props.Text{Size: 9, Align: align.Right})),
			col.New(2).Add(text.New(formatCOP(detail.Order.SubtotalCOP), props.Text{Size: 9, Align: align.Right})),
		),
		row.New(6).Add(
			col.New(8).Add(text.New("", props.Text{})),
			col.New(2).Add(text.New("Shipping:", props.Text{Size: 9, Align: align.Right})),
			col.New(2).Add(text.New(formatCOP(detail.Order.ShippingCOP), props.Text{Size: 9, Align: align.Right})),
		),
		row.New(10).Add(
			col.New(8).Add(text.New("", props.Text{})),
			col.New(2).Add(text.New("GRAND TOTAL:", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
			col.New(2).Add(text.New(formatCOP(detail.Order.TotalCOP), props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
		),
	)

	// Footer
	m.AddRows(row.New(20))
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New(settings.ReceiptFooterText, props.Text{
					Size:  8,
					Align: align.Center,
					Color: &props.Color{Red: 120, Green: 120, Blue: 120},
				}),
			),
		),
	)

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document.GetBytes(), nil
}

func derefString(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

func formatCOP(val float64) string {
	return fmt.Sprintf("$%s COP", formatNumber(int64(val)))
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var res []string
	for len(s) > 3 {
		res = append([]string{s[len(s)-3:]}, res...)
		s = s[:len(s)-3]
	}
	if len(s) > 0 {
		res = append([]string{s}, res...)
	}
	return strings.Join(res, ".")
}

func (s *PDFService) fetchImage(url string) ([]byte, extension.Type, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bad status: %s", resp.Status)
	}

	ext := extension.Png
	if strings.HasSuffix(strings.ToLower(url), ".jpg") || strings.HasSuffix(strings.ToLower(url), ".jpeg") {
		ext = extension.Jpg
	}

	data, err := io.ReadAll(resp.Body)
	return data, ext, err
}
