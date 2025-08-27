<template>
	<div class="merchant-overview">
		<!-- 顶部操作区域 -->
		<t-space
			direction="vertical"
			size="16"
		>
			<!-- 页面标题和操作按钮 -->
			<div class="page-header">
				<div class="header-content">
					<div class="title-section">
						<h1 class="page-title">
							{{ $t('merchant.title') }}
						</h1>
						<p class="page-description">
							{{ $t('merchant.description') }}
						</p>
					</div>
					<t-space size="8">
						<t-button
							theme="primary"
							@click="() => router.push({ name: 'merchants.create' })"
						>
							<template #icon>
								<Icon icon="plus" />
							</template>
							{{ $t('merchant.create.title') }}
						</t-button>
						<t-button
							theme="default"
							variant="outline"
							@click="showImportModal = true"
						>
							<template #icon>
								<Icon icon="upload" />
							</template>
							{{ $t('merchant.import.title') }}
						</t-button>
					</t-space>
				</div>
			</div>

			<!-- 搜索筛选区域 -->
			<t-card class="search-card">
				<t-space
					direction="vertical"
					size="16"
				>
					<t-row :gutter="16">
						<t-col :span="8">
							<t-input
								v-model="searchQuery"
								:placeholder="$t('merchant.search.placeholder')"
								clearable
								@input="handleSearchInput"
							>
								<template #prefix-icon>
									<Icon icon="search" />
								</template>
							</t-input>
						</t-col>
						<t-col :span="4">
							<t-button
								v-if="hasActiveFilters"
								theme="default"
								variant="outline"
								block
								@click="clearFilters"
							>
								{{ $t('misc.clearFilters') }}
							</t-button>
						</t-col>
					</t-row>
					
					<!-- 批量操作提示 -->
					<t-alert
						v-if="selectedRowKeys.length > 0"
						theme="info"
						close
						@close="clearSelection"
					>
						<template #message>
							<t-space align="center">
								<span>{{ $t('merchant.selected', { count: selectedRowKeys.length }) }}</span>
								<t-button
									theme="danger"
									size="small"
									variant="text"
									@click="bulkDeleteConfirm"
								>
									<template #icon>
										<Icon icon="trash-alt" />
									</template>
									{{ $t('merchant.bulkDelete') }}
								</t-button>
							</t-space>
						</template>
					</t-alert>
				</t-space>
			</t-card>

			<!-- 表格区域 -->
			<t-card>
				<t-table
					:data="filteredMerchants"
					:columns="columns"
					:selected-row-keys="selectedRowKeys"
					:loading="loading"
					:pagination="pagination"
					:filter-value="filterValue"
					:sort="sort"
					row-key="id"
					stripe
					hover
					resizable
					size="medium"
					table-layout="fixed"
					cell-empty-content="-"
					@selectChange="onSelectChange"
					@filterChange="onFilterChange"
					@sortChange="onSortChange"
					@pageChange="onPageChange"
					@rowClick="onRowClick"
				>
					<!-- 商户名称列 -->
					<template #title="{ row }">
						<t-link
							:href="`#/merchants/${row.id}/edit`"
							theme="primary"
							hover="color"
							@click.prevent="editMerchant(row)"
						>
							{{ row.title }}
						</t-link>
					</template>

					<!-- 状态列 -->
					<template #status="{ row }">
						<t-tag
							:theme="getStatusTheme(row.status)"
							variant="light-outline"
						>
							{{ row.status }}
						</t-tag>
					</template>

					<!-- 操作列 -->
					<template #actions="{ row }">
						<t-space size="0">
							<t-button
								theme="primary"
								size="small"
								variant="text"
								@click="editMerchant(row)"
							>
								<template #icon>
									<Icon icon="pencil-alt" />
								</template>
								{{ $t('misc.edit') }}
							</t-button>
							<t-button
								theme="default"
								size="small"
								variant="text"
								@click="selectMerchantForTask(row)"
							>
								<template #icon>
									<Icon icon="copy" />
								</template>
								{{ $t('merchant.useInTask') }}
							</t-button>
							<t-popconfirm
								:content="$t('merchant.delete.text1', { merchant: row.title })"
								theme="danger"
								@confirm="() => deleteMerchant(row)"
							>
								<t-button
									theme="danger"
									size="small"
									variant="text"
								>
									<template #icon>
										<Icon icon="trash-alt" />
									</template>
									{{ $t('misc.delete') }}
								</t-button>
							</t-popconfirm>
						</t-space>
					</template>
				</t-table>
			</t-card>
		</t-space>

		<!-- 导入模态框 -->
		<t-dialog
			v-model:visible="showImportModal"
			:header="$t('merchant.import.title')"
			:confirm-btn="$t('misc.doit')"
			:cancel-btn="$t('misc.cancel')"
			width="500px"
			@confirm="importMerchants"
		>
			<t-space direction="vertical">
				<p>{{ $t('merchant.import.description') }}</p>
				<t-upload
					ref="uploadRef"
					v-model="uploadFiles"
					:auto-upload="false"
					:multiple="false"
					accept=".xlsx,.xls"
					theme="file-input"
					placeholder="选择 Excel 文件"
					@change="handleFileSelect"
				>
					<template #file-input>
						<t-button
							theme="default"
							variant="outline"
						>
							<template #icon>
								<Icon icon="upload" />
							</template>
							选择文件
						</t-button>
					</template>
				</t-upload>
				<div
					v-if="selectedFile"
					class="selected-file-info"
				>
					<t-tag
						theme="success"
						variant="light"
					>
						{{ selectedFile.name }}
					</t-tag>
				</div>
			</t-space>
		</t-dialog>

		<!-- 批量删除确认对话框 -->
		<t-dialog
			v-model:visible="showBulkDeleteConfirm"
			:header="$t('merchant.bulkDelete.title')"
			:confirm-btn="$t('misc.doit')"
			:cancel-btn="$t('misc.cancel')"
			theme="danger"
			@confirm="bulkDelete"
		>
			<p>{{ $t('merchant.bulkDelete.text1', { count: selectedRowKeys.length }) }}</p>
			<p>{{ $t('merchant.bulkDelete.text2') }}</p>
		</t-dialog>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import type { PrimaryTableCol, FilterValue, TableSort, PageInfo } from 'tdesign-vue-next'

import MerchantService from '@/services/merchant'
import MerchantModel from '@/models/merchant'
import { success, error } from '@/message'

const { t } = useI18n()
const router = useRouter()
const merchantService = new MerchantService()

// 数据状态
const merchants = ref<MerchantModel[]>([])
const loading = ref(false)

// 搜索和筛选状态
const searchQuery = ref('')
const searchTimeout = ref<NodeJS.Timeout | null>(null)

// 表格选择状态
const selectedRowKeys = ref<(string | number)[]>([])

// 弹窗状态
const showImportModal = ref(false)
const showBulkDeleteConfirm = ref(false)
const selectedFile = ref<File | null>(null)
const uploadFiles = ref([])
const uploadRef = ref()

// TDesign 表格控制状态
const filterValue = ref<FilterValue>({})
const sort = ref<TableSort>()

// 分页配置
const pagination = ref({
	current: 1,
	pageSize: 20,
	total: 0,
	showJumper: true,
	showSizeChanger: true,
	pageSizeOptions: [10, 20, 50, 100],
})

// 表格列配置
const columns = computed<PrimaryTableCol[]>(() => [
	{
		colKey: 'row-select',
		type: 'multiple',
		width: 64,
		fixed: 'left',
	},
	{
		colKey: 'title',
		title: t('merchant.attributes.title'),
		width: 200,
		cell: 'title',
		sorter: true,
		ellipsis: true,
	},
	{
		colKey: 'legalRepresentative',
		title: t('merchant.attributes.legalRepresentative'),
		width: 150,
		sorter: true,
		ellipsis: true,
	},
	{
		colKey: 'businessDistrict',
		title: t('merchant.attributes.businessDistrict'),
		width: 150,
		filter: {
			type: 'multiple',
			list: uniqueDistricts.value.map(d => ({ label: d, value: d })),
			showConfirmAndReset: true,
		},
		sorter: true,
	},
	{
		colKey: 'validTime',
		title: t('merchant.attributes.validTime'),
		width: 120,
		sorter: true,
	},
	{
		colKey: 'terminalType',
		title: t('merchant.attributes.terminalType'),
		width: 120,
		filter: {
			type: 'multiple',
			list: uniqueTerminalTypes.value.map(d => ({ label: d, value: d })),
			showConfirmAndReset: true,
		},
		sorter: true,
	},
	{
		colKey: 'status',
		title: t('merchant.attributes.status'),
		width: 100,
		cell: 'status',
		filter: {
			type: 'single',
			list: [
				{ label: '活跃', value: 'active' },
				{ label: '停用', value: 'inactive' },
			],
		},
	},
	{
		colKey: 'actions',
		title: t('misc.actions'),
		width: 200,
		cell: 'actions',
		fixed: 'right',
	},
])

// 计算属性
const filteredMerchants = computed(() => {
	let filtered = merchants.value

	// 搜索过滤
	if (searchQuery.value) {
		const query = searchQuery.value.toLowerCase()
		filtered = filtered.filter(merchant =>
			merchant.title.toLowerCase().includes(query) ||
			merchant.legalRepresentative.toLowerCase().includes(query) ||
			merchant.businessAddress.toLowerCase().includes(query) ||
			merchant.businessDistrict.toLowerCase().includes(query),
		)
	}

	return filtered
})

const uniqueDistricts = computed(() => {
	const districts = new Set(merchants.value.map(m => m.businessDistrict).filter(Boolean))
	return Array.from(districts).sort()
})

const uniqueTerminalTypes = computed(() => {
	const types = new Set(merchants.value.map(m => m.terminalType).filter(Boolean))
	return Array.from(types).sort()
})

const hasActiveFilters = computed(() => {
	const hasFilter = Object.keys(filterValue.value).some(key => {
		const value = filterValue.value[key]
		return value !== undefined && value !== null && 
			(Array.isArray(value) ? value.length > 0 : value !== '')
	})
	return !!searchQuery.value || hasFilter
})

// 方法
onMounted(() => {
	loadMerchants()
})

async function loadMerchants(page = 1) {
	loading.value = true
	try {
		// 添加分页参数到API请求 - 使用后端期望的参数名
		const params = {
			page: page,
			perPage: pagination.value.pageSize, // 注意：后端期望 perPage 而不是 per_page
		}
		
		const response = await merchantService.getAll(new MerchantModel(), params, page)

		// AbstractService的getAll方法返回的是IMerchant[]，需要转换为MerchantModel[]
		merchants.value = response.map(merchant => new MerchantModel(merchant))
		
		// 从service中获取分页信息（这些值在getAll方法中被设置）
		pagination.value.total = merchantService.resultCount
		pagination.value.current = page
		
		// 如果没有分页信息，使用返回的数据长度作为总数
		if (!pagination.value.total || pagination.value.total === 0) {
			pagination.value.total = merchants.value.length
		}
		
		console.log(`Loaded page ${page} with ${merchants.value.length} merchants, total: ${pagination.value.total}`)
	} catch (err: unknown) {
		if (err && typeof err === 'object' && 'message' in err) {
			const msg = (err as { message?: unknown }).message
			if (typeof msg === 'string') {
				error({ message: msg })
			} else {
				error({ message: 'Unknown error' })
			}
		} else {
			error({ message: 'Unknown error' })
		}
		// 出错时重置数据
		merchants.value = []
		pagination.value.total = 0
		pagination.value.current = page
	} finally {
		loading.value = false
	}
}

function handleSearchInput() {
	if (searchTimeout.value) {
		clearTimeout(searchTimeout.value)
	}
	searchTimeout.value = setTimeout(() => {
		// 实时搜索通过 computed 属性处理
	}, 300)
}

function clearFilters() {
	searchQuery.value = ''
	filterValue.value = {}
	sort.value = undefined
}

// TDesign 表格事件处理
function onSelectChange(selectedKeys: (string | number)[]) {
	selectedRowKeys.value = selectedKeys
}

function onFilterChange(filters: FilterValue) {
	filterValue.value = filters
}

function onSortChange(sortInfo: TableSort) {
	sort.value = sortInfo
}

function onPageChange(pageInfo: PageInfo) {
	// 更新分页配置
	pagination.value.current = pageInfo.current
	pagination.value.pageSize = pageInfo.pageSize
	
	// 重新加载数据
	loadMerchants(pageInfo.current)
}

function onRowClick() {
	// 处理行点击事件
}

function clearSelection() {
	selectedRowKeys.value = []
}

function getStatusTheme(status: string) {
	const themeMap = {
		'active': 'success',
		'inactive': 'danger',
		'pending': 'warning',
	}
	return themeMap[status as keyof typeof themeMap] || 'default'
}

function editMerchant(merchant: MerchantModel) {
	router.push({ name: 'merchants.edit', params: { id: merchant.id } })
}

function bulkDeleteConfirm() {
	showBulkDeleteConfirm.value = true
}

async function bulkDelete() {
	try {
		const deletePromises = selectedRowKeys.value.map(id => {
			const merchant = merchants.value.find(m => m.id === id)
			return merchant ? merchantService.delete(merchant) : Promise.resolve()
		})

		await Promise.all(deletePromises)
		success({ message: t('merchant.bulkDelete.success', { count: selectedRowKeys.value.length }) })
		selectedRowKeys.value = []
		showBulkDeleteConfirm.value = false
		await loadMerchants(pagination.value.current)
	} catch (err: unknown) {
		if (err && typeof err === 'object' && 'message' in err) {
			const msg = (err as { message?: unknown }).message
			if (typeof msg === 'string') {
				error({ message: msg })
			} else {
				error({ message: 'Unknown error' })
			}
		} else {
			error({ message: 'Unknown error' })
		}
	}
}

// 任务集成
function selectMerchantForTask(merchant: MerchantModel) {
	router.push({
		name: 'tasks.create',
		query: {
			merchantId: merchant.id.toString(),
			merchantTitle: merchant.title,
		},
	})
}

// 删除商户
async function deleteMerchant(merchant: MerchantModel) {
	try {
		await merchantService.delete(merchant)
		success({ message: t('merchant.delete.success') })
		await loadMerchants(pagination.value.current)
	} catch (err: unknown) {
		if (err && typeof err === 'object' && 'message' in err) {
			const msg = (err as { message?: unknown }).message
			if (typeof msg === 'string') error({ message: msg })
			else error({ message: 'Unknown error' })
		} else {
			error({ message: 'Unknown error' })
		}
	}
}

// 文件上传处理
function handleFileSelect(files: File[]) {
	if (files && files.length > 0) {
		selectedFile.value = files[0]
	}
}

async function importMerchants() {
	if (!selectedFile.value) {
		error({ message: t('merchant.import.noFileSelected') })
		return
	}

	try {
		loading.value = true
		const importedMerchants = await merchantService.importFromXlsx(selectedFile.value)
		success({ message: t('merchant.import.success', { count: importedMerchants.length }) })
		showImportModal.value = false
		selectedFile.value = null
		uploadFiles.value = []
		await loadMerchants()
	} catch (err: unknown) {
		let errorMessage = 'Import failed'
		if (err && typeof err === 'object') {
			const maybeResponse = (err as { response?: unknown }).response
			const data = maybeResponse && typeof maybeResponse === 'object' ? (maybeResponse as { data?: unknown }).data : undefined
			if (data && typeof data === 'object') {
				const msg = (data as { error?: unknown; message?: unknown }).error ?? (data as { error?: unknown; message?: unknown }).message
				if (typeof msg === 'string') errorMessage = msg
			} else if ('message' in err) {
				const msg = (err as { message?: unknown }).message
				if (typeof msg === 'string') errorMessage = msg
			}
		}
		error({ message: errorMessage })
	} finally {
		loading.value = false
	}
}
</script>

<style scoped lang="scss">
.merchant-overview {
	padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
	background-color: var(--td-bg-color-container);
	min-height: 100vh;
}

// 页面头部
.page-header {
	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		margin-bottom: var(--td-comp-margin-l);

		.title-section {
			.page-title {
				font-size: var(--td-font-size-title-large);
				font-weight: var(--td-font-weight-bold);
				color: var(--td-text-color-primary);
				margin: 0 0 var(--td-comp-margin-xs) 0;
				line-height: var(--td-line-height-title-large);
			}

			.page-description {
				color: var(--td-text-color-secondary);
				margin: 0;
				font-size: var(--td-font-size-body-medium);
				line-height: var(--td-line-height-body-medium);
			}
		}
	}
}

// 搜索卡片
.search-card {
	margin-bottom: var(--td-comp-margin-l);

	:deep(.t-card__body) {
		padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
	}
}

// 文件选择提示
.selected-file-info {
	margin-top: var(--td-comp-margin-s);
}

// TDesign 表格自定义样式
:deep(.t-table) {
	.t-table__header {
		background-color: var(--td-bg-color-container-hover);

		.t-table__th {
			color: var(--td-text-color-primary);
			font-weight: var(--td-font-weight-bold);
			border-bottom: 1px solid var(--td-component-border);
		}
	}

	.t-table__body {
		.t-table__row {
			&:hover {
				background-color: var(--td-bg-color-container-hover);
			}

			.t-table__td {
				border-bottom: 1px solid var(--td-component-stroke);
				color: var(--td-text-color-primary);
			}
		}

		// 选中行样式
		.t-table__row--selected {
			background-color: var(--td-bg-color-container-select);
		}
	}

	// 固定列阴影
	.t-table__content--fixed-left::after {
		box-shadow: inset 10px 0 8px -8px var(--td-shadow-inset-right);
	}

	.t-table__content--fixed-right::before {
		box-shadow: inset -10px 0 8px -8px var(--td-shadow-inset-left);
	}

	// 修复左侧多选框位置：更具体地定位到表格内的选择列和复选框元素，确保覆盖 TDesign 默认样式
	:deep(.t-table) td[key=row-select],
	:deep(.t-table) th[key=row-select],
	:deep(.t-table) .t-table__cell--selectable[key=row-select] {
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
		padding-top: 0 !important;
		padding-bottom: 0 !important;
		padding-left: var(--td-comp-paddingLR-l) !important;
		padding-right: var(--td-comp-paddingLR-l) !important;
	}

	// 如果复选框被包裹在 .t-checkbox 中，确保其内部 input/视觉容器居中
	:deep(.t-table) td[key=row-select] .t-checkbox,
	:deep(.t-table) th[key=row-select] .t-checkbox,
	:deep(.t-table) td[key=row-select] .t-checkbox__input,
	:deep(.t-table) th[key=row-select] .t-checkbox__input {
		display: inline-flex !important;
		align-items: center !important;
		justify-content: center !important;
		margin: 0 !important;
	}
}

// 链接样式
:deep(.t-link) {
	font-weight: var(--td-font-weight-medium);
	
	&:hover {
		text-decoration: none;
	}
}

// 操作按钮间距
:deep(.t-space--size-0 .t-space-item:not(:last-child)) {
	margin-right: var(--td-comp-margin-xs);
}

// 标签样式
:deep(.t-tag) {
	border-radius: var(--td-radius-medium);
}

// 响应式设计
@media (max-width: 768px) {
	.merchant-overview {
		padding: var(--td-comp-paddingTB-m) var(--td-comp-paddingLR-m);
	}

	.page-header .header-content {
		flex-direction: column;
		align-items: stretch;
		gap: var(--td-comp-margin-l);

		.title-section {
			text-align: center;
		}
	}

	:deep(.t-row) {
		flex-direction: column;

		.t-col {
			width: 100% !important;
			margin-bottom: var(--td-comp-margin-m);

			&:last-child {
				margin-bottom: 0;
			}
		}
	}

	// 移动端表格横向滚动
	:deep(.t-table) {
		.t-table__content {
			overflow-x: auto;
		}
	}
}

// 暗色主题适配
@media (prefers-color-scheme: dark) {
	.merchant-overview {
		background-color: var(--td-bg-color-page);
	}
}

// 加载状态样式
:deep(.t-loading) {
	background-color: var(--td-bg-color-container);
}

// 分页器样式
:deep(.t-pagination) {
	margin-top: var(--td-comp-margin-l);
	justify-content: center;
}

// 对话框样式
:deep(.t-dialog) {
	.t-dialog__header {
		border-bottom: 1px solid var(--td-component-border);
	}

	.t-dialog__body {
		padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
	}
}

// 文件上传样式
:deep(.t-upload) {
	.t-upload__dragger {
		border: 2px dashed var(--td-component-border);
		border-radius: var(--td-radius-medium);
		transition: all var(--td-transition);

		&:hover {
			border-color: var(--td-brand-color);
		}
	}
}

// 警告信息样式
:deep(.t-alert) {
	border-radius: var(--td-radius-medium);
	
	&.t-alert--info {
		background-color: var(--td-brand-color-light);
		border-color: var(--td-brand-color);
	}
}

// 卡片样式
:deep(.t-card) {
	border-radius: var(--td-radius-large);
	box-shadow: var(--td-shadow-1);
	border: 1px solid var(--td-component-border);
}
</style>
