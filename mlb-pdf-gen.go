package main

import (
	"errors"
	"log"
	"os"

	"codeberg.org/go-pdf/fpdf"
)

func FindPageLength(datastring string) int {
	newInit := fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size:           fpdf.SizeType{Wd: 80, Ht: 10000},
		FontDirStr:     "/usr/share/fonts",
	}
	pdf := fpdf.NewCustom(&newInit)
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()
	pdf.AddUTF8Font("FreeMono", "", "liberation/LiberationMono-Regular.ttf")
	pdf.SetFont("FreeMono", "", 8)
	databytes := []byte(datastring)
	lines := pdf.SplitLines(databytes, 74)
	return len(lines)
}

func CreateReportDirs(config ConfigData) {
	if _, err := os.Stat(config.PagePath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(config.PagePath, 0750)
		if err != nil {
			log.Fatal("failed to create dir", config.PagePath)
		}
	}
	if _, err := os.Stat(config.ReceiptPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(config.ReceiptPath, 0750)
		if err != nil {
			log.Fatal("failed to create dir", config.ReceiptPath)
		}
	}
}

func GenerateReportPDF(datastring string, filename string, config ConfigData) {
	pdf := fpdf.New("P", "pt", "Letter", "/usr/share/fonts")
	pdf.AddPage()
	pdf.AddUTF8Font("FreeMono", "", "liberation/LiberationMono-Regular.ttf")
	pdf.SetFont("FreeMono", "", 8)
	pdf.MultiCell(0, 8, datastring, "", "L", false)
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		log.Fatal("FAILURE TO WRITE PDF OUTPUT", err)
	}
}

func GenerateReportPDFReceipt(datastring string, filename string, config ConfigData) {

	pageLength := FindPageLength(datastring)
	newInit := fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size:           fpdf.SizeType{Wd: 80, Ht: float64(float64(pageLength) * 3.25)},
		FontDirStr:     "/usr/share/fonts",
	}
	pdf := fpdf.NewCustom(&newInit)
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()
	pdf.AddUTF8Font("FreeMono", "", "liberation/LiberationMono-Regular.ttf")
	pdf.SetFont("FreeMono", "", 8)
	pdf.MultiCell(80, 3, datastring, "", "L", false)
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		log.Fatal("FAILURE TO WRITE PDF OUTPUT", err)
	}
}
