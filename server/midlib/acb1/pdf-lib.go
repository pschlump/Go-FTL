package Acb1

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// NOT USED at this point.
// func LoadCSV(path string) [][]string {
// 	f, err := os.Open(path)
// 	if err != nil {
// 		log.Fatalf("Cannot open '%s': %s\n", path, err.Error())
// 	}
// 	defer f.Close()
// 	r := csv.NewReader(f)
// 	rows, err := r.ReadAll()
// 	if err != nil {
// 		log.Fatalln("Cannot read CSV data:", err.Error())
// 	}
// 	return rows
// }

// NewReport retunes an initialized pdf document with a title of 'Document'.
func NewReport() *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "Letter", "") // Set dimentions to 'mm'

	pdf.AddPage()

	pdf.SetFont("Times", "B", 28)

	pdf.Cell(40, 10, "Doucment")

	pdf.Ln(12)

	pdf.SetFont("Times", "", 20)
	pdf.Cell(40, 10, time.Now().Format("Mon Jan 2, 2006  3:04PM"))
	pdf.Ln(20)

	// pjs xyzzy TODO - this is the spot

	return pdf
}

//func Header(pdf *gofpdf.Fpdf, hdr []string) *gofpdf.Fpdf {
//	pdf.SetFont("Times", "B", 16)
//	pdf.SetFillColor(240, 240, 240)
//	for _, str := range hdr {
//		// The `CellFormat()` method takes a couple of parameters to format
//		// the cell. We make use of this to create a visible border around
//		// the cell, and to enable the background fill.
//		pdf.CellFormat(40, 7, str, "1", 0, "", true, 0, "")
//	}
//
//	// Passing `-1` to `Ln()` uses the height of the last printed cell as
//	// the line height.
//	pdf.Ln(-1)
//	return pdf
//}

func Table(pdf *gofpdf.Fpdf, tbl [][]string) *gofpdf.Fpdf {
	pdf.SetFont("Times", "", 16)
	pdf.SetFillColor(255, 255, 255)

	align := []string{"R", "L"}
	for _, line := range tbl {
		for i, str := range line {
			pdf.CellFormat(40, 7, str, "1", 0, align[i], false, 0, "")
		}
		pdf.Ln(-1)
	}
	return pdf
}

//func Table(pdf *gofpdf.Fpdf, tbl [][]string) *gofpdf.Fpdf {
//	pdf.SetFont("Times", "", 16)
//	pdf.SetFillColor(255, 255, 255)
//
//	align := []string{"L", "C", "L", "R", "R", "R"}
//	for _, line := range tbl {
//		for i, str := range line {
//			pdf.CellFormat(40, 7, str, "1", 0, align[i], false, 0, "")
//		}
//		pdf.Ln(-1)
//	}
//	return pdf
//}

// InsertImage adds an image to the document on the current page.  The image will be sized to fit in the page.
func InsertImage(pdf *gofpdf.Fpdf, fn string, pos int) *gofpdf.Fpdf {
	ln := len(fn) // Xyzzy - really should pull of extension - sizlib.dir has a function
	var ext string
	if ln > 4 {
		ext = fn[ln-4:]
	}

	/*
	   A4 measures 210 × 297 millimeters or 8.27 × 11.69 inches. In PostScript, its dimensions
	   are rounded off to 595 × 842 points. Folded twice, an A4 sheet fits in a C6 size envelope
	   (114 × 162 mm).
	*/

	// get image size
	ih, iw := SizeOfImg(fn)
	aspect := ih / iw
	fmt.Printf("aspect = %v\n", aspect)
	// dw, dh := float64(595), float64(842) // size in points
	dw, dh := float64(210), float64(297) // size in mm
	var fw, fh float64
	if ih < (dh-25) && iw < (dw-25) {
		fw, fh = iw, ih
	} else {
		fw = (dw - 25)
		fh = ((dw - 25) * aspect)
	}

	// fmt.Printf("ext ->%s<-, %s\n", ext, godebug.LF())
	if ext == ".PNG" || ext == ".png" {
		if pos == 0 {
			//					 x    y   w   h   flow
			pdf.ImageOptions(fn, 225, 10, 25, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
		} else {
			pdf.ImageOptions(fn, 25, 10, fw, fh, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
		}
	} else {
		pdf.ImageOptions(fn, 25, 10, fw, fh, false, gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}, 0, "")
	}
	return pdf
}

// SavePDF writes the pdf to the specified file name.
func SavePDF(pdf *gofpdf.Fpdf, fn string) error {
	return pdf.OutputFileAndClose(fn)
}

/* vim: set noai ts=4 sw=4: */
