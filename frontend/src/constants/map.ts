/**
 * 地图配置常量
 */

// 高德地图API Key - 从环境变量获取，如果没有则使用默认值
export const AMAP_API_KEY = import.meta.env.VITE_AMAP_API_KEY || 'YOUR_AMAP_KEY'

// 高德地图API URL
export const AMAP_API_URL = 'https://webapi.amap.com/maps'

// 默认地图配置
export const DEFAULT_MAP_CONFIG = {
	zoom: 11,
	center: [114.375483, 36.105807], // 新坐标
	mapStyle: 'amap://styles/normal',
}
