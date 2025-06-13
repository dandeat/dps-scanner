package helpers

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func ExportToExcel(header []string, data [][]interface{}, fileName string) (err error) {

	var sheet = "Sheet1"

	// Define Sheet
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Set Header
	for i, h := range header {
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", string('A'+i), 1), h)
	}

	// Arange data into sheet
	for i, row := range data {
		for j, cell := range row {
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", string('A'+j), i+2), cell)
		}
	}

	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}

	return
}
