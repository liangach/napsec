/**
 * NapSec 图表管理
 */
const Charts = (() => {
    let pieChart = null
    let lineChart = null

    const COLORS = {
        blue:   'rgba(78,142,247,0.8)',
        green:  'rgba(46,204,135,0.8)',
        orange: 'rgba(247,147,78,0.8)',
        purple: 'rgba(168,85,247,0.8)',
        red:    'rgba(247,90,90,0.8)',
    }

    const BORDER = {
        blue:   'rgba(78,142,247,1)',
        green:  'rgba(46,204,135,1)',
        orange: 'rgba(247,147,78,1)',
        purple: 'rgba(168,85,247,1)',
        red:    'rgba(247,90,90,1)',
    }

    // 通用图表默认选项
    const baseOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                labels: {
                    color: '#7b82a0',
                    font: { size: 12 },
                    boxWidth: 14,
                    padding: 16,
                }
            },
            tooltip: {
                backgroundColor: '#20242f',
                borderColor: '#2e3347',
                borderWidth: 1,
                titleColor: '#e6e9f0',
                bodyColor: '#7b82a0',
            }
        }
    }

    /**
     * 初始化饼图（操作类型分布）
     * @param {Object} data - { ENCRYPT: n, DETECT: n, RECOVER: n }
     */
    function initPie(data) {
        const ctx = document.getElementById('pieChart')
        if (!ctx) return

        const labels   = ['加密保护', '检测发现', '文件恢复']
        const values   = [
            data.ENCRYPT || 0,
            data.DETECT  || 0,
            data.RECOVER || 0,
        ]

        if (pieChart) pieChart.destroy()

        pieChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels,
                datasets: [{
                    data: values,
                    backgroundColor: [COLORS.blue, COLORS.orange, COLORS.green],
                    borderColor:     [BORDER.blue, BORDER.orange, BORDER.green],
                    borderWidth: 2,
                    hoverOffset: 6,
                }]
            },
            options: {
                ...baseOptions,
                cutout: '65%',
                plugins: {
                    ...baseOptions.plugins,
                    legend: {
                        ...baseOptions.plugins.legend,
                        position: 'bottom',
                    }
                }
            }
        })
    }

    /**
     * 初始化折线图（7天趋势）
     * @param {Array} records - 审计记录数组
     */
    function initLine(records) {
        const ctx = document.getElementById('lineChart')
        if (!ctx) return

        // 生成最近 7 天标签
        const days   = []
        const counts = []

        for (let i = 6; i >= 0; i--) {
            const d = new Date()
            d.setDate(d.getDate() - i)
            const label = `${d.getMonth() + 1}/${d.getDate()}`
            days.push(label)

            // 统计当天 ENCRYPT 数量
            const dayStr = d.toISOString().slice(0, 10)
            const count  = records.filter(r =>
                r.operation === 'ENCRYPT' &&
                r.success   === true &&
                r.timestamp && r.timestamp.startsWith(dayStr)
            ).length

            counts.push(count)
        }

        if (lineChart) lineChart.destroy()

        lineChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: days,
                datasets: [{
                    label: '保护文件数',
                    data: counts,
                    fill: true,
                    backgroundColor: 'rgba(78,142,247,0.1)',
                    borderColor: BORDER.blue,
                    pointBackgroundColor: BORDER.blue,
                    pointRadius: 4,
                    pointHoverRadius: 6,
                    tension: 0.4,
                }]
            },
            options: {
                ...baseOptions,
                scales: {
                    x: {
                        ticks: { color: '#7b82a0', font: { size: 11 } },
                        grid:  { color: 'rgba(46,51,71,0.5)' },
                    },
                    y: {
                        ticks: {
                            color: '#7b82a0',
                            font: { size: 11 },
                            stepSize: 1,
                            precision: 0,
                        },
                        grid: { color: 'rgba(46,51,71,0.5)' },
                        beginAtZero: true,
                    }
                }
            }
        })
    }

    return { initPie, initLine }
})()