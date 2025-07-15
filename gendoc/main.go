package main

import (
	"baliance.com/gooxml/color"
	"baliance.com/gooxml/measurement"
	"baliance.com/gooxml/schema/soo/wml"
	"fmt"
	"fyne.io/fyne/v2"
	"os"
	"path/filepath"
	"strings"

	"baliance.com/gooxml/document"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/xuri/excelize/v2"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Институт OPEN - Генератор приказов")

	// Название курса
	courseEntry := widget.NewEntry()
	courseEntry.SetPlaceHolder("Введите название курса")

	// Количество часов
	hoursEntry := widget.NewEntry()
	hoursEntry.SetPlaceHolder("Введите количество часов")

	// Дата зачисления
	dateInEntry := widget.NewEntry()
	dateInEntry.SetPlaceHolder("Введите дату зачисления")

	// Дата отчисления
	dateOutEntry := widget.NewEntry()
	dateOutEntry.SetPlaceHolder("Введите дату отчисления")

	// Получаем список Excel-файлов
	excelFiles, err := listExcelFiles("input")
	if err != nil || len(excelFiles) == 0 {
		excelFiles = []string{"Нет Excel файлов в /input"}
	}

	// Select для Excel-файла
	selectedFile := ""
	fileSelect := widget.NewSelect(excelFiles, func(value string) {
		selectedFile = value
	})

	fileSelect.PlaceHolder = "Выберите Excel-файл"

	statusLabel := widget.NewLabel("")

	generateBtn := widget.NewButton("Сгенерировать документы", func() {
		course := courseEntry.Text
		if course == "" {
			statusLabel.SetText("Введите название курса")
			return
		}
		hours := hoursEntry.Text
		if hours == "" {
			statusLabel.SetText("Введите количество часов")
			return
		}
		dateInTxt := dateInEntry.Text
		if dateInTxt == "" {
			statusLabel.SetText("Введите дату зачисления")
			return
		}
		dateOutTxt := dateOutEntry.Text
		if dateOutTxt == "" {
			statusLabel.SetText("Введите дату отчисления")
			return
		}
		if selectedFile == "" {
			statusLabel.SetText("Выберите Excel-файл")
			return
		}

		err := generateDocuments(filepath.Join("input", selectedFile), course, hours, dateInTxt, dateOutTxt)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Ошибка: %v", err))
		} else {
			statusLabel.SetText("Успешно сгенерировано!")
		}
	})

	myWindow.SetContent(container.NewVBox(
		widget.NewLabel("Название курса:"),
		courseEntry,
		widget.NewLabel("Количество часов:"),
		hoursEntry,
		widget.NewLabel("Дата зачисления:"),
		dateInEntry,
		widget.NewLabel("Дата отчисления:"),
		dateOutEntry,
		widget.NewLabel("Excel-файл:"),
		fileSelect,
		widget.NewLabel("Нажмите кнопку для генерации файлов"),
		generateBtn,
		statusLabel,
	))

	myWindow.Resize(fyne.NewSize(500, 250))
	myWindow.ShowAndRun()
}

func generateDocuments(excelPath string, course, hours, dateIn, dateOut string) error {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть Excel: %w", err)
	}

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return fmt.Errorf("в Excel-файле нет листов")
	}
	sheetName := sheetList[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("не удалось прочитать строки: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("Excel-файл должен содержать хотя бы одну строку данных")
	}

	header := rows[0]
	var colIdx = map[string]int{
		"Фамилия":      -1,
		"Имя":          -1,
		"Отчество":     -1,
		"Место работы": -1,
		"Должность":    -1,
	}

	for i, col := range header {
		col = strings.TrimSpace(col)
		if idx, ok := colIdx[col]; ok {
			if idx != -1 {
				continue
			}
			colIdx[col] = i
		}
	}

	for key, idx := range colIdx {
		if idx == -1 {
			return fmt.Errorf("не найдена колонка '%s'", key)
		}
	}

	os.MkdirAll("output", os.ModePerm)

	for _, tmpl := range templates {
		if _, err := os.Stat(tmpl.TemplatePath); os.IsNotExist(err) {
			return fmt.Errorf("шаблон не найден: %s", tmpl.TemplatePath)
		}

		doc, err := document.Open(tmpl.TemplatePath)
		if err != nil {
			return fmt.Errorf("не удалось открыть шаблон %s: %w", tmpl.Name, err)
		}

		// Подставляем общие плейсхолдеры (например {КУРС}, {ЧАСЫ})
		replacePlaceholders(doc, map[string]string{
			"{КУРС}":        course,
			"{ЧАСЫ}":        hours,
			"{ДАТА_ЗАЧИСЛ}": dateIn,
			"{ДАТА_ОТЧИСЛ}": dateOut,
		})

		for _, p := range doc.Paragraphs() {
			for _, r := range p.Runs() {
				if strings.Contains(r.Text(), "{ТАБЛИЦА}") {
					r.ClearContent()

					// Добавляем таблицу ФИО
					insertFioTable(doc, rows[1:], colIdx)
					break
				}
			}
		}

		// Сохраняем
		filename := fmt.Sprintf("%s.docx", tmpl.OutputPrefix)
		outputPath := filepath.Join("output", filename)
		if err := doc.SaveToFile(outputPath); err != nil {
			return fmt.Errorf("не удалось сохранить документ %s: %w", tmpl.Name, err)
		}
	}

	return nil
}

func listExcelFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".xlsx") || strings.HasSuffix(entry.Name(), ".xls")) {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// Вспомогательная функция: безопасно получить ячейку
func getCell(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return row[idx]
	}
	return ""
}

func replacePlaceholders(doc *document.Document, replacements map[string]string) {
	for _, p := range doc.Paragraphs() {
		for _, r := range p.Runs() {
			text := r.Text()
			for old, new := range replacements {
				text = strings.ReplaceAll(text, old, new)
			}
			r.ClearContent()
			r.AddText(text)
		}
	}
}

func insertFioTable(doc *document.Document, dataRows [][]string, colIdx map[string]int) {
	table := doc.AddTable()

	b := table.Properties().Borders()
	th := measurement.Distance(1)
	b.SetAll(wml.ST_BorderSingle, color.Black, th)
	b.SetInsideHorizontal(wml.ST_BorderSingle, color.Black, th)
	b.SetInsideVertical(wml.ST_BorderSingle, color.Black, th)

	// Заголовок
	header := table.AddRow()
	addTableCell(header, "№", true, 32)
	addTableCell(header, "ФИО", true, 0)
	addTableCell(header, "Должность / курс", true, 0)
	addTableCell(header, "Место работы (учебы)", true, 0)

	// Данные
	for i, row := range dataRows {
		fam := strings.TrimSpace(getCell(row, colIdx["Фамилия"]))
		name := strings.TrimSpace(getCell(row, colIdx["Имя"]))
		pat := strings.TrimSpace(getCell(row, colIdx["Отчество"]))
		if fam == "" && name == "" && pat == "" {
			continue
		}
		fullName := fmt.Sprintf("%s %s %s", fam, name, pat)
		workPosition := strings.TrimSpace(getCell(row, colIdx["Должность"]))
		workPlace := strings.TrimSpace(getCell(row, colIdx["Место работы"]))

		r := table.AddRow()
		addTableCell(r, fmt.Sprintf("%d", i+1), false, 8)
		addTableCell(r, fullName, false, 0)
		addTableCell(r, workPosition, false, 0)
		addTableCell(r, workPlace, false, 0)
	}
}

func addTableCell(row document.Row, text string, bold bool, width int64) {
	cell := row.AddCell()

	if width != 0 {
		fmtWidth := measurement.Distance(width)
		cell.Properties().SetWidth(fmtWidth)
	}

	para := cell.AddParagraph()
	run := para.AddRun()
	if bold {
		run.Properties().SetBold(true)
	}
	run.AddText(text)
}
