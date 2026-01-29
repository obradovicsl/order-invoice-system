package service

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type InvoiceService struct {
	logger *slog.Logger
}

type InvoiceGeneratorInput struct {
	OrderID   string
	Timestamp time.Time
}

func NewInvoiceService(logger *slog.Logger) *InvoiceService {
	return &InvoiceService{
		logger: logger,
	}
}

// GenerateInvoicePDF generates a PDF invoice from an order
func (is *InvoiceService) GenerateInvoicePDF(orderData *OrderResponse) ([]byte, error) {
	is.logger.Info("generating invoice PDF", "order_id", orderData.ID.String())

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 10, "INVOICE", "", 1, "C", false, 0, "")

	// Order details
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 7, "Order ID:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 7, orderData.ID.String(), "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 7, "Customer:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 7, orderData.UserName, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 7, "Date:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 7, orderData.CreatedAt.Time.Format("2006-01-02"), "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 7, "Status:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 7, orderData.Status, "", 1, "L", false, 0, "")

	// Items table
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 10)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(60, 8, "Item Name", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 8, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 8, "Total", "1", 1, "R", true, 0, "")

	// Table rows
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)

	totalAmount := 0.0

	for _, item := range orderData.Items {
		pdf.CellFormat(60, 7, item.ItemName, "1", 0, "L", false, 0, "")

		pdf.CellFormat(30, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")

		priceFloat := float64(item.PriceAtOrder.Int.Int64()) / 100.0
		pdf.CellFormat(40, 7, fmt.Sprintf("$%.2f", priceFloat), "1", 0, "R", false, 0, "")

		itemTotal := priceFloat * float64(item.Quantity)
		pdf.CellFormat(40, 7, fmt.Sprintf("$%.2f", itemTotal), "1", 1, "R", false, 0, "")

		totalAmount += itemTotal
	}

	// Total
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(130, 8, "Total:", "T", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("$%.2f", totalAmount), "TR", 1, "R", false, 0, "")

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 5, "Thank you for your business!", "", 1, "C", false, 0, "")

	// Generate PDF bytes
	var buf bytes.Buffer
	pdf.Output(&buf)

	is.logger.Info("invoice PDF generated successfully", "order_id", orderData.ID.String(), "size", buf.Len())

	return buf.Bytes(), nil
}

// InvoiceFileName generates a filename for the invoice
func (is *InvoiceService) InvoiceFileName(orderID string) string {
	return fmt.Sprintf("invoice_%s.pdf", orderID)
}
