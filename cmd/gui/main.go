package main

import (
	"fmt"
	"steamprops/internal/steamprops"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Application представляет главное приложение SteamProps
type Application struct {
	app    fyne.App
	window fyne.Window

	// Компоненты интерфейса
	inputPanel   *InputPanel
	resultsPanel *ResultsPanel
	controlPanel *ControlPanel
	historyPanel *HistoryPanel

	// Вычислительное ядро
	calculator *steamprops.Calculator
}

// NewApplication создает новое приложение
func NewApplication() *Application {
	app := &Application{
		app:        app.NewWithID("com.steamprops"),
		calculator: steamprops.NewCalculator(),
	}

	app.setupWindow()
	app.setupComponents()
	app.setupLayout()
	app.setupEventHandlers()

	return app
}

func (a *Application) setupWindow() {
	a.window = a.app.NewWindow("SteamProps - IF-97 Calculator")
	a.window.Resize(fyne.NewSize(1000, 700))
	a.window.CenterOnScreen()
	a.window.SetFixedSize(false)
}

func (a *Application) setupComponents() {
	a.inputPanel = NewInputPanel()
	a.resultsPanel = NewResultsPanel()
	a.controlPanel = NewControlPanel()
	a.historyPanel = NewHistoryPanel()
}

func (a *Application) setupLayout() {
	// Основной контейнер с разделением
	mainSplit := container.NewHSplit(
		a.inputPanel.GetContainer(),
		a.resultsPanel.GetContainer(),
	)
	mainSplit.SetOffset(0.4) // 40% для ввода, 60% для результатов

	// Контент вкладки "Калькулятор"
	calculatorContent := container.NewVBox(
		mainSplit,
		widget.NewSeparator(),
		a.controlPanel.GetContainer(),
	)

	// Вкладка "О программе"
	aboutContent := widget.NewCard("О программе", "",
		container.NewVBox(
			widget.NewLabel("SteamProps — калькулятор свойств воды и пара (IAPWS IF-97)."),
			widget.NewLabel("Эта вкладка будет дополнена справкой, ссылками и горячими клавишами."),
		),
	)

	// Вкладки приложения
	tabs := container.NewAppTabs(
		container.NewTabItem("Калькулятор", calculatorContent),
		container.NewTabItem("История", a.historyPanel.GetContainer()),
		container.NewTabItem("О программе", aboutContent),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	a.window.SetContent(tabs)
}

func (a *Application) setupEventHandlers() {
	a.controlPanel.SetCalculateCallback(a.calculate)
	a.controlPanel.SetClearCallback(a.clear)
}

func (a *Application) calculate() {
	// Очищаем предыдущие результаты
	a.resultsPanel.Clear()

	// Получаем входные данные
	inputs, err := a.inputPanel.GetInputs()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Валидируем входные данные
	if err := inputs.Validate(); err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Выполняем расчет
	result, err := a.calculator.Calculate(inputs)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// Отображаем результаты
	a.resultsPanel.UpdateResults(result)

	// Добавляем запись в историю (краткое резюме)
	var in string
	if inputs.Mode == "TP" {
		in = fmt.Sprintf("T=%.3f°C, p=%.0f Pa", inputs.Temperature, inputs.Pressure)
	} else {
		in = fmt.Sprintf("h=%.3f kJ/kg, s=%.3f kJ/(kg·K)", inputs.Enthalpy, inputs.Entropy)
	}
	summary := fmt.Sprintf("%s | %s | %s | %s", time.Now().Format("15:04:05"), inputs.Mode, in, regionToString(int(result.Region)))
	if a.historyPanel != nil {
		a.historyPanel.AddEntry(summary)
	}
}

func (a *Application) clear() {
	a.resultsPanel.Clear()
	a.inputPanel.Clear()
}

func (a *Application) Run() {
	a.window.ShowAndRun()
}

func main() {
	app := NewApplication()
	app.Run()
}
