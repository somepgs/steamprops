// SteamProps Web UI JavaScript

class SteamPropsApp {
    constructor() {
        this.chart = null;
        this.history = JSON.parse(localStorage.getItem('steamprops_history') || '[]');
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.updateHistoryDisplay();
        this.initChart();
    }

    setupEventListeners() {
        // Переключение режима расчета
        document.getElementById('mode').addEventListener('change', (e) => {
            this.toggleMode(e.target.value);
        });

        // Кнопки
        document.getElementById('calculate-btn').addEventListener('click', () => {
            this.calculate();
        });

        document.getElementById('clear-btn').addEventListener('click', () => {
            this.clear();
        });

        document.getElementById('clear-history-btn').addEventListener('click', () => {
            this.clearHistory();
        });

        // Конвертация единиц
        document.getElementById('temp-unit').addEventListener('change', () => {
            this.convertTemperature();
        });

        document.getElementById('pressure-unit').addEventListener('change', () => {
            this.convertPressure();
        });
    }

    toggleMode(mode) {
        const tpInputs = document.getElementById('tp-inputs');
        const hsInputs = document.getElementById('hs-inputs');

        if (mode === 'TP') {
            tpInputs.style.display = 'block';
            hsInputs.style.display = 'none';
        } else {
            tpInputs.style.display = 'none';
            hsInputs.style.display = 'block';
        }
    }

    convertTemperature() {
        const tempInput = document.getElementById('temperature');
        const unit = document.getElementById('temp-unit').value;
        const value = parseFloat(tempInput.value);

        if (isNaN(value)) return;

        let celsius;
        switch (unit) {
            case 'K':
                celsius = value - 273.15;
                break;
            case 'F':
                celsius = (value - 32) * 5 / 9;
                break;
            default: // C
                celsius = value;
        }

        // Обновляем значение для отправки на сервер
        tempInput.dataset.celsius = celsius;
    }

    convertPressure() {
        const pressureInput = document.getElementById('pressure');
        const unit = document.getElementById('pressure-unit').value;
        const value = parseFloat(pressureInput.value);

        if (isNaN(value)) return;

        let pascal;
        switch (unit) {
            case 'kPa':
                pascal = value * 1000;
                break;
            case 'MPa':
                pascal = value * 1000000;
                break;
            case 'bar':
                pascal = value * 100000;
                break;
            case 'atm':
                pascal = value * 101325;
                break;
            default: // Pa
                pascal = value;
        }

        // Обновляем значение для отправки на сервер
        pressureInput.dataset.pascal = pascal;
    }

    async calculate() {
        this.showLoading(true);

        try {
            const mode = document.getElementById('mode').value;
            const region = document.getElementById('region').value;

            let requestData;

            if (mode === 'HS') {
                const enthalpy = parseFloat(document.getElementById('enthalpy').value);
                const entropy = parseFloat(document.getElementById('entropy').value);

                if (isNaN(enthalpy) || isNaN(entropy)) {
                    throw new Error('Пожалуйста, введите корректные значения энтальпии и энтропии');
                }

                requestData = {
                    mode: mode,
                    enthalpy: enthalpy,
                    entropy: entropy,
                    region: region
                };
            } else {
                // Конвертируем единицы перед отправкой
                this.convertTemperature();
                this.convertPressure();

                const temperature = parseFloat(document.getElementById('temperature').dataset.celsius || 
                                             document.getElementById('temperature').value);
                const pressure = parseFloat(document.getElementById('pressure').dataset.pascal || 
                                           document.getElementById('pressure').value);

                if (isNaN(temperature) || isNaN(pressure)) {
                    throw new Error('Пожалуйста, введите корректные значения температуры и давления');
                }

                requestData = {
                    mode: mode,
                    temperature: temperature,
                    pressure: pressure,
                    region: region
                };
            }

            const response = await fetch('/api/calculate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestData)
            });

            const result = await response.json();

            if (result.success) {
                this.displayResults(result);
                this.addToHistory(requestData, result);
                this.updateChart(result.properties);
            } else {
                this.showNotification(result.error, 'error');
            }

        } catch (error) {
            this.showNotification('Ошибка соединения: ' + error.message, 'error');
        } finally {
            this.showLoading(false);
        }
    }

    displayResults(result) {
        const properties = result.properties;

        // Обновляем статус
        document.getElementById('phase').textContent = properties.phase || '-';
        document.getElementById('region-result').textContent = this.getRegionName(properties.region);

        // Обновляем результаты
        const resultsContainer = document.getElementById('results');
        resultsContainer.innerHTML = '';

        const resultItems = [
            { label: 'Температура', value: `${properties.temperature.toFixed(3)} °C` },
            { label: 'Давление', value: `${(properties.pressure / 1e6).toFixed(3)} MPa` },
            { label: 'Плотность', value: `${properties.density.toExponential(3)} кг/м³` },
            { label: 'Удельный объем', value: `${properties.specific_volume.toExponential(3)} м³/кг` },
            { label: 'Энтальпия', value: `${properties.specific_enthalpy.toFixed(3)} кДж/кг` },
            { label: 'Энтропия', value: `${properties.specific_entropy.toFixed(3)} кДж/(кг·К)` },
            { label: 'Внутренняя энергия', value: `${properties.specific_internal_energy.toFixed(3)} кДж/кг` },
            { label: 'Изобарная теплоемкость', value: `${properties.specific_isobaric_heat_capacity.toFixed(3)} кДж/(кг·К)` },
            { label: 'Изохорная теплоемкость', value: `${properties.specific_isochoric_heat_capacity.toFixed(3)} кДж/(кг·К)` },
            { label: 'Скорость звука', value: `${properties.speed_of_sound.toFixed(1)} м/с` },
            { label: 'Динамическая вязкость', value: properties.dynamic_viscosity },
            { label: 'Кинематическая вязкость', value: properties.kinematic_viscosity },
            { label: 'Теплопроводность', value: properties.thermal_conductivity }
        ];

        resultItems.forEach(item => {
            const resultItem = document.createElement('div');
            resultItem.className = 'result-item fade-in';
            resultItem.innerHTML = `
                <div class="result-label">${item.label}</div>
                <div class="result-value">${item.value}</div>
            `;
            resultsContainer.appendChild(resultItem);
        });
    }

    getRegionName(region) {
        const regionNames = {
            1: 'Region 1 (Сжатая жидкость)',
            2: 'Region 2 (Перегретый пар)',
            3: 'Region 3 (Критическая область)',
            4: 'Region 4 (Линия насыщения)',
            5: 'Region 5 (Высокотемпературный газ)'
        };
        return regionNames[region] || 'Неизвестный регион';
    }

    updateChart(properties) {
        if (!this.chart) return;

        const labels = ['Плотность', 'Энтальпия', 'Энтропия', 'Теплоемкость'];
        const values = [
            properties.density,
            properties.specific_enthalpy,
            properties.specific_entropy,
            properties.specific_isobaric_heat_capacity
        ];

        this.chart.data.labels = labels;
        this.chart.data.datasets[0].data = values;
        this.chart.update();
    }

    initChart() {
        const ctx = document.getElementById('propertiesChart').getContext('2d');
        this.chart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: ['Плотность', 'Энтальпия', 'Энтропия', 'Теплоемкость'],
                datasets: [{
                    label: 'Значения свойств',
                    data: [0, 0, 0, 0],
                    backgroundColor: 'rgba(37, 99, 235, 0.8)',
                    borderColor: 'rgba(37, 99, 235, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Значение'
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    addToHistory(request, result) {
        const timestamp = new Date().toLocaleTimeString();
        const mode = request.mode;
        
        let inputStr;
        if (mode === 'TP') {
            inputStr = `T=${request.temperature.toFixed(1)}°C, p=${(request.pressure/1000).toFixed(0)}kPa`;
        } else {
            inputStr = `h=${request.enthalpy.toFixed(1)}kJ/kg, s=${request.entropy.toFixed(3)}kJ/(kg·K)`;
        }

        const historyItem = {
            timestamp,
            mode,
            input: inputStr,
            region: result.properties.region,
            phase: result.properties.phase
        };

        this.history.unshift(historyItem);
        
        // Ограничиваем историю 50 записями
        if (this.history.length > 50) {
            this.history = this.history.slice(0, 50);
        }

        localStorage.setItem('steamprops_history', JSON.stringify(this.history));
        this.updateHistoryDisplay();
    }

    updateHistoryDisplay() {
        const historyList = document.getElementById('history-list');
        historyList.innerHTML = '';

        if (this.history.length === 0) {
            historyList.innerHTML = '<div class="history-item">История пуста</div>';
            return;
        }

        this.history.forEach(item => {
            const historyItem = document.createElement('div');
            historyItem.className = 'history-item';
            historyItem.innerHTML = `
                <div><strong>${item.timestamp}</strong> - ${item.mode}</div>
                <div>${item.input} | ${this.getRegionName(item.region)}</div>
            `;
            historyList.appendChild(historyItem);
        });
    }

    clearHistory() {
        this.history = [];
        localStorage.removeItem('steamprops_history');
        this.updateHistoryDisplay();
    }

    clear() {
        document.getElementById('temperature').value = '200';
        document.getElementById('pressure').value = '101325';
        document.getElementById('enthalpy').value = '2000';
        document.getElementById('entropy').value = '5';
        document.getElementById('mode').value = 'TP';
        document.getElementById('region').value = 'auto';
        
        this.toggleMode('TP');
        
        // Очищаем результаты
        document.getElementById('phase').textContent = '-';
        document.getElementById('region-result').textContent = '-';
        document.getElementById('results').innerHTML = '';
        
        // Сбрасываем график
        if (this.chart) {
            this.chart.data.datasets[0].data = [0, 0, 0, 0];
            this.chart.update();
        }
    }

    showLoading(show) {
        document.getElementById('loading').style.display = show ? 'flex' : 'none';
    }

    showNotification(message, type = 'success') {
        const notifications = document.getElementById('notifications');
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        notifications.appendChild(notification);
        
        // Автоматически удаляем уведомление через 5 секунд
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }
}

// Инициализация приложения
document.addEventListener('DOMContentLoaded', () => {
    new SteamPropsApp();
});
