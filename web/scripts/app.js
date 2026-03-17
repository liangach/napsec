/**
 * NapSec 仪表盘主逻辑
 */
const app = (() => {
    let refreshTimer = null

    function $(id) { return document.getElementById(id) }

    function setText(id, value) {
        const el = $(id)
        if (el) el.textContent = value
    }

    function setStatus(online) {
        const dot  = $('statusDot')
        const text = $('statusText')
        if (online) {
            dot.className  = 'status-dot online'
            text.textContent = 'NapSec 已连接'
        } else {
            dot.className  = 'status-dot offline'
            text.textContent = '未连接'
        }
    }

    // 数据格式化

    function formatTime(ts) {
        if (!ts) return '--'
        const d = new Date(ts)
        return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ` +
            `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    }

    function pad(n) { return String(n).padStart(2, '0') }

    function formatFileName(path) {
        if (!path) return '--'
        return path.split('/').pop() || path
    }

    function opBadge(op) {
        const map = {
            'ENCRYPT': ['badge-encrypt', '加密'],
            'DETECT':  ['badge-detect',  '检测'],
            'RECOVER': ['badge-recover', '恢复'],
        }
        const [cls, label] = map[op] || ['badge-default', op || '--']
        return `<span class="badge ${cls}">${label}</span>`
    }

    function statusBadge(success) {
        return success
            ? '<span class="status-ok">✓ 成功</span>'
            : '<span class="status-fail">✗ 失败</span>'
    }

    // 渲染函数

    /**
     * 渲染统计卡片
     */
    function renderStats(data) {
        const engine  = data.engine  || {}
        const history = data.history || {}

        setText('totalScanned',   engine.files_scanned   ?? '--')
        setText('totalProtected', history.TotalProtected ?? '--')
        setText('todayProtected', history.TodayProtected ?? '--')
        setText('totalLogs',      history.TotalLogs      ?? '--')
    }

    /**
     * 渲染审计记录表格
     * @param {Array} records
     */
    function renderTable(records) {
        const tbody = $('auditTableBody')
        if (!tbody) return

        if (!records || records.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" class="loading">暂无审计记录</td></tr>'
            return
        }

        tbody.innerHTML = records.map(r => `
      <tr>
        <td style="color:var(--text-muted);font-size:12px;">${formatTime(r.timestamp)}</td>
        <td>${opBadge(r.operation)}</td>
        <td title="${r.original_path || ''}" style="
          max-width:260px;
          overflow:hidden;
          text-overflow:ellipsis;
          white-space:nowrap;
        ">${formatFileName(r.original_path)}</td>
        <td style="color:var(--text-muted)">${r.rule_name || '--'}</td>
        <td>${statusBadge(r.success)}</td>
      </tr>
    `).join('')
    }

    /**
     * 从记录中统计操作类型分布
     */
    function calcOpDistribution(records) {
        const dist = { ENCRYPT: 0, DETECT: 0, RECOVER: 0 }
        for (const r of records) {
            if (r.operation in dist) dist[r.operation]++
        }
        return dist
    }

    // 主刷新流程

    async function refresh() {
        try {
            // 健康检查
            await API.health()
            setStatus(true)

            // 并行拉取数据
            const [statsData, records] = await Promise.all([
                API.getStats(),
                API.getRecords(),
            ])

            renderStats(statsData)
            renderTable(records)

            // 图表（每次刷新重绘）
            const dist = calcOpDistribution(records)
            Charts.initPie(dist)
            Charts.initLine(records)

        } catch (err) {
            setStatus(false)
            console.warn('NapSec API 不可用:', err.message)
        }
    }

    //  初始化

    function init() {
        // 首次加载
        refresh()

        // 每 10 秒自动刷新
        refreshTimer = setInterval(refresh, 10_000)
    }

    // DOM 就绪后启动
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init)
    } else {
        init()
    }

    return { refresh }
})()