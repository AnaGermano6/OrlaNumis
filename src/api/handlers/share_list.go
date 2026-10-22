package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

// ShareListHandler makes the lists a collector sends to others: the
// duplicates to trade and the coins wanted, as a PDF or a CSV.
type ShareListHandler struct {
	svc       *services.ShareListService
	uploadDir string
	logger    *services.Logger
}

// NewShareListHandler creates the handler.
func NewShareListHandler(svc *services.ShareListService, uploadDir string, logger *services.Logger) *ShareListHandler {
	return &ShareListHandler{svc: svc, uploadDir: uploadDir, logger: logger}
}

// Download makes a trade list, a wanted list, or both, to send to other collectors.
//
//	@Summary		Download a list to share
//	@Description	Makes, as a PDF (with a photo of each item) or a CSV, the user's duplicates to trade (coins and banknotes held more than once, with how many there are to spare), the wishlist, or both; sorted by country, year and name, in Portuguese or English. Values are included only when asked for, and the contact, when given, is printed at the top and not stored.
//	@Tags			Coins
//	@Accept			json
//	@Produce		application/pdf
//	@Produce		text/csv
//	@Param			body	body		services.ShareListRequest	true	"Which list and how"
//	@Success		200		"The file"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/share-list [post]
func (h *ShareListHandler) Download(c *gin.Context) {
	var request services.ShareListRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Say which list (trades, wishlist or both) and which format (pdf or csv)"})
		return
	}
	request, err := services.NormalizeShareListRequest(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	list, err := h.svc.Build(c.GetUint("userId"), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make the list"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, list.FileName(request.List, request.Format)))
	if request.Format == "csv" {
		body, err := list.CSV()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make the list"})
			return
		}
		c.Data(http.StatusOK, "text/csv; charset=utf-8", body)
		return
	}
	pdf := writeShareListPDF(list, h.uploadDir, h.logger)
	c.Header("Content-Type", "application/pdf")
	if err := pdf.Output(c.Writer); err != nil {
		h.logger.Error("pdf", "Failed to write the shared list: %v", err)
	}
}

const (
	shareRowH   = 17.0
	sharePhotoW = 16.0
	shareGradeW = 16.0
	shareQtyW   = 18.0
	shareValueW = 22.0
)

// writeShareListPDF lays a shared list out on A4 pages, light for printing:
// a title with the date (and the contact), then each list as a table, an
// item a row with its photo.
func writeShareListPDF(list services.ShareList, uploadDir string, logger *services.Logger) *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	labels := list.Labels
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(false, margin)
	pdf.AliasNbPages("{nb}")
	pdf.SetTitle(list.Title, true)
	pdf.SetCreator("Orla Numis", true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(130, 130, 130)
		pdf.CellFormat(contentW/2, 5, tr(labels.Footer+" · "+list.Title), "", 0, "L", false, 0, "")
		page := strings.ReplaceAll(labels.Page, "{page}", strconv.Itoa(pdf.PageNo()))
		pdf.CellFormat(contentW/2, 5, tr(page), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetTextColor(darkR, darkG, darkB)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(contentW, 10, tr(list.Title), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(100, 100, 104)
	date := list.Date.Format("02/01/2006")
	if list.Lang != "pt" {
		date = list.Date.Format("2 January 2006")
	}
	pdf.CellFormat(contentW, 5, tr(labels.Date+": "+date), "", 1, "L", false, 0, "")
	if list.Contact != "" {
		pdf.CellFormat(contentW, 5, tr(labels.Contact+": "+list.Contact), "", 1, "L", false, 0, "")
	}
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(0.6)
	pdf.Line(margin, pdf.GetY()+2, pageW-margin, pdf.GetY()+2)
	pdf.Ln(6)

	for _, section := range list.Sections {
		writeShareSection(pdf, tr, list, section, uploadDir, logger)
	}
	return pdf
}

func writeShareSection(pdf *fpdf.Fpdf, tr func(string) string, list services.ShareList, section services.ShareListSection, uploadDir string, logger *services.Logger) {
	labels := list.Labels
	trades := section.Kind == services.ShareListTrades
	if pdf.GetY()+30 > pageH-20 {
		pdf.AddPage()
	}
	pdf.SetTextColor(darkR, darkG, darkB)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(contentW, 8, tr(fmt.Sprintf("%s (%d)", section.Title, len(section.Items))), "", 1, "L", false, 0, "")
	if section.Intro != "" {
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(100, 100, 104)
		pdf.CellFormat(contentW, 5, tr(section.Intro), "", 1, "L", false, 0, "")
	}
	pdf.Ln(2)
	if len(section.Items) == 0 {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.CellFormat(contentW, 8, tr(labels.Empty), "", 1, "L", false, 0, "")
		pdf.Ln(6)
		return
	}

	textW := contentW - sharePhotoW - shareGradeW
	if trades {
		textW -= shareQtyW
	}
	if list.Values {
		textW -= shareValueW
	}
	header := func() {
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetFillColor(244, 235, 213)
		pdf.SetTextColor(darkR, darkG, darkB)
		pdf.CellFormat(sharePhotoW, 6, "", "", 0, "L", true, 0, "")
		pdf.CellFormat(textW, 6, tr(labels.Item), "", 0, "L", true, 0, "")
		pdf.CellFormat(shareGradeW, 6, tr(labels.Grade), "", 0, "C", true, 0, "")
		if trades {
			pdf.CellFormat(shareQtyW, 6, tr(labels.Available), "", 0, "C", true, 0, "")
		}
		if list.Values {
			pdf.CellFormat(shareValueW, 6, tr(labels.Value), "", 0, "R", true, 0, "")
		}
		pdf.Ln(6)
	}
	header()

	for i, item := range section.Items {
		if pdf.GetY()+shareRowH > pageH-18 {
			pdf.AddPage()
			header()
		}
		x, y := margin, pdf.GetY()
		if i%2 == 1 {
			pdf.SetFillColor(250, 248, 243)
			pdf.Rect(x, y, contentW, shareRowH, "F")
		}
		if photo := shareListPhotoPath(uploadDir, item.Photo); photo != "" {
			placeShareImage(pdf, photo, x+1, y+1.5, sharePhotoW-2, shareRowH-3, logger)
		}

		textX := x + sharePhotoW
		pdf.SetXY(textX, y+1.5)
		pdf.SetTextColor(darkR, darkG, darkB)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(textW, 5, fitText(pdf, tr(item.Name), textW-2), "", 2, "L", false, 0, "")
		details := []string{}
		for _, part := range []string{item.Country, item.Denomination, item.Year} {
			if part != "" {
				details = append(details, part)
			}
		}
		if item.Banknote {
			details = append(details, labels.Banknote)
		}
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(90, 90, 94)
		pdf.CellFormat(textW, 4.5, fitText(pdf, tr(strings.Join(details, " · ")), textW-2), "", 2, "L", false, 0, "")
		if item.References != "" {
			pdf.SetFont("Helvetica", "", 8)
			pdf.SetTextColor(130, 130, 134)
			pdf.CellFormat(textW, 4, fitText(pdf, tr(item.References), textW-2), "", 2, "L", false, 0, "")
		}

		pdf.SetXY(textX+textW, y)
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(darkR, darkG, darkB)
		pdf.CellFormat(shareGradeW, shareRowH, tr(item.Grade), "", 0, "C", false, 0, "")
		if trades {
			pdf.SetFont("Helvetica", "B", 10)
			pdf.CellFormat(shareQtyW, shareRowH, strconv.Itoa(item.Available), "", 0, "C", false, 0, "")
		}
		if list.Values {
			value := ""
			if item.Value != nil {
				value = services.FormatShareValue(*item.Value, list.Currency, list.Lang)
			}
			pdf.SetFont("Helvetica", "", 10)
			pdf.CellFormat(shareValueW, shareRowH, tr(value), "", 0, "R", false, 0, "")
		}
		pdf.SetXY(margin, y+shareRowH)
	}
	pdf.Ln(6)
}

// shareListPhotoPath is the small copy of a coin's photo on disk (its full
// photo when there is none), or "" when it is missing or outside the folder.
func shareListPhotoPath(uploadDir, photo string) string {
	if photo == "" {
		return ""
	}
	root := filepath.Clean(uploadDir)
	path := filepath.Clean(filepath.Join(root, filepath.FromSlash(photo)))
	if !strings.HasPrefix(path, root+string(filepath.Separator)) {
		return ""
	}
	thumb := strings.TrimSuffix(path, filepath.Ext(path)) + "_thumb.jpg"
	for _, candidate := range []string{thumb, path} {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

// placeShareImage draws a photo fitted in a box; a photo that cannot be
// read is left out, and the rest of the PDF is still made.
func placeShareImage(pdf *fpdf.Fpdf, path string, x, y, w, h float64, logger *services.Logger) {
	var kind string
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		kind = "JPEG"
	case ".png":
		kind = "PNG"
	default:
		return
	}
	options := fpdf.ImageOptions{ImageType: kind}
	info := pdf.RegisterImageOptions(path, options)
	if pdf.Err() || info == nil || info.Width() == 0 || info.Height() == 0 {
		if pdf.Err() {
			logger.Warn("pdf", "Skipping a photo of the shared list: %v", pdf.Error())
			pdf.ClearError()
		}
		return
	}
	scale := min(w/info.Width(), h/info.Height())
	fw, fh := info.Width()*scale, info.Height()*scale
	pdf.ImageOptions(path, x+(w-fw)/2, y+(h-fh)/2, fw, fh, false, options, 0, "")
}

// fitText shortens a text (already in the PDF's encoding) to a width, with an ellipsis.
func fitText(pdf *fpdf.Fpdf, text string, width float64) string {
	if pdf.GetStringWidth(text) <= width {
		return text
	}
	const ellipsis = "\x85" // … in the PDF's single-byte encoding
	for len(text) > 0 && pdf.GetStringWidth(text+ellipsis) > width {
		text = text[:len(text)-1]
	}
	return strings.TrimRight(text, " ") + ellipsis
}
