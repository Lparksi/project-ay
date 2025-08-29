<template>
	<section class="page">
		<h1>{{ $t('navigation.map') }}</h1>
		<div id="map-container" />
	</section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { AMAP_API_KEY, AMAP_API_URL, DEFAULT_MAP_CONFIG } from '@/constants/map'

onMounted(() => {
	// 创建 script 标签加载高德地图 API
	const script = document.createElement('script')
	script.type = 'text/javascript'
	script.src = `${AMAP_API_URL}?v=2.0&key=${AMAP_API_KEY}`
	script.onload = () => {
		// @ts-ignore
		const map = new AMap.Map('map-container', DEFAULT_MAP_CONFIG)
	}
	document.head.appendChild(script)
})
</script>

<style scoped>
.page {
	padding: 1rem;
}

#map-container {
	width: 100%;
	height: calc(100vh - 100px);
}
</style>
