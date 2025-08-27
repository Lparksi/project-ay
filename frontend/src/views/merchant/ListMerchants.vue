<template>
	<div class="content merchant-overview">
		<header class="merchant-header">
			<div class="fancylists-title">
				<h1>{{ $t('merchant.title') }}</h1>
				<p>{{ $t('merchant.description') }}</p>
			</div>
			<div class="filter-container">
				<div class="items">
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
						@click="showImportModal = true"
					>
						<template #icon>
							<Icon icon="upload" />
						</template>
						{{ $t('merchant.import.title') }}
					</t-button>
				</div>
			</div>
		</header>

		<!-- Search and Filter Controls -->
		<Card class="filter-section">
			<div class="search-and-filters">
				<div class="search-box">
					<t-input
						v-model="searchQuery"
						:placeholder="$t('merchant.search.placeholder')"
						clearable
						@change="handleSearchInput"
					>
						<template #prefix-icon>
							<Icon icon="search" />
						</template>
					</t-input>
				</div>
				<div class="filter-controls">
					<!-- 使用表头内置筛选，删除页面级下拉 -->
					<t-button
						v-if="hasActiveFilters"
						theme="default"
						size="small"
						@click="clearFilters"
					>
						{{ $t('misc.clearFilters') }}
					</t-button>
				</div>
			</div>
		</Card>

		<!-- Bulk Actions -->
		<Card
			v-if="selectedRowKeys.length > 0"
			class="bulk-actions"
		>
			<div class="selected-info">
				<span>{{ $t('merchant.selected', { count: selectedRowKeys.length }) }}</span>
				<div class="bulk-buttons">
					<t-button
						theme="danger"
						size="small"
						@click="bulkDeleteConfirm"
					>
						<template #icon>
							<Icon icon="trash-alt" />
						</template>
						{{ $t('merchant.bulkDelete') }}
					</t-button>
					<t-button
						theme="default"
						size="small"
						@click="clearSelection"
					>
						{{ $t('misc.clearSelection') }}
					</t-button>
				</div>
			</div>
		</Card>

		<Card
			:padding="false"
			:has-content="false"
			class="has-table"
		>
			<t-table
				:data="filteredMerchants"
				:columns="columns"
				:selected-row-keys="selectedRowKeys"
				:loading="loading"
				row-key="id"
				select-on-row-click
				stripe
				hover
				size="medium"
				table-layout="auto"
				@selectChange="onSelectChange"
				@change="onTableChange"
				@rowClick="onRowClick"
			>
				<template #title="{ row }">
					<RouterLink
						:to="{ name: 'merchants.edit', params: { id: row.id } }"
						class="title-link"
					>
						{{ row.title }}
					</RouterLink>
				</template>
				<template #actions="{ row }">
					<t-space size="small">
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
						<t-button
							theme="danger"
							size="small"
							variant="text"
							@click="showDeleteModal(row)"
						>
							<template #icon>
								<Icon icon="trash-alt" />
							</template>
							{{ $t('misc.delete') }}
						</t-button>
					</t-space>
				</template>
			</t-table>
		</Card>

		<!-- TDesign Pagination -->
		<div
			v-if="totalPages > 1"
			class="pagination-wrapper"
		>
			<t-pagination
				v-model="currentPage"
				:total="totalItems"
				:page-size="pageSize"
				:show-jumper="true"
				:show-page-size="false"
				@change="onPageChange"
			/>
		</div>

		<!-- Import Modal -->
		<Modal
			:enabled="showImportModal"
			@close="closeImportModal"
			@submit="importMerchants"
		>
			<template #header>
				<card-title>{{ $t('merchant.import.title') }}</card-title>
			</template>
			<template #text>
				<p>{{ $t('merchant.import.description') }}</p>
				<div class="file-input-container">
					<input
						ref="importFileInput"
						type="file"
						accept=".xlsx,.xls"
						class="file-input"
						@change="handleFileSelect"
					>
					<p
						v-if="selectedFile"
						class="selected-file"
					>
						{{ $t('merchant.import.selectedFile') }}: {{ selectedFile.name }}
					</p>
				</div>
			</template>
		</Modal>

		<!-- Delete Modal -->
		<t-dialog
			v-model:visible="showDeleteConfirm"
			:header="$t('merchant.delete.title')"
			:confirm-btn="$t('misc.doit')"
			:cancel-btn="$t('misc.cancel')"
			theme="danger"
			@confirm="deleteMerchant"
		>
			<div v-if="merchantToDelete">
				<p>{{ $t('merchant.delete.text1', { merchant: merchantToDelete.title }) }}</p>
				<p>{{ $t('merchant.delete.text2') }}</p>
			</div>
		</t-dialog>

		<!-- Bulk Delete Modal -->
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
import type { PrimaryTableCol } from 'tdesign-vue-next'

import MerchantService from '@/services/merchant'
import MerchantModel from '@/models/merchant'
import { success, error } from '@/message'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'

const { t } = useI18n()
const router = useRouter()
const merchantService = new MerchantService()

// Data
const merchants = ref<MerchantModel[]>([])
const totalPages = ref(0)
const totalItems = ref(0)
const pageSize = ref(20)
const currentPage = ref(1)
const loading = ref(false)

// Search and filter
const searchQuery = ref('')
// page-level selects removed in favor of table built-in filters
// const filterDistrict = ref('')
// const filterTerminalType = ref('')
const searchTimeout = ref<NodeJS.Timeout | null>(null)
// TDesign 表格的筛选值（由表格触发器维护）
type TableFilters = Record<string, string[] | string | undefined>
const tableFilterValue = ref<TableFilters>({})

// Selection
const selectedRowKeys = ref<(string | number)[]>([])

// Modals
const showImportModal = ref(false)
const showDeleteConfirm = ref(false)
const showBulkDeleteConfirm = ref(false)
const merchantToDelete = ref<MerchantModel | null>(null)
const importFileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)

// Table columns configuration
const columns = computed<PrimaryTableCol[]>(() => [
	{
		colKey: 'title',
		title: t('merchant.attributes.title'),
		width: 200,
		cell: 'title',
		sorter: true,
	},
	{
		colKey: 'legalRepresentative',
		title: t('merchant.attributes.legalRepresentative'),
		width: 150,
		sorter: true,
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
		colKey: 'actions',
		title: t('misc.actions'),
		width: 200,
		cell: 'actions',
		fixed: 'right',
	},
])

// Computed properties
const filteredMerchants = computed(() => {
	let filtered = merchants.value

	// Apply search filter
	if (searchQuery.value) {
		const query = searchQuery.value.toLowerCase()
		filtered = filtered.filter(merchant =>
			merchant.title.toLowerCase().includes(query) ||
			merchant.legalRepresentative.toLowerCase().includes(query) ||
			merchant.businessAddress.toLowerCase().includes(query) ||
			merchant.businessDistrict.toLowerCase().includes(query),
		)
	}

	// Apply table column filters (from TDesign table filter value)
	const fv = tableFilterValue.value || {}
	if (fv.businessDistrict && Array.isArray(fv.businessDistrict) && fv.businessDistrict.length) {
		filtered = filtered.filter(merchant => fv.businessDistrict.includes(merchant.businessDistrict))
	}
	if (fv.terminalType && Array.isArray(fv.terminalType) && fv.terminalType.length) {
		filtered = filtered.filter(merchant => fv.terminalType.includes(merchant.terminalType))
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
	const tv = tableFilterValue.value || {}
	const hasTableFilter = Object.keys(tv).length > 0 && Object.values(tv).some(v => v != null && (Array.isArray(v) ? v.length > 0 : String(v) !== ''))
	return !!searchQuery.value || hasTableFilter
})



// Methods
onMounted(() => {
	loadMerchants()
})

async function loadMerchants(page = 1) {
	loading.value = true
	try {
		const response = await merchantService.getAll(new MerchantModel(), {}, page)

		// merchantService.getAll may return an array or a paginated object
		if (Array.isArray(response)) {
			merchants.value = response
			totalPages.value = 1
			totalItems.value = merchants.value.length
		} else if (response && typeof response === 'object') {
			// try to read data/total/totalPages in a type-safe way
			const data = (response as { data?: MerchantModel[]; total?: number; totalPages?: number }).data
			merchants.value = Array.isArray(data) ? data : []
			totalPages.value = (response as { totalPages?: number }).totalPages || 1
			totalItems.value = (response as { total?: number }).total || merchants.value.length
		} else {
			merchants.value = []
			totalPages.value = 1
			totalItems.value = 0
		}
		currentPage.value = page
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
	} finally {
		loading.value = false
	}
}

// TDesign pagination change handler
function onPageChange(pageInfo: { current: number; previous: number }) {
	loadMerchants(pageInfo.current)
}

function handleSearchInput() {
	if (searchTimeout.value) {
		clearTimeout(searchTimeout.value)
	}
	searchTimeout.value = setTimeout(() => {
		// If we implement server-side search in the future, call loadMerchants here
		// For now, the computed filteredMerchants handles client-side filtering
	}, 300)
}


function clearFilters() {
	searchQuery.value = ''
	// reset table filters
	tableFilterValue.value = {}
}

// Table change handler from TDesign table (captures filter changes)
function onTableChange(params: { filter?: TableFilters } = {}) {
	tableFilterValue.value = params.filter || {}
}

// Selection methods
function onSelectChange(selectedKeys: (string | number)[]) {
	selectedRowKeys.value = selectedKeys
}

function onRowClick() {
	// Handle row click if needed
}

function clearSelection() {
	selectedRowKeys.value = []
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
		await loadMerchants(currentPage.value)
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

// Task integration
function selectMerchantForTask(merchant: MerchantModel) {
	// For now, we'll navigate to task creation with merchant data as query params
	router.push({
		name: 'tasks.create',
		query: {
			merchantId: merchant.id.toString(),
			merchantTitle: merchant.title,
		},
	})
}

// Modal methods
function showDeleteModal(merchant: MerchantModel) {
	merchantToDelete.value = merchant
	showDeleteConfirm.value = true
}

async function deleteMerchant() {
	if (!merchantToDelete.value) return

	try {
		await merchantService.delete(merchantToDelete.value)
		success({ message: t('merchant.delete.success') })
		await loadMerchants(currentPage.value)
		showDeleteConfirm.value = false
		merchantToDelete.value = null
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

function handleFileSelect(event: Event) {
	const target = event.target as HTMLInputElement
	selectedFile.value = target.files?.[0] || null
	console.log('File selected:', selectedFile.value?.name)
}

function closeImportModal() {
	showImportModal.value = false
	selectedFile.value = null
	// Reset file input
	if (importFileInput.value) {
		importFileInput.value.value = ''
	}
}

async function importMerchants() {
	console.log('=== Import merchants called ===')
	console.log('Selected file:', selectedFile.value)
	console.log('File name:', selectedFile.value?.name)
	console.log('File size:', selectedFile.value?.size)
	console.log('File type:', selectedFile.value?.type)

	if (!selectedFile.value) {
		console.error('No file selected')
		error({ message: t('merchant.import.noFileSelected') })
		return
	}

	// Additional validation
	if (selectedFile.value.size === 0) {
		console.error('File is empty')
		error({ message: 'Selected file is empty' })
		return
	}

	const allowedTypes = [
		'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', // .xlsx
		'application/vnd.ms-excel', // .xls
	]

	if (!allowedTypes.includes(selectedFile.value.type) &&
		!selectedFile.value.name.toLowerCase().endsWith('.xlsx') &&
		!selectedFile.value.name.toLowerCase().endsWith('.xls')) {
		console.error('Invalid file type:', selectedFile.value.type)
		error({ message: 'Please select a valid Excel file (.xlsx or .xls)' })
		return
	}

	try {
		loading.value = true
		console.log('Starting import...')
		const importedMerchants = await merchantService.importFromXlsx(selectedFile.value)
		console.log('Import completed, merchants:', importedMerchants)
		success({ message: t('merchant.import.success', { count: importedMerchants.length }) })
		closeImportModal()
		await loadMerchants()
	} catch (err: unknown) {
		console.error('Import error:', err)

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
	padding: 1rem;
}

.merchant-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-end;
	margin-bottom: 1rem;

	.fancylists-title {
		h1 {
			font-size: 2rem;
			margin: 0;
		}

		p {
			color: var(--grey-500);
			margin: 0.5rem 0 0;
		}
	}

	.filter-container .items {
		display: flex;
		gap: 0.5rem;
	}
}

.filter-section {
	margin-bottom: 1rem;

	.search-and-filters {
		display: flex;
		gap: 1rem;
		align-items: center;
		flex-wrap: wrap;

		.search-box {
			flex: 1;
			min-width: 300px;
		}

		.filter-controls {
			display: flex;
			gap: 0.5rem;
			align-items: center;
			flex-wrap: wrap;

			:deep(.t-select) {
				min-width: 150px;
			}
		}
	}
}

.bulk-actions {
	margin-bottom: 1rem;
	background-color: var(--primary-light);
	border: 1px solid var(--primary);

	.selected-info {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-weight: 500;

		.bulk-buttons {
			display: flex;
			gap: 0.5rem;
		}
	}
}

.title-link {
	font-weight: 500;
	color: var(--primary);
	text-decoration: none;

	&:hover {
		text-decoration: underline;
	}
}

.file-input-container {
	margin: 1rem 0;

	.file-input {
		width: 100%;
		padding: 0.5rem;
		border: 2px dashed var(--grey-300);
		border-radius: 4px;
		background: var(--white);
		cursor: pointer;

		&:hover {
			border-color: var(--primary);
		}
	}

	.selected-file {
		margin-top: 0.5rem;
		padding: 0.5rem;
		background: var(--success-light);
		border-radius: 4px;
		color: var(--success-dark);
		font-size: 0.9rem;
	}
}

// TDesign table customization
:deep(.t-table) {
	.t-table__header {
		background-color: var(--grey-50);
	}

	.t-table__row:hover {
		background-color: var(--grey-25);
	}
}

// TDesign pagination customization
.pagination-wrapper {
	display: flex;
	justify-content: center;
	margin-top: 1rem;
	padding: 1rem 0;

	/* 使用 TDesign 默认样式，无需额外定制 */
}

@media (max-width: 768px) {
	.search-and-filters {
		flex-direction: column;
		align-items: stretch;

		.search-box {
			min-width: 100%;
		}

		.filter-controls {
			justify-content: center;
		}
	}

	.bulk-actions .selected-info {
		flex-direction: column;
		gap: 0.5rem;
		text-align: center;
	}
}
</style>
