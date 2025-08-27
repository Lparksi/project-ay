<template>
	<div class="content merchant-overview">
		<header class="merchant-header">
			<div class="fancylists-title">
				<h1>{{ $t('merchant.title') }}</h1>
				<p>{{ $t('merchant.description') }}</p>
			</div>
			<div class="filter-container">
				<div class="items">
					<XButton
						icon="plus"
						:to="{name: 'merchant.create'}"
						variant="primary"
					>
						{{ $t('merchant.create.title') }}
					</XButton>
					<XButton
						icon="upload"
						@click="showImportModal = true"
						variant="secondary"
					>
						{{ $t('merchant.import.title') }}
					</XButton>
				</div>
			</div>
		</header>

		<!-- Search and Filter Controls -->
		<Card class="filter-section">
			<div class="search-and-filters">
				<div class="search-box">
					<input
						v-model="searchQuery"
						type="text"
						:placeholder="$t('merchant.search.placeholder')"
						class="input"
						@input="handleSearchInput"
					>
				</div>
				<div class="filter-controls">
					<select
						v-model="filterDistrict"
						class="select"
						@change="applyFilters"
					>
						<option value="">{{ $t('merchant.filter.allDistricts') }}</option>
						<option 
							v-for="district in uniqueDistricts" 
							:key="district" 
							:value="district"
						>
							{{ district }}
						</option>
					</select>
					<select
						v-model="filterTerminalType"
						class="select"
						@change="applyFilters"
					>
						<option value="">{{ $t('merchant.filter.allTerminalTypes') }}</option>
						<option 
							v-for="terminalType in uniqueTerminalTypes" 
							:key="terminalType" 
							:value="terminalType"
						>
							{{ terminalType }}
						</option>
					</select>
					<XButton
						v-if="hasActiveFilters"
						@click="clearFilters"
						variant="secondary"
						size="small"
					>
						{{ $t('misc.clearFilters') }}
					</XButton>
				</div>
			</div>
		</Card>

		<!-- Bulk Actions -->
		<Card v-if="selectedMerchants.length > 0" class="bulk-actions">
			<div class="selected-info">
				<span>{{ $t('merchant.selected', {count: selectedMerchants.length}) }}</span>
				<div class="bulk-buttons">
					<XButton
						@click="bulkDeleteConfirm"
						variant="danger"
						size="small"
						icon="trash-alt"
					>
						{{ $t('merchant.bulkDelete') }}
					</XButton>
					<XButton
						@click="clearSelection"
						variant="secondary"
						size="small"
					>
						{{ $t('misc.clearSelection') }}
					</XButton>
				</div>
			</div>
		</Card>

		<Card :padding="false" :has-content="false" class="has-table">
			<div class="table-wrapper">
				<table class="table is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th class="select-column">
								<input
									type="checkbox"
									:checked="isAllSelected"
									@change="toggleSelectAll"
								>
							</th>
							<th>{{ $t('merchant.attributes.title') }}</th>
							<th>{{ $t('merchant.attributes.legalRepresentative') }}</th>
							<th>{{ $t('merchant.attributes.businessDistrict') }}</th>
							<th>{{ $t('merchant.attributes.validTime') }}</th>
							<th>{{ $t('merchant.attributes.terminalType') }}</th>
							<th>{{ $t('misc.actions') }}</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="merchant in filteredMerchants" :key="merchant.id">
							<td class="select-column">
								<input
									type="checkbox"
									:checked="isSelected(merchant.id)"
									@change="toggleMerchantSelection(merchant.id)"
								>
							</td>
							<td>
								<router-link
									:to="{name: 'merchant.edit', params: {id: merchant.id}}"
									class="title-link"
								>
									{{ merchant.title }}
								</router-link>
							</td>
							<td>{{ merchant.legalRepresentative }}</td>
							<td>{{ merchant.businessDistrict }}</td>
							<td>{{ merchant.validTime }}</td>
							<td>{{ merchant.terminalType }}</td>
							<td class="actions">
								<div class="buttons">
									<XButton
										icon="pencil-alt"
										:to="{name: 'merchant.edit', params: {id: merchant.id}}"
										size="small"
									>
										{{ $t('misc.edit') }}
									</XButton>
									<XButton
										icon="copy"
										@click="selectMerchantForTask(merchant)"
										size="small"
										variant="secondary"
									>
										{{ $t('merchant.useInTask') }}
									</XButton>
									<XButton
										icon="trash-alt"
										@click="showDeleteModal(merchant)"
										variant="danger"
										size="small"
									>
										{{ $t('misc.delete') }}
									</XButton>
								</div>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</Card>

		<Pagination
			:total-pages="totalPages"
			:current-page="currentPage"
			@update:current-page="loadMerchants"
		/>

		<!-- Import Modal -->
		<Modal
			:enabled="showImportModal"
			@close="showImportModal = false"
			@submit="importMerchants"
		>
			<template #header>
				<card-title>{{ $t('merchant.import.title') }}</card-title>
			</template>
			<template #text>
				<p>{{ $t('merchant.import.description') }}</p>
				<input
					ref="importFileInput"
					type="file"
					accept=".xlsx,.xls"
					@change="handleFileSelect"
				/>
			</template>
		</Modal>

		<!-- Delete Modal -->
		<Modal
			v-if="merchantToDelete"
			:enabled="showDeleteConfirm"
			@close="showDeleteConfirm = false"
			@submit="deleteMerchant"
			variant="danger"
		>
			<template #header>
				<card-title>{{ $t('merchant.delete.title') }}</card-title>
			</template>
			<template #text>
				<p>{{ $t('merchant.delete.text1', {merchant: merchantToDelete.title}) }}</p>
				<p>{{ $t('merchant.delete.text2') }}</p>
			</template>
		</Modal>

		<!-- Bulk Delete Modal -->
		<Modal
			:enabled="showBulkDeleteConfirm"
			@close="showBulkDeleteConfirm = false"
			@submit="bulkDelete"
			variant="danger"
		>
			<template #header>
				<card-title>{{ $t('merchant.bulkDelete.title') }}</card-title>
			</template>
			<template #text>
				<p>{{ $t('merchant.bulkDelete.text1', {count: selectedMerchants.length}) }}</p>
				<p>{{ $t('merchant.bulkDelete.text2') }}</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import MerchantService from '@/services/merchant'
import MerchantModel from '@/models/merchant'
import {success, error} from '@/message'

import Card from '@/components/misc/Card.vue'
import XButton from '@/components/input/Button.vue'
import Modal from '@/components/misc/Modal.vue'
import Pagination from '@/components/misc/Pagination.vue'

const {t} = useI18n()
const router = useRouter()
const merchantService = new MerchantService()

// Data
const merchants = ref<MerchantModel[]>([])
const totalPages = ref(0)
const currentPage = ref(1)
const loading = ref(false)

// Search and filter
const searchQuery = ref('')
const filterDistrict = ref('')
const filterTerminalType = ref('')
const searchTimeout = ref<NodeJS.Timeout | null>(null)

// Selection
const selectedMerchants = ref<number[]>([])

// Modals
const showImportModal = ref(false)
const showDeleteConfirm = ref(false)
const showBulkDeleteConfirm = ref(false)
const merchantToDelete = ref<MerchantModel | null>(null)
const importFileInput = ref<HTMLInputElement>()
const selectedFile = ref<File | null>(null)

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
			merchant.businessDistrict.toLowerCase().includes(query)
		)
	}

	// Apply district filter
	if (filterDistrict.value) {
		filtered = filtered.filter(merchant => merchant.businessDistrict === filterDistrict.value)
	}

	// Apply terminal type filter
	if (filterTerminalType.value) {
		filtered = filtered.filter(merchant => merchant.terminalType === filterTerminalType.value)
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
	return searchQuery.value || filterDistrict.value || filterTerminalType.value
})

const isAllSelected = computed(() => {
	return filteredMerchants.value.length > 0 && 
		filteredMerchants.value.every(merchant => selectedMerchants.value.includes(merchant.id))
})

// Methods
onMounted(() => {
	loadMerchants()
})

async function loadMerchants(page = 1) {
	loading.value = true
	try {
		const response = await merchantService.getAll({}, {}, page)
		merchants.value = response.data
		totalPages.value = response.totalPages
		currentPage.value = page
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
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

function applyFilters() {
	// Filters are applied automatically via computed property
}

function clearFilters() {
	searchQuery.value = ''
	filterDistrict.value = ''
	filterTerminalType.value = ''
}

// Selection methods
function isSelected(merchantId: number): boolean {
	return selectedMerchants.value.includes(merchantId)
}

function toggleMerchantSelection(merchantId: number) {
	const index = selectedMerchants.value.indexOf(merchantId)
	if (index > -1) {
		selectedMerchants.value.splice(index, 1)
	} else {
		selectedMerchants.value.push(merchantId)
	}
}

function toggleSelectAll() {
	if (isAllSelected.value) {
		// Deselect all visible merchants
		filteredMerchants.value.forEach(merchant => {
			const index = selectedMerchants.value.indexOf(merchant.id)
			if (index > -1) {
				selectedMerchants.value.splice(index, 1)
			}
		})
	} else {
		// Select all visible merchants
		filteredMerchants.value.forEach(merchant => {
			if (!selectedMerchants.value.includes(merchant.id)) {
				selectedMerchants.value.push(merchant.id)
			}
		})
	}
}

function clearSelection() {
	selectedMerchants.value = []
}

function bulkDeleteConfirm() {
	showBulkDeleteConfirm.value = true
}

async function bulkDelete() {
	try {
		const deletePromises = selectedMerchants.value.map(id => {
			const merchant = merchants.value.find(m => m.id === id)
			return merchant ? merchantService.delete(merchant) : Promise.resolve()
		})
		
		await Promise.all(deletePromises)
		success({message: t('merchant.bulkDelete.success', {count: selectedMerchants.value.length})})
		selectedMerchants.value = []
		showBulkDeleteConfirm.value = false
		await loadMerchants(currentPage.value)
	} catch (e) {
		error(e)
	}
}

// Task integration
function selectMerchantForTask(merchant: MerchantModel) {
	// Store merchant data for use in task creation
	// This could be implemented via a store or event bus
	const merchantData = {
		id: merchant.id,
		title: merchant.title,
		businessAddress: merchant.businessAddress,
		businessDistrict: merchant.businessDistrict
	}
	
	// For now, we'll navigate to task creation with merchant data as query params
	router.push({
		name: 'tasks.create',
		query: {
			merchantId: merchant.id.toString(),
			merchantTitle: merchant.title
		}
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
		success({message: t('merchant.delete.success')})
		await loadMerchants(currentPage.value)
		showDeleteConfirm.value = false
		merchantToDelete.value = null
	} catch (e) {
		error(e)
	}
}

function handleFileSelect(event: Event) {
	const target = event.target as HTMLInputElement
	selectedFile.value = target.files?.[0] || null
}

async function importMerchants() {
	if (!selectedFile.value) {
		error({message: t('merchant.import.noFileSelected')})
		return
	}

	try {
		const importedMerchants = await merchantService.importFromXlsx(selectedFile.value)
		success({message: t('merchant.import.success', {count: importedMerchants.length})})
		showImportModal.value = false
		selectedFile.value = null
		await loadMerchants()
	} catch (e) {
		error(e)
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

			.input {
				width: 100%;
			}
		}

		.filter-controls {
			display: flex;
			gap: 0.5rem;
			align-items: center;
			flex-wrap: wrap;

			.select {
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

.table-wrapper {
	overflow-x: auto;
}

.table {
	.select-column {
		width: 40px;
		text-align: center;
	}

	th.select-column,
	td.select-column {
		padding: 0.5rem;
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

.actions .buttons {
	display: flex;
	gap: 0.25rem;
	flex-wrap: wrap;
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

	.actions .buttons {
		justify-content: center;
	}
}
</style>