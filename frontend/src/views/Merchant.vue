<template>
	<div class="merchant-page">
		<t-card
			:title="$t('navigation.merchant')"
			class="merchant-card"
		>
			<template #actions>
				<t-button
					variant="base"
					class="action-button"
					@click="showCreateModal = true"
				>
					<template #icon>
						<t-icon name="add" />
					</template>
					{{ $t('merchant.actions.create') }}
				</t-button>
				<t-button
					variant="danger"
					class="action-button"
					:disabled="selectedRowKeys.length === 0"
					@click="confirmBulkDelete"
				>
					<template #icon>
						<t-icon name="delete" />
					</template>
					{{ $t('merchant.actions.deleteSelected') }}
				</t-button>
				<t-button
					variant="outline"
					class="action-button"
					@click="showMappingEditor = true"
				>
					{{ $t('merchant.actions.editMappings') }}
				</t-button>
			</template>

			<t-loading :loading="loading">
				<t-table
					row-key="id"
					:data="merchants"
					:columns="columns"
					:selected-row-keys="selectedRowKeys"
					:on-select-change="handleSelectChange"
					:pagination="pagination"
					:on-page-change="handlePageChange"
					:header-affixed-top="true"
					table-layout="fixed"
					bordered
					hover
					striped
				>
					<template #businessDistrict="{ row }">
						{{ getBusinessDistrictLabel(row.businessDistrict) }}
					</template>
					
					<template #validTime="{ row }">
						<t-tag theme="primary">
							{{ row.validTime }}
						</t-tag>
					</template>
					
					<template #trafficConditions="{ row }">
						<t-tag theme="success">
							{{ row.trafficConditions }}
						</t-tag>
					</template>
					
					<template #fixedEvents="{ row }">
						<t-tag theme="warning">
							{{ row.fixedEvents }}
						</t-tag>
					</template>
					
					<template #terminalType="{ row }">
						<t-tag theme="default">
							{{ row.terminalType }}
						</t-tag>
					</template>
					
					<template #specialTimePeriods="{ row }">
						<t-tag theme="danger">
							{{ row.specialTimePeriods }}
						</t-tag>
					</template>
					
					<template #action="{ row }">
						<t-space>
							<t-button
								size="small"
								variant="outline"
								@click="editMerchant(row)"
							>
								{{ $t('merchant.actions.edit') }}
							</t-button>
							<t-button
								size="small"
								variant="outline"
								theme="danger"
								@click="deleteMerchant(row)"
							>
								{{ $t('merchant.actions.delete') }}
							</t-button>
						</t-space>
					</template>
				</t-table>
			</t-loading>
		</t-card>

		<!-- Bulk delete confirmation -->
		<t-dialog
			v-model:visible="showBulkDeleteConfirm"
			:header="t('merchant.actions.deleteMultiple')"
			@confirm="bulkDeleteSelected"
		>
			<p>{{ t('merchant.messages.confirmDeleteMultiple', {count: selectedRowKeys.length}) }}</p>
		</t-dialog>

		<!-- Label mapping editor -->
		<t-dialog
			v-model:visible="showMappingEditor"
			:header="t('merchant.actions.editMappings')"
			@confirm="saveMappings"
		>
			<div>
				<p>{{ t('merchant.mapping.help') }}</p>
				<div
					v-for="(fm, idx) in fieldMappings"
					:key="idx"
					style="margin-bottom:12px;"
				>
					<div style="font-weight:600">
						{{ fm.field }}
					</div>
					<div style="display:flex;gap:8px;margin-top:6px;flex-wrap:wrap;">
						<div
							v-for="(m, j) in fm.mappings"
							:key="j"
							style="display:flex;gap:6px;align-items:center;"
						>
							<t-input
								v-model="m.placeholder"
								placeholder="{{ t('merchant.mapping.placeholder') }}"
								style="width:160px;"
							/>
							<t-input
								v-model="m.labelId"
								placeholder="labelId"
								style="width:100px;"
							/>
							<t-button
								size="small"
								variant="text"
								@click="removeMapping(idx, j)"
							>
								-
							</t-button>
						</div>
						<t-button
							size="small"
							@click="addMapping(idx)"
						>
							+ {{ t('merchant.mapping.add') }}
						</t-button>
					</div>
				</div>
			</div>
		</t-dialog>

		<!-- 创建/编辑商户模态框 -->
		<t-dialog 
			v-model:visible="showCreateModal" 
			:header="editingMerchant ? $t('merchant.actions.edit') : $t('merchant.actions.create')"
			width="500px"
			@confirm="saveMerchant"
			@cancel="resetForm"
		>
			<t-form
				ref="formRef"
				:data="form"
				:rules="rules"
				label-align="top"
				@submit="saveMerchant"
			>
				<t-form-item
					:label="$t('merchant.title')"
					name="title"
				>
					<t-input v-model="form.title" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.legalRepresentative')"
					name="legalRepresentative"
				>
					<t-input v-model="form.legalRepresentative" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.businessAddress')"
					name="businessAddress"
				>
					<t-textarea v-model="form.businessAddress" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.businessDistrict')"
					name="businessDistrict"
				>
					<t-select v-model="form.businessDistrict">
						<t-option 
							v-for="district in businessDistrictOptions" 
							:key="district.value" 
							:value="district.value" 
							:label="district.label"
						/>
					</t-select>
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.validTime')"
					name="validTime"
				>
					<t-input v-model="form.validTime" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.trafficConditions')"
					name="trafficConditions"
				>
					<t-input v-model="form.trafficConditions" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.fixedEvents')"
					name="fixedEvents"
				>
					<t-input v-model="form.fixedEvents" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.terminalType')"
					name="terminalType"
				>
					<t-input v-model="form.terminalType" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.specialTimePeriods')"
					name="specialTimePeriods"
				>
					<t-input v-model="form.specialTimePeriods" />
				</t-form-item>
				
				<t-form-item
					:label="$t('merchant.customFilters')"
					name="customFilters"
				>
					<t-input v-model="form.customFilters" />
				</t-form-item>
			</t-form>
		</t-dialog>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import MerchantService from '@/services/merchant'
import { MessagePlugin } from 'tdesign-vue-next'
import type { IMerchant } from '@/modelTypes/IMerchant'

const { t } = useI18n()

// 表格相关
const merchants = ref<IMerchant[]>([])
const loading = ref(false)
const selectedRowKeys = ref<number[]>([])
const pagination = reactive({
	current: 1,
	pageSize: 10,
	total: 0,
})

// 表单相关
const showCreateModal = ref(false)
const editingMerchant = ref(null)
const formRef = ref<any>()
const form = reactive({
	id: 0,
	title: '',
	legalRepresentative: '',
	businessAddress: '',
	businessDistrict: '',
	validTime: '',
	trafficConditions: '',
	fixedEvents: '',
	terminalType: '',
	specialTimePeriods: '',
	customFilters: '',
})

const rules = {
	title: [{ required: true, message: t('merchant.rules.titleRequired'), trigger: 'blur' }],
	legalRepresentative: [{ required: true, message: t('merchant.rules.legalRepresentativeRequired'), trigger: 'blur' }],
	businessAddress: [{ required: true, message: t('merchant.rules.businessAddressRequired'), trigger: 'blur' }],
	businessDistrict: [{ required: true, message: t('merchant.rules.businessDistrictRequired'), trigger: 'change' }],
}

// 商圈选项
const businessDistrictOptions = [
	{ value: 'residential', label: t('merchant.businessDistrict.residential') },
	{ value: 'commercial', label: t('merchant.businessDistrict.commercial') },
	{ value: 'other', label: t('merchant.businessDistrict.other') },
]

// 表格列定义
const columns = [
	{ colKey: 'title', title: t('merchant.title'), width: 150 },
	{ colKey: 'legalRepresentative', title: t('merchant.legalRepresentative'), width: 120 },
	{ colKey: 'businessAddress', title: t('merchant.businessAddress'), width: 200 },
	{ colKey: 'businessDistrict', title: t('merchant.businessDistrict'), width: 120 },
	{ colKey: 'validTime', title: t('merchant.validTime'), width: 120 },
	{ colKey: 'trafficConditions', title: t('merchant.trafficConditions'), width: 120 },
	{ colKey: 'fixedEvents', title: t('merchant.fixedEvents'), width: 120 },
	{ colKey: 'terminalType', title: t('merchant.terminalType'), width: 120 },
	{ colKey: 'specialTimePeriods', title: t('merchant.specialTimePeriods'), width: 120 },
	{ colKey: 'customFilters', title: t('merchant.customFilters'), width: 120 },
	{ colKey: 'action', title: t('merchant.actions.title'), width: 150, fixed: 'right' },
]

// bulk delete & mapping editor state
const showBulkDeleteConfirm = ref(false)
const showMappingEditor = ref(false)
const fieldMappings = ref<any[]>([
	{ field: 'validTime', mappings: [{ placeholder: '', labelId: '' }] },
	{ field: 'trafficConditions', mappings: [{ placeholder: '', labelId: '' }] },
	{ field: 'fixedEvents', mappings: [{ placeholder: '', labelId: '' }] },
	{ field: 'specialTimePeriods', mappings: [{ placeholder: '', labelId: '' }] },
])

const confirmBulkDelete = (): void => {
	showBulkDeleteConfirm.value = true
}

const bulkDeleteSelected = async (): Promise<void> => {
	try {
		const service = new MerchantService()
		const ids = selectedRowKeys.value
		await service.bulkDelete(ids)
		MessagePlugin.success(t('merchant.messages.deleteMultipleSuccess', { count: ids.length }))
		// reset
		selectedRowKeys.value = []
		showBulkDeleteConfirm.value = false
		loadMerchants()
	} catch (err) {
		MessagePlugin.error(t('merchant.messages.deleteMultipleError'))
		console.error('Bulk delete error', err)
	}
}

const addMapping = (idx: number) => {
	fieldMappings.value[idx]?.mappings.push({ placeholder: '', labelId: '' })
}

const removeMapping = (idx: number, j: number) => {
	fieldMappings.value[idx]?.mappings.splice(j, 1)
}

const saveMappings = async (): Promise<void> => {
	// convert to field-specific mapping format and set to form.customFilters
	const out = fieldMappings.value.map(fm => ({ field: fm.field, mappings: fm.mappings.map(m => ({ placeholder: m.placeholder, labelId: Number(m.labelId) })) }))
	form.customFilters = JSON.stringify(out)
	MessagePlugin.success(t('merchant.mapping.saved'))
	showMappingEditor.value = false
}

const getBusinessDistrictLabel = (value: string) => {
	const option = businessDistrictOptions.find(opt => opt.value === value)
	return option ? option.label : value
}

// 获取商圈标签 (typed function is defined above)

// 获取商户列表
const loadMerchants = async () => {
	loading.value = true
	try {
		const service = new MerchantService()
		const result = await service.getAll({}, {}, pagination.current)
		
		merchants.value = result.map((merchant: any) => ({
			...merchant,
			created: new Date(merchant.created).toLocaleString(),
			updated: new Date(merchant.updated).toLocaleString(),
		})) as IMerchant[]
		
		pagination.total = service.resultCount
	} catch (error) {
		MessagePlugin.error(t('merchant.messages.loadError'))
		console.error('Failed to load merchants:', error)
	} finally {
		loading.value = false
	}
}

// 处理分页变化
const handlePageChange = (current: number, pageSize: number) => {
	pagination.current = current
	pagination.pageSize = pageSize
	loadMerchants()
}

// 处理选择变化
const handleSelectChange = (value: number[]) => {
	selectedRowKeys.value = value
}

// 编辑商户
const editMerchant = (merchant: any) => {
	editingMerchant.value = merchant
	Object.assign(form, merchant)
	showCreateModal.value = true
}

// 删除商户
const deleteMerchant = async (merchant: any) => {
	try {
		const service = new MerchantService()
		await service.delete(merchant)
		MessagePlugin.success(t('merchant.messages.deleteSuccess'))
		loadMerchants()
	} catch (error) {
		MessagePlugin.error(t('merchant.messages.deleteError'))
		console.error('Failed to delete merchant:', error)
	}
}

// 保存商户
const saveMerchant = async () => {
	const result = await formRef.value.validate()
	if (result !== true) return
	
	try {
		const service = new MerchantService()
		if (editingMerchant.value) {
			// 更新商户
			await service.update({ ...(form as any) } as any)
			MessagePlugin.success(t('merchant.messages.updateSuccess'))
		} else {
			// 创建商户
			await service.create({ ...(form as any) } as any)
			MessagePlugin.success(t('merchant.messages.createSuccess'))
		}

		resetForm()
		loadMerchants()
	} catch (error) {
		const message = editingMerchant.value 
			? t('merchant.messages.updateError') 
			: t('merchant.messages.createError')
		MessagePlugin.error(message)
		console.error('Failed to save merchant:', error)
	}
}

// 重置表单
const resetForm = () => {
	showCreateModal.value = false
	editingMerchant.value = null
	formRef.value?.reset()
	Object.assign(form, {
		id: 0,
		title: '',
		legalRepresentative: '',
		businessAddress: '',
		businessDistrict: '',
		validTime: '',
		trafficConditions: '',
		fixedEvents: '',
		terminalType: '',
		specialTimePeriods: '',
		customFilters: '',
	})
}

// 组件挂载时加载数据
onMounted(() => {
	loadMerchants()
})
</script>

<style scoped>
.merchant-page {
	padding: 20px;
}

.merchant-card {
	margin-bottom: 20px;
}

.action-button {
	margin-right: 10px;
}

:deep(.t-table) {
	margin-top: 20px;
}
</style>
