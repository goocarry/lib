package main

type TemplateSpec struct {
	Name         string // Название (для логов)
	TemplatePath string // Путь до шаблона .docx
	OutputPrefix string // Префикс для имени файла
}

var templates = []TemplateSpec{
	{
		Name:         "Приказ о зачислении",
		TemplatePath: "assets/Prikaz_za.docx",
		OutputPrefix: "Prikaz_zachislenie",
	},
	{
		Name:         "Ведомость",
		TemplatePath: "assets/Vedomost.docx",
		OutputPrefix: "Vedomost",
	},
	{
		Name:         "Приказ об отчислении",
		TemplatePath: "assets/Prikaz_ot.docx",
		OutputPrefix: "Prikaz_otchislenie",
	},
}
