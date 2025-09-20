package main

import (
	"fmt"
	"strconv"

	"steamprops/internal/steamprops"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// regionToString returns human-readable region name
func regionToString(r int) string {
	switch r {
	case 1:
		return "Region 1 (Сжатая жидкость)"
	case 2:
		return "Region 2 (Перегретый пар)"
	case 3:
		return "Region 3 (Критическая область)"
	case 4:
		return "Region 4 (Линия насыщения)"
	case 5:
		return "Region 5 (Высокотемпературный газ)"
	default:
		return "Неизвестный регион"
	}
}

// InputPanel представляет панель ввода данных
type InputPanel struct {
	// Основные элементы
	modeSelect       *widget.Select
	temperatureEntry *widget.Entry
	pressureEntry    *widget.Entry
	enthalpyEntry    *widget.Entry
	entropyEntry     *widget.Entry

	// Единицы измерения
	tempUnitSelect     *widget.Select
	pressureUnitSelect *widget.Select

	// Контейнеры
	mainContainer *fyne.Container
	tpContainer   *fyne.Container
	hsContainer   *fyne.Container
}

// NewInputPanel создает новую панель ввода
func NewInputPanel() *InputPanel {
	ip := &InputPanel{}
	ip.setupElements()
	ip.setupLayout()
	ip.setupEventHandlers()
	return ip
}

func (ip *InputPanel) setupElements() {
	// Режим расчета
	ip.modeSelect = widget.NewSelect([]string{"TP", "HS"}, nil)
	ip.modeSelect.SetSelected("TP")

	// Поля ввода
	ip.temperatureEntry = widget.NewEntry()
	ip.temperatureEntry.SetPlaceHolder("20.0")
	ip.temperatureEntry.SetText("20.0")

	ip.pressureEntry = widget.NewEntry()
	ip.pressureEntry.SetPlaceHolder("101325")
	ip.pressureEntry.SetText("101325")

	ip.enthalpyEntry = widget.NewEntry()
	ip.enthalpyEntry.SetPlaceHolder("83.95")
	ip.enthalpyEntry.SetText("83.95")

	ip.entropyEntry = widget.NewEntry()
	ip.entropyEntry.SetPlaceHolder("0.2965")
	ip.entropyEntry.SetText("0.2965")

	// Единицы измерения
	ip.tempUnitSelect = widget.NewSelect([]string{"°C", "K", "°F"}, nil)
	ip.tempUnitSelect.SetSelected("°C")

	ip.pressureUnitSelect = widget.NewSelect([]string{"Pa", "kPa", "MPa", "bar", "atm"}, nil)
	ip.pressureUnitSelect.SetSelected("Pa")
}

func (ip *InputPanel) setupLayout() {
	// TP режим
	ip.tpContainer = container.NewVBox(
		widget.NewCard("Температура", "", container.NewHBox(
			ip.temperatureEntry,
			ip.tempUnitSelect,
		)),
		widget.NewCard("Давление", "", container.NewHBox(
			ip.pressureEntry,
			ip.pressureUnitSelect,
		)),
	)

	// HS режим
	ip.hsContainer = container.NewVBox(
		widget.NewCard("Энтальпия", "", container.NewHBox(
			ip.enthalpyEntry,
			widget.NewLabel("кДж/кг"),
		)),
		widget.NewCard("Энтропия", "", container.NewHBox(
			ip.entropyEntry,
			widget.NewLabel("кДж/(кг·К)"),
		)),
		widget.NewCard("Информация", "", widget.NewLabel("Режим HS работает для Region 3\nДля других регионов используйте режим TP")),
	)

	// Основной контейнер
	ip.mainContainer = container.NewVBox(
		widget.NewCard("Режим расчета", "", ip.modeSelect),
		container.NewMax(ip.tpContainer),
	)
}

func (ip *InputPanel) setupEventHandlers() {
	ip.modeSelect.OnChanged = func(mode string) {
		if mode == "HS" {
			ip.mainContainer.Objects[1] = ip.hsContainer
		} else {
			ip.mainContainer.Objects[1] = ip.tpContainer
		}
		ip.mainContainer.Refresh()
	}
}

func (ip *InputPanel) GetContainer() *fyne.Container {
	return ip.mainContainer
}

func (ip *InputPanel) GetInputs() (*steamprops.InputData, error) {
	mode := ip.modeSelect.Selected

	if mode == "HS" {
		h, err := strconv.ParseFloat(ip.enthalpyEntry.Text, 64)
		if err != nil {
			return nil, fmt.Errorf("неверное значение энтальпии: %v", err)
		}

		s, err := strconv.ParseFloat(ip.entropyEntry.Text, 64)
		if err != nil {
			return nil, fmt.Errorf("неверное значение энтропии: %v", err)
		}

		return &steamprops.InputData{
			Mode:     mode,
			Enthalpy: h,
			Entropy:  s,
		}, nil
	}

	// TP режим
	t, err := strconv.ParseFloat(ip.temperatureEntry.Text, 64)
	if err != nil {
		return nil, fmt.Errorf("неверное значение температуры: %v", err)
	}

	p, err := strconv.ParseFloat(ip.pressureEntry.Text, 64)
	if err != nil {
		return nil, fmt.Errorf("неверное значение давления: %v", err)
	}

	// Конвертация единиц
	t = ip.convertTemperature(t)
	p = ip.convertPressure(p)

	return &steamprops.InputData{
		Mode:        mode,
		Temperature: t,
		Pressure:    p,
	}, nil
}

func (ip *InputPanel) convertTemperature(t float64) float64 {
	switch ip.tempUnitSelect.Selected {
	case "K":
		return t - 273.15
	case "°F":
		return (t - 32) * 5 / 9
	default: // °C
		return t
	}
}

func (ip *InputPanel) convertPressure(p float64) float64 {
	switch ip.pressureUnitSelect.Selected {
	case "kPa":
		return p * 1000
	case "MPa":
		return p * 1000000
	case "bar":
		return p * 100000
	case "atm":
		return p * 101325
	default: // Pa
		return p
	}
}

func (ip *InputPanel) Clear() {
	ip.temperatureEntry.SetText("20.0")
	ip.pressureEntry.SetText("101325")
	ip.enthalpyEntry.SetText("83.95")
	ip.entropyEntry.SetText("0.2965")
}

// ResultsPanel представляет панель результатов
type ResultsPanel struct {
	phaseLabel    *widget.Label
	regionLabel   *widget.Label
	resultsTable  *widget.Table
	mainContainer *fyne.Container
}

// NewResultsPanel создает новую панель результатов
func NewResultsPanel() *ResultsPanel {
	rp := &ResultsPanel{}
	rp.setupElements()
	rp.setupLayout()
	return rp
}

func (rp *ResultsPanel) setupElements() {
	rp.phaseLabel = widget.NewLabel("Фаза: Не определена")
	rp.phaseLabel.TextStyle.Bold = true

	rp.regionLabel = widget.NewLabel("Регион: Не определен")
	rp.regionLabel.TextStyle.Italic = true

	rp.resultsTable = widget.NewTable(
		func() (int, int) { return 0, 2 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			label.SetText("")
		},
	)
	rp.resultsTable.SetColumnWidth(0, 200)
	rp.resultsTable.SetColumnWidth(1, 150)
}

func (rp *ResultsPanel) setupLayout() {
	headerContainer := container.NewHBox(
		rp.phaseLabel,
		rp.regionLabel,
	)

	rp.mainContainer = container.NewVBox(
		widget.NewCard("Результаты расчета", "", container.NewVBox(
			headerContainer,
			rp.resultsTable,
		)),
	)
}

func (rp *ResultsPanel) GetContainer() *fyne.Container {
	return rp.mainContainer
}

func (rp *ResultsPanel) UpdateResults(result *steamprops.Result) {
	rp.phaseLabel.SetText(fmt.Sprintf("Фаза: %s", result.Phase))
	rp.regionLabel.SetText(fmt.Sprintf("Регион: %s", regionToString(int(result.Region))))

	results := [][]string{
		{"Температура", fmt.Sprintf("%.3f °C", result.Temperature)},
		{"Давление", fmt.Sprintf("%.3f MPa (%.0f Pa)", result.Pressure/1e6, result.Pressure)},
		{"Плотность", fmt.Sprintf("%.6g кг/м³", result.Properties.Density)},
		{"Удельный объем", fmt.Sprintf("%.6g м³/кг", result.Properties.SpecificVolume)},
		{"Энтальпия", fmt.Sprintf("%.6g кДж/кг", result.Properties.SpecificEnthalpy)},
		{"Энтропия", fmt.Sprintf("%.6g кДж/(кг·К)", result.Properties.SpecificEntropy)},
		{"Внутренняя энергия", fmt.Sprintf("%.6g кДж/кг", result.Properties.SpecificInternalEnergy)},
		{"Изобарная теплоемкость", fmt.Sprintf("%.6g кДж/(кг·К)", result.Properties.SpecificIsobaricHeatCapacity)},
		{"Изохорная теплоемкость", fmt.Sprintf("%.6g кДж/(кг·К)", result.Properties.SpecificIsochoricHeatCapacity)},
		{"Скорость звука", fmt.Sprintf("%.6g м/с", result.Properties.SpeedOfSound)},
		{"Динамическая вязкость", result.TransportProps["dynamic_viscosity"]},
		{"Кинематическая вязкость", result.TransportProps["kinematic_viscosity"]},
		{"Теплопроводность", result.TransportProps["thermal_conductivity"]},
	}

	rp.resultsTable.UpdateCell = func(id widget.TableCellID, obj fyne.CanvasObject) {
		label := obj.(*widget.Label)
		if id.Row < len(results) && id.Col < len(results[id.Row]) {
			label.SetText(results[id.Row][id.Col])
		}
	}
	rp.resultsTable.Length = func() (int, int) {
		return len(results), 2
	}
	rp.resultsTable.Refresh()
}

func (rp *ResultsPanel) Clear() {
	rp.phaseLabel.SetText("Фаза: Не определена")
	rp.regionLabel.SetText("Регион: Не определен")
	rp.resultsTable.Length = func() (int, int) { return 0, 2 }
	rp.resultsTable.Refresh()
}

// ControlPanel представляет панель управления
type ControlPanel struct {
	calculateButton *widget.Button
	clearButton     *widget.Button
	mainContainer   *fyne.Container
}

// NewControlPanel создает новую панель управления
func NewControlPanel() *ControlPanel {
	cp := &ControlPanel{}
	cp.setupElements()
	cp.setupLayout()
	return cp
}

func (cp *ControlPanel) setupElements() {
	cp.calculateButton = widget.NewButton("Рассчитать", nil)
	cp.calculateButton.Importance = widget.HighImportance

	cp.clearButton = widget.NewButton("Очистить", nil)
}

func (cp *ControlPanel) setupLayout() {
	cp.mainContainer = container.NewHBox(
		cp.calculateButton,
		cp.clearButton,
	)
}

func (cp *ControlPanel) GetContainer() *fyne.Container {
	return cp.mainContainer
}

func (cp *ControlPanel) SetCalculateCallback(callback func()) {
	cp.calculateButton.OnTapped = callback
}

func (cp *ControlPanel) SetClearCallback(callback func()) {
	cp.clearButton.OnTapped = callback
}

// HistoryPanel представляет панель истории расчетов
type HistoryPanel struct {
	list          *widget.List
	items         []string
	clearButton   *widget.Button
	mainContainer *fyne.Container
}

// NewHistoryPanel создает новую панель истории
func NewHistoryPanel() *HistoryPanel {
	h := &HistoryPanel{}
	// Список элементов истории
	h.list = widget.NewList(
		func() int { return len(h.items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= 0 && i < len(h.items) {
				o.(*widget.Label).SetText(h.items[i])
			}
		},
	)
	// Кнопка очистки
	h.clearButton = widget.NewButton("Очистить историю", func() { h.Clear() })
	// Компоновка панели
	h.mainContainer = container.NewVBox(
		widget.NewCard("История расчетов", "Последние результаты сессии", container.NewBorder(nil, h.clearButton, nil, nil, h.list)),
	)
	return h
}

// GetContainer возвращает контейнер панели истории
func (h *HistoryPanel) GetContainer() *fyne.Container { return h.mainContainer }

// AddEntry добавляет запись в историю
func (h *HistoryPanel) AddEntry(s string) {
	if s == "" {
		return
	}
	h.items = append([]string{s}, h.items...) // последние сверху
	h.list.Refresh()
}

// Clear очищает историю
func (h *HistoryPanel) Clear() {
	h.items = nil
	h.list.Refresh()
}
