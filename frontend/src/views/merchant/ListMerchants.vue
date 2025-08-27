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

		<Card :padding="false" :has-content="false" class="has-table">
			<div class="table-wrapper">
				<table class="table is-striped is-hoverable is-fullwidth">
					<thead>
						<tr>
							<th>{{ $t('merchant.attributes.title') }}</th>
							<th>{{ $t('merchant.attributes.legalRepresentative') }}</th>
							<th>{{ $t('merchant.attributes.businessDistrict') }}</th>
							<th>{{ $t('merchant.attributes.validTime') }}</th>
							<th>{{ $t('merchant.attributes.terminalType') }}</th>
							<th>{{ $t('misc.actions') }}</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="merchant in merchants" :key="merchant.id">
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
	</div>
</template>

<script setup lang="ts">
import {ref, computed, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import MerchantService from '@/services/merchant'
import MerchantModel from '@/models/merchant'
import {success, error} from '@/message'

import Card from '@/components/misc/Card.vue'
import XButton from '@/components/input/Button.vue'
import Modal from '@/components/misc/Modal.vue'
import Pagination from '@/components/misc/Pagination.vue'

const {t} = useI18n()
const merchantService = new MerchantService()

const merchants = ref<MerchantModel[]>([])
const totalPages = ref(0)
const currentPage = ref(1)
const loading = ref(false)

const showImportModal = ref(false)
const showDeleteConfirm = ref(false)
const merchantToDelete = ref<MerchantModel | null>(null)
const importFileInput = ref<HTMLInputElement>()
const selectedFile = ref<File | null>(null)

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

.table-wrapper {
	overflow-x: auto;
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
}
</style>