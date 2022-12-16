package excelize

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"testing"

	"github.com/xuri/excelize/v2"
)

// go test -v -run TestHelloWorld
func TestHelloWorld(t *testing.T) {
	f := excelize.NewFile()
	// 创建工作表
	index := f.NewSheet("Sheet2")
	// 设置单元格数据
	f.SetCellValue("Sheet2", "A2", "Hello world.")
	f.SetCellValue("Sheet1", "B2", 100)
	// 设置工作簿的活动工作表
	f.SetActiveSheet(index)
	// 按给定路径保存电子表格
	if err := f.SaveAs("Book1.xlsx"); err != nil {
		fmt.Println(err)
	}
}

// go test -v -run TestEnc
func TestEnc(t *testing.T) {
	f := excelize.NewFile()
	f.NewSheet("Sheet1")
	if err := f.SaveAs("Book-ENC.xlsx", excelize.Options{
		Password: "123",
	}); err != nil {
		fmt.Println(err)
	}
}

// go test -v -run TestReadBook
func TestReadBook(t *testing.T) {
	f, err := excelize.OpenFile("Book1.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// 获取工作表对应单元格数据
	cell, err := f.GetCellValue("Sheet1", "B2")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cell)
	// 获取 Sheet1 中的所有行
	rows, err := f.GetRows("Sheet2")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}

// go test -v -run TestAddChart
func TestAddChart(t *testing.T) {
	categories := map[string]string{
		"A2": "Small", "A3": "Normal", "A4": "Large",
		"B1": "Apple", "C1": "Orange", "D1": "Pear"}
	values := map[string]int{
		"B2": 2, "C2": 3, "D2": 3, "B3": 5, "C3": 2, "D3": 4, "B4": 6, "C4": 7, "D4": 8}
	f := excelize.NewFile()
	for k, v := range categories {
		f.SetCellValue("Sheet1", k, v)
	}
	for k, v := range values {
		f.SetCellValue("Sheet1", k, v)
	}
	if err := f.AddChart("Sheet1", "E1", `{
        "type": "col3DClustered",
        "series": [
        {
            "name": "Sheet1!$A$2",
            "categories": "Sheet1!$B$1:$D$1",
            "values": "Sheet1!$B$2:$D$2"
        },
        {
            "name": "Sheet1!$A$3",
            "categories": "Sheet1!$B$1:$D$1",
            "values": "Sheet1!$B$3:$D$3"
        },
        {
            "name": "Sheet1!$A$4",
            "categories": "Sheet1!$B$1:$D$1",
            "values": "Sheet1!$B$4:$D$4"
        }],
        "title":
        {
            "name": "Fruit 3D Clustered Column Chart"
        }
    }`); err != nil {
		fmt.Println(err)
		return
	}
	// Save spreadsheet by the given path.
	if err := f.SaveAs("Book2.xlsx"); err != nil {
		fmt.Println(err)
	}
}

// go test -v -run TestAddPicture
func TestAddPicture(t *testing.T) {
	/* 下面3个库的引入必不可少,因为注册了全局图片读取方法
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	*/
	f, err := excelize.OpenFile("Book1.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Insert a picture.
	if err := f.AddPicture("Sheet1", "A2", "image.png",
		`{"x_scale": 0.5, "y_scale": 0.5}`); err != nil {
		fmt.Println("image.png", err)
	}
	// Insert a picture to worksheet with scaling.
	if err := f.AddPicture("Sheet1", "D2", "image.jpg",
		`{"x_scale": 0.5, "y_scale": 0.5}`); err != nil {
		fmt.Println("image.jpg", err)
	}
	// Insert a picture offset in the cell with printing support.
	if err := f.AddPicture("Sheet1", "H2", "image.gif", `{
        "x_offset": 15,
        "y_offset": 10,
        "print_obj": true,
        "lock_aspect_ratio": false,
        "locked": false
    }`); err != nil {
		fmt.Println("image.gif", err)
	}
	// Save the spreadsheet with the origin path.
	if err = f.Save(); err != nil {
		fmt.Println(err)
	}
}
