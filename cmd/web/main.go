package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/somepgs/steamprops/internal/steamprops"
)

// WebServer представляет веб-сервер приложения
type WebServer struct {
	calculator *steamprops.Calculator
	templates  *template.Template
}

// NewWebServer создает новый веб-сервер
func NewWebServer() *WebServer {
	ws := &WebServer{
		calculator: steamprops.NewCalculator(),
	}

	// Загружаем HTML шаблоны
	ws.templates = template.Must(template.ParseGlob("web/templates/*.html"))

	return ws
}

// CalculationRequest представляет запрос на расчет
type CalculationRequest struct {
	Mode        string  `json:"mode"`
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Enthalpy    float64 `json:"enthalpy"`
	Entropy     float64 `json:"entropy"`
	Region      string  `json:"region"`
}

// CalculationResponse представляет ответ с результатами расчета
type CalculationResponse struct {
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Result     *steamprops.Result     `json:"result,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// handleIndex обрабатывает главную страницу
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := ws.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleCalculate обрабатывает API запросы на расчет
func (ws *WebServer) handleCalculate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := CalculationResponse{
			Success: false,
			Error:   fmt.Sprintf("Ошибка парсинга JSON: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Создаем InputData
	var inputData *steamprops.InputData
	if req.Mode == "HS" {
		inputData = &steamprops.InputData{
			Mode:     req.Mode,
			Enthalpy: req.Enthalpy,
			Entropy:  req.Entropy,
		}
	} else {
		inputData = &steamprops.InputData{
			Mode:        req.Mode,
			Temperature: req.Temperature,
			Pressure:    req.Pressure,
		}
	}

	// Валидируем входные данные
	if err := inputData.Validate(); err != nil {
		response := CalculationResponse{
			Success: false,
			Error:   fmt.Sprintf("Ошибка валидации: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Выполняем расчет
	result, err := ws.calculator.Calculate(inputData)
	if err != nil {
		response := CalculationResponse{
			Success: false,
			Error:   fmt.Sprintf("Ошибка расчета: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Формируем ответ
	properties := map[string]interface{}{
		"temperature":                      result.Temperature,
		"pressure":                         result.Pressure,
		"density":                          result.Properties.Density,
		"specific_volume":                  result.Properties.SpecificVolume,
		"specific_enthalpy":                result.Properties.SpecificEnthalpy,
		"specific_entropy":                 result.Properties.SpecificEntropy,
		"specific_internal_energy":         result.Properties.SpecificInternalEnergy,
		"specific_isobaric_heat_capacity":  result.Properties.SpecificIsobaricHeatCapacity,
		"specific_isochoric_heat_capacity": result.Properties.SpecificIsochoricHeatCapacity,
		"speed_of_sound":                   result.Properties.SpeedOfSound,
		"dynamic_viscosity":                result.TransportProps["dynamic_viscosity"],
		"kinematic_viscosity":              result.TransportProps["kinematic_viscosity"],
		"thermal_conductivity":             result.TransportProps["thermal_conductivity"],
		"phase":                            result.Phase,
		"region":                           result.Region,
	}

	response := CalculationResponse{
		Success:    true,
		Result:     result,
		Properties: properties,
	}

	json.NewEncoder(w).Encode(response)
}

// handleStatic обрабатывает статические файлы
func (ws *WebServer) handleStatic(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))).ServeHTTP(w, r)
}

// Run запускает веб-сервер
func (ws *WebServer) Run(port int) {
	// Настраиваем маршруты
	http.HandleFunc("/", ws.handleIndex)
	http.HandleFunc("/api/calculate", ws.handleCalculate)
	http.HandleFunc("/static/", ws.handleStatic)

	log.Printf("Веб-сервер запущен на порту %d", port)
	log.Printf("Откройте http://localhost:%d в браузере", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

func main() {
	// Получаем порт из аргументов командной строки или используем по умолчанию
	port := 8080
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		}
	}

	server := NewWebServer()
	server.Run(port)
}
