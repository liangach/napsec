/**
 * NapSec API 客户端
 */
const API = (() => {
    const BASE = 'http://localhost:8080/api'

    async function fetchJSON(endpoint) {
        const res = await fetch(`${BASE}${endpoint}`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
    }

    return {
        /** 获取统计数据 */
        async getStats() {
            return fetchJSON('/stats')
        },

        /** 获取审计记录 */
        async getRecords() {
            return fetchJSON('/records')
        },

        /** 健康检查 */
        async health() {
            return fetchJSON('/health')
        }
    }
})()