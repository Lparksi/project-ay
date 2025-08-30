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
					variant="outline"
					class="action-button"
					@click="triggerFileImport"
				>
					<template #icon>
						<t-icon name="upload" />
					</template>
					{{ $t('merchant.actions.importFromExcel') }}
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
				<t-button
					variant="danger"
					class="action-button"
					:disabled="merchants.length === 0"
					@click="confirmDeleteAll"
				>
					<template #icon>
						<t-icon name="delete" />
					</template>
					{{ $t('merchant.actions.deleteAll') }}
				</t-button>
				<!-- 隐藏的文件输入框 -->
				<input
					ref="fileInputRef"
					type="file"
					accept=".xlsx,.xls"
					style="display: none;"
					@change="handleFileImport"
				>
			</template>

			<!-- 数据概览区域 -->
			<div class="data-overview">
				<div class="stats-section">
					<t-tag
						theme="primary"
						size="large"
					>
						{{ $t('merchant.stats.total', { count: merchants.length }) }}
					</t-tag>
					<t-tag 
						v-if="hasActiveFilters" 
						theme="success" 
						size="large"
					>
						{{ $t('merchant.stats.filtered', { count: filteredMerchants.length }) }}
					</t-tag>
					<t-tag 
						v-if="selectedRowKeys.length > 0" 
						theme="warning" 
						size="large"
					>
						{{ $t('merchant.stats.selected', { count: selectedRowKeys.length }) }}
					</t-tag>
					<t-tag 
						v-if="!loading && merchants.length > 0 && !dataLoadError" 
						theme="default" 
						size="large"
					>
						{{ $t('merchant.pagination.currentPage', { 
							current: pagination.current, 
							total: Math.ceil(filteredMerchants.length / pagination.pageSize) || 1 
						}) }}
					</t-tag>
				</div>
			</div>

			<t-loading 
				:loading="loading" 
				:loading-props="{ 
					text: loadingProgress > 0 ? `${$t('merchant.loading.progress')} ${loadingProgress}%` : $t('merchant.loading.text'),
					size: 'large'
				}"
			>
				<!-- 数据为空时的友好提示 -->
				<div 
					v-if="!loading && merchants.length === 0 && !dataLoadError" 
					class="empty-state"
				>
					<t-icon
						name="inbox"
						size="48px"
						class="empty-icon"
					/>
					<h3>{{ $t('merchant.empty.title') }}</h3>
					<p>{{ $t('merchant.empty.description') }}</p>
					<t-button
						variant="outline"
						@click="showCreateModal = true"
					>
						<template #icon>
							<t-icon name="add" />
						</template>
						{{ $t('merchant.empty.createFirst') }}
					</t-button>
				</div>

				<!-- 数据加载错误提示 -->
				<div 
					v-else-if="!loading && dataLoadError" 
					class="error-state"
				>
					<t-icon
						name="error-circle"
						size="48px"
						class="error-icon"
					/>
					<h3>{{ $t('merchant.error.title') }}</h3>
					<p>{{ dataLoadError }}</p>
					<div class="error-actions">
						<t-button 
							variant="outline" 
							:loading="isRetrying"
							@click="retryLoadMerchants"
						>
							{{ $t('merchant.error.retry') }}
						</t-button>
						<t-button 
							variant="base" 
							@click="showCreateModal = true"
						>
							{{ $t('merchant.error.createAnyway') }}
						</t-button>
					</div>
				</div>

				<t-table
					v-else
					row-key="id"
					:data="currentPageData"
					:columns="columns"
					:selected-row-keys="selectedRowKeys"
					:filter-value="filterValue"
					:on-select-change="handleSelectChange"
					:on-filter-change="handleFilterChange"
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
						<t-tag
							v-if="row.validTime"
							theme="primary"
						>
							{{ row.validTime }}
						</t-tag>
					</template>
					
					<template #trafficConditions="{ row }">
						<t-tag
							v-if="row.trafficConditions"
							theme="success"
						>
							{{ row.trafficConditions }}
						</t-tag>
					</template>
					
					<template #fixedEvents="{ row }">
						<t-tag
							v-if="row.fixedEvents"
							theme="warning"
						>
							{{ row.fixedEvents }}
						</t-tag>
					</template>
					
					<template #terminalType="{ row }">
						<t-tag
							v-if="row.terminalType"
							theme="default"
						>
							{{ row.terminalType }}
						</t-tag>
					</template>
					
					<template #specialTimePeriods="{ row }">
						<t-tag
							v-if="row.specialTimePeriods"
							theme="danger"
						>
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

				<!-- 分页组件 -->
				<div
					v-if="!loading && merchants.length > 0 && !dataLoadError"
					style="margin-top: 20px; display: flex; justify-content: space-between; align-items: center;"
				>
					<div style="color: var(--text-secondary); font-size: 14px;">
						{{ $t('merchant.pagination.showing', { 
							start: Math.min((pagination.current - 1) * pagination.pageSize + 1, filteredMerchants.length),
							end: Math.min(pagination.current * pagination.pageSize, filteredMerchants.length),
							total: filteredMerchants.length
						}) }}
					</div>
					<t-pagination
						v-model:current="pagination.current"
						v-model:page-size="pagination.pageSize"
						:total="pagination.total"
						:page-size-options="pagination.pageSizeOptions"
						:show-size-changer="pagination.showSizeChanger"
						:show-quick-jumper="pagination.showQuickJumper"
						theme="simple"
						size="medium"
					/>
				</div>
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

		<!-- Delete all confirmation -->
		<t-dialog
			v-model:visible="showDeleteAllConfirm"
			:header="t('merchant.actions.deleteAll')"
			@confirm="deleteAllMerchants"
		>
			<p>{{ t('merchant.messages.confirmDeleteAll', {count: merchants.length}) }}</p>
			<p style="color: var(--error-color, #e34850); font-weight: 500; margin-top: 12px;">
				{{ t('merchant.messages.deleteAllWarning') }}
			</p>
		</t-dialog>

		<!-- Label mapping editor -->
		<t-dialog
			v-model:visible="showMappingEditor"
			:header="t('merchant.actions.editMappings')"
			width="800px"
			:confirm-btn="{ content: t('misc.save'), loading: mappingsLoading }"
			@confirm="saveMappings"
			@opened="loadMappings"
		>
			<div>
				<div style="display: flex; gap: 12px; margin-bottom: 16px; align-items: center;">
					<p>{{ t('merchant.mapping.help') }}</p>
					<t-button
						size="small"
						variant="outline"
						:disabled="mappingsLoading"
						@click="importDefaultMappings"
					>
						{{ t('merchant.mapping.importDefault') }}
					</t-button>
					<t-button
						size="small"
						variant="outline"
						:loading="mappingsLoading"
						@click="loadMappings"
					>
						{{ t('merchant.mapping.reload') }}
					</t-button>
					<t-button
						size="small"
						variant="base"
						theme="success"
						:disabled="mappingsLoading || fieldMappings.every(fm => fm.mappings.length === 0)"
						@click="testMappings"
					>
						{{ t('merchant.mapping.test') }}
					</t-button>
				</div>
				<div
					v-for="(fm, idx) in fieldMappings"
					:key="idx"
					style="margin-bottom:20px; padding: 16px; border: 1px solid var(--border-color, #e0e0e0); border-radius: 8px;"
				>
					<div style="font-weight:600; margin-bottom: 12px; color: var(--text-primary);">
						{{ getFieldDisplayName(fm.field) }}
					</div>
					<div style="display:flex;gap:8px;flex-direction: column;">
						<div
							v-for="(m, j) in fm.mappings"
							:key="j"
							style="display:flex;gap:8px;align-items:center; padding: 8px; background: var(--bg-color-container, #f5f5f5); border-radius: 4px;"
						>
							<div style="display: flex; flex-direction: column; gap: 4px; flex: 1;">
								<div style="font-size: 12px; color: var(--text-secondary); font-weight: 500;">
									{{ t('merchant.mapping.placeholder') }}
								</div>
								<t-input
									v-model="m.placeholder"
									:placeholder="t('merchant.mapping.placeholderHint')"
									style="width: 200px;"
									size="small"
								/>
							</div>
							<div style="display: flex; flex-direction: column; gap: 4px; flex: 2;">
								<div style="font-size: 12px; color: var(--text-secondary); font-weight: 500;">
									{{ t('merchant.mapping.displayText') }}
								</div>
								<t-input
									v-model="m.displayText"
									:placeholder="t('merchant.mapping.displayTextHint')"
									style="flex: 1;"
									size="small"
								/>
							</div>
							<div style="display: flex; align-items: center; gap: 4px;">
								<t-tag
									v-if="m.labelId"
									size="small"
									theme="primary"
								>
									{{ t('merchant.mapping.autoGenerated') }}
								</t-tag>
								<t-button
									size="small"
									variant="text"
									theme="danger"
									@click="removeMapping(idx, j)"
								>
									{{ t('merchant.mapping.remove') }}
								</t-button>
							</div>
						</div>
						<t-button
							size="small"
							variant="dashed"
							style="margin-top: 8px;"
							@click="addMapping(idx)"
						>
							<t-icon
								name="add"
								style="margin-right: 4px;"
							/>
							{{ t('merchant.mapping.add') }}
						</t-button>
					</div>
				</div>
				<div style="margin-top: 16px; padding: 12px; background: var(--bg-color-secondry-container, #f0f0f0); border-radius: 4px;">
					<div style="font-size: 12px; color: var(--text-secondary); margin-bottom: 8px;">
						{{ t('merchant.mapping.note') }}
					</div>
					<div style="font-size: 11px; color: var(--text-placeholder);">
						{{ t('merchant.mapping.noteDetail') }}
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
				<!-- selection handled by table (type: 'multiple') -->
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

		<!-- Excel导入预览对话框 -->
		<t-dialog
			v-model:visible="showImportPreview"
			:header="t('merchant.import.previewTitle')"
			width="90%"
			:confirm-btn="{ content: t('merchant.import.confirmImport'), loading: importLoading }"
			@confirm="confirmImport"
			@cancel="cancelImport"
		>
			<div v-if="importData.length > 0">
				<div style="margin-bottom: 16px; padding: 12px; background: var(--bg-color-container, #f5f5f5); border-radius: 4px;">
					<div style="font-weight: 600; margin-bottom: 8px; color: var(--text-primary);">
						{{ t('merchant.import.summary') }}
					</div>
					<div style="color: var(--text-secondary);">
						{{ t('merchant.import.totalRecords', { count: importData.length }) }}
					</div>
					<div style="color: var(--text-secondary); margin-top: 4px;">
						{{ t('merchant.import.supportedFields') }}
					</div>
					<div style="color: var(--text-secondary); margin-top: 4px; font-style: italic;">
						{{ t('merchant.import.mappingApplied') }}
					</div>
				</div>

				<t-table
					:data="importData.slice(0, 10)"
					:columns="importPreviewColumns"
					border
					size="small"
					style="margin-bottom: 16px;"
				/>

				<div
					v-if="importData.length > 10"
					style="text-align: center; color: var(--text-secondary); font-size: 12px;"
				>
					{{ t('merchant.import.previewNote', { shown: 10, total: importData.length }) }}
				</div>
			</div>
			<div
				v-else
				style="text-align: center; color: var(--text-secondary); padding: 40px;"
			>
				{{ t('merchant.import.noData') }}
			</div>
		</t-dialog>
	</div>
</template>

<script setup lang="ts">
// 导入computed
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import MerchantService from '@/services/merchant'
import MerchantMappingService from '@/services/merchantMapping'
import MerchantModel from '@/models/merchant'
import { MessagePlugin } from 'tdesign-vue-next'
import type { IMerchant, IFieldLabelMapping } from '@/modelTypes/IMerchant'
import * as XLSX from 'xlsx'

const { t } = useI18n()

// 表格相关
const merchants = ref<IMerchant[]>([])
const loading = ref(false)
const selectedRowKeys = ref<number[]>([])
// 使用TDesign内置过滤功能
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const filterValue = ref<Record<string, any>>({})
// 数据加载状态优化
const loadingProgress = ref(0)
const dataLoadError = ref<string | null>(null)
const isRetrying = ref(false)

// 分页相关
const pagination = reactive({
	current: 1,
	pageSize: 20,
	total: 0,
	pageSizeOptions: [10, 20, 50, 100],
	showSizeChanger: true,
	showQuickJumper: true,
})

// 过滤后的商户数据
const filteredMerchants = computed(() => {
	if (!merchants.value.length) return []
	
	// 如果没有筛选条件，返回所有数据
	if (!hasActiveFilters.value) {
		return merchants.value
	}
	
	// 应用筛选条件
	return merchants.value.filter(merchant => {
		for (const [key, value] of Object.entries(filterValue.value)) {
			if (!value || value === '') continue
			
			const merchantValue = merchant[key as keyof IMerchant]?.toString().toLowerCase() || ''
			const filterStr = value.toString().toLowerCase()
			
			if (!merchantValue.includes(filterStr)) {
				return false
			}
		}
		return true
	})
})

// 当前页数据
const currentPageData = computed(() => {
	const start = (pagination.current - 1) * pagination.pageSize
	const end = start + pagination.pageSize
	return filteredMerchants.value.slice(start, end)
})

// 更新分页总数
watch(filteredMerchants, (newData) => {
	pagination.total = newData.length
	// 如果当前页超出范围，跳转到第一页
	if (pagination.current > Math.ceil(newData.length / pagination.pageSize)) {
		pagination.current = 1
	}
}, { immediate: true })

// 表单相关
const showCreateModal = ref(false)
const editingMerchant = ref<IMerchant | null>(null)
const formRef = ref<HTMLFormElement | null>(null)
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
	legalRepresentative: [{ required: true, message: t('merchant.rules.legalRepresentativeRequired'), trigger: 'blur' }],
	businessAddress: [{ required: true, message: t('merchant.rules.businessAddressRequired'), trigger: 'blur' }],
}

// 商圈选项
const businessDistrictOptions = [
	{ value: 'residential', label: t('merchant.businessDistrict.residential') },
	{ value: 'commercial', label: t('merchant.businessDistrict.commercial') },
	{ value: 'other', label: t('merchant.businessDistrict.other') },
]

// 表格列定义
const columns = [
	{ type: 'multiple', colKey: 'select', title: '', width: 60 },
	{ 
		colKey: 'title', 
		title: t('merchant.title'), 
		width: 150,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.titlePlaceholder'),
			},
		},
	},
	{ 
		colKey: 'legalRepresentative', 
		title: t('merchant.legalRepresentative'), 
		width: 120,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.legalRepresentativePlaceholder'),
			},
		},
	},
	{ 
		colKey: 'businessAddress', 
		title: t('merchant.businessAddress'), 
		width: 200,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.businessAddressPlaceholder'),
			},
		},
	},
	{ 
		colKey: 'businessDistrict', 
		title: t('merchant.businessDistrict'), 
		width: 120,
		filter: {
			type: 'single',
			list: businessDistrictOptions,
			resetValue: '',
		},
	},
	{ 
		colKey: 'validTime', 
		title: t('merchant.validTime'), 
		width: 120,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.validTimePlaceholder'),
			},
		},
	},
	{ 
		colKey: 'trafficConditions', 
		title: t('merchant.trafficConditions'), 
		width: 120,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.trafficConditionsPlaceholder'),
			},
		},
	},
	{ 
		colKey: 'fixedEvents', 
		title: t('merchant.fixedEvents'), 
		width: 120,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.fixedEventsPlaceholder'),
			},
		},
	},
	{ 
		colKey: 'terminalType', 
		title: t('merchant.terminalType'), 
		width: 120,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.terminalTypePlaceholder'),
			},
		},
	},
	{ 
		colKey: 'specialTimePeriods', 
		title: t('merchant.specialTimePeriods'), 
		width: 120,
		filter: {
			type: 'input',
			resetValue: '',
			confirmEvents: ['onEnter'],
			props: {
				placeholder: t('merchant.search.specialTimePeriodsPlaceholder'),
			},
		},
	},
	{ colKey: 'customFilters', title: t('merchant.customFilters'), width: 120 },
	{ colKey: 'action', title: t('merchant.actions.title'), width: 150, fixed: 'right' },
]

// bulk delete & mapping editor state
const showBulkDeleteConfirm = ref(false)
const showDeleteAllConfirm = ref(false)
const showMappingEditor = ref(false)
type FieldMapping = { 
	field: string; 
	mappings: Array<{ 
		placeholder: string; 
		displayText: string; 
		labelId: string | number 
	}> 
}
const fieldMappings = ref<FieldMapping[]>([
	{ field: 'validTime', mappings: [] },
	{ field: 'trafficConditions', mappings: [] },
	{ field: 'fixedEvents', mappings: [] },
	{ field: 'terminalType', mappings: [] },
	{ field: 'specialTimePeriods', mappings: [] },
])

// 映射相关变量
const mappingService = new MerchantMappingService()
const mappingsLoading = ref(false)

// Excel导入数据类型定义
interface ImportMerchantData {
	validTime: string
	trafficConditions: string
	fixedEvents: string
	terminalType: string
	specialTimePeriods: string
	customFilters: string
}

// Excel导入相关状态
const showImportPreview = ref(false)
const importData = ref<ImportMerchantData[]>([])
const importLoading = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

// 导入预览表格列定义
const importPreviewColumns = [
	{ colKey: 'validTime', title: t('merchant.validTime'), width: 120 },
	{ colKey: 'trafficConditions', title: t('merchant.trafficConditions'), width: 120 },
	{ colKey: 'fixedEvents', title: t('merchant.fixedEvents'), width: 120 },
	{ colKey: 'terminalType', title: t('merchant.terminalType'), width: 120 },
	{ colKey: 'specialTimePeriods', title: t('merchant.specialTimePeriods'), width: 120 },
	{ colKey: 'customFilters', title: t('merchant.customFilters'), width: 120 },
]

// 自动递增的标签ID计数器
let labelIdCounter = 1000

// 默认映射数据
const defaultMappingData = {
	validTime: [
		{ key: 'A', value: 'A：开门较晚，上午10点后拜访' },
		{ key: 'B', value: 'B：中午门店没有人' },
		{ key: 'C', value: 'C：关门时间较晚' },
	],
	trafficConditions: [
		{ key: 'A', value: 'A：附近学校，中午傍晚绕行' },
		{ key: 'B', value: 'B：上午有集市，建议下午拜访' },
		{ key: 'C', value: 'C：麦收期间道路拥堵' },
		{ key: 'D', value: 'D：周边道路施工' },
	],
	fixedEvents: [
		{ key: 'A', value: 'A：老人家在家，订货人特定时间在店内' },
		{ key: 'B', value: 'B：下午麻将场，建议上午拜访' },
		{ key: 'C', value: 'C：门店装修' },
	],
	terminalType: [
		{ key: 'A', value: 'A：加盟终端' },
		{ key: 'B', value: 'B：现代终端' },
		{ key: 'C', value: 'C：普通终端' },
	],
	specialTimePeriods: [
		{ key: 'A', value: 'A：大学周边，寒暑假期间营业时间不固定' },
		{ key: 'B', value: 'B：农忙季节' },
		{ key: 'C', value: 'C：周期性集市' },
	],
}

const confirmBulkDelete = (): void => {
	showBulkDeleteConfirm.value = true
}

const confirmDeleteAll = (): void => {
	showDeleteAllConfirm.value = true
}

const deleteAllMerchants = async (): Promise<void> => {
	try {
		const service = new MerchantService()
		const allIds = merchants.value.map(m => m.id)
		await service.bulkDelete(allIds)
		MessagePlugin.success(t('merchant.messages.deleteAllSuccess', { count: allIds.length }))
		// reset
		selectedRowKeys.value = []
		showDeleteAllConfirm.value = false
		// 清除缓存并重新加载
		clearCache()
		loadMerchants(false)
	} catch (err) {
		MessagePlugin.error(t('merchant.messages.deleteAllError'))
		console.error('Delete all merchants error', err)
	}
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
	const newMapping = {
		placeholder: '', 
		displayText: '', 
		labelId: generateLabelId(),
	}
	fieldMappings.value[idx]?.mappings.push(newMapping)
	console.log(`Added new mapping with labelId: ${newMapping.labelId} for field: ${fieldMappings.value[idx]?.field}`)
}

const removeMapping = (idx: number, j: number) => {
	const removedMapping = fieldMappings.value[idx]?.mappings[j]
	if (removedMapping) {
		console.log(`Removing mapping with labelId: ${removedMapping.labelId} from field: ${fieldMappings.value[idx]?.field}`)
	}
	fieldMappings.value[idx]?.mappings.splice(j, 1)
}

// 测试映射功能
const testMappings = async (): Promise<void> => {
	console.log('开始测试映射功能...')
	
	// 创建测试数据
	const testData: ImportMerchantData[] = [
		{
			validTime: 'A',
			trafficConditions: 'B',
			fixedEvents: 'C',
			terminalType: 'A',
			specialTimePeriods: 'A',
			customFilters: '',
		},
	]
	
	console.log('测试数据:', testData)
	
	try {
		const result = await applyMappingsToImportData(testData)
		console.log('测试结果:', result)
		
		// 检查是否有映射被应用
		let hasAnyMappingApplied = false
		for (const field of Object.keys(testData[0])) {
			if (testData[0][field as keyof ImportMerchantData] !== result[0][field as keyof ImportMerchantData]) {
				hasAnyMappingApplied = true
				break
			}
		}
		
		if (hasAnyMappingApplied) {
			MessagePlugin.success('映射测试成功！映射功能正常工作')
		} else {
			MessagePlugin.warning('映射测试完成，但没有找到匹配的映射规则。请检查您的映射配置。')
		}
	} catch (error) {
		console.error('映射测试失败:', error)
		MessagePlugin.error('映射测试失败: ' + (error instanceof Error ? error.message : '未知错误'))
	}
}

// 生成唯一标签ID，确保从1000开始递增
const generateLabelId = (): number => {
	return ++labelIdCounter
}

// 获取字段显示名称
const getFieldDisplayName = (field: string): string => {
	const fieldNameMap: Record<string, string> = {
		validTime: t('merchant.validTime'),
		trafficConditions: t('merchant.trafficConditions'),
		fixedEvents: t('merchant.fixedEvents'),
		terminalType: t('merchant.terminalType'),
		specialTimePeriods: t('merchant.specialTimePeriods'),
	}
	return fieldNameMap[field] || field
}

// 导入默认映射
const importDefaultMappings = (): void => {
	type DefaultMappingItem = { key: string; value: string }
	const typedMappingData = defaultMappingData as Record<string, DefaultMappingItem[]>
	
	// 重置labelIdCounter以确保从1000开始
	labelIdCounter = 999
	
	fieldMappings.value = Object.keys(defaultMappingData).map(field => {
		const fieldData = typedMappingData[field]
		return {
			field,
			mappings: fieldData ? fieldData.map((item: DefaultMappingItem) => ({
				placeholder: item.key.trim(),
				displayText: item.value.trim(),
				labelId: generateLabelId(),
			})) : [],
		}
	})
	MessagePlugin.success(t('merchant.mapping.importSuccess'))
}

// 加载映射
const loadMappings = async (): Promise<void> => {
	try {
		mappingsLoading.value = true
		console.log('开始加载映射数据...')
		
		const mappings = await mappingService.loadAllMappings()
		console.log('从服务端加载的映射数据:', mappings)
		
		if (mappings.length > 0) {
			// 转换为前端格式
			fieldMappings.value = mappings.map(mapping => ({
				field: mapping.field,
				mappings: mapping.mappings.map(m => ({
					placeholder: m.placeholder,
					displayText: m.displayText,
					labelId: m.labelId,
				})),
			}))
			
			console.log('转换后的前端映射数据:', fieldMappings.value)
			
			// 更新labelIdCounter以避免冲突，确保从最大值开始
			let maxId = 999 // 从999开始，下一个生成的ID将是1000
			mappings.forEach(mapping => {
				mapping.mappings.forEach(m => {
					if (m.labelId > maxId) {
						maxId = m.labelId
					}
				})
			})
			labelIdCounter = maxId
			console.log('更新labelIdCounter为:', labelIdCounter)
			
			// 确保所有映射都有有效的labelId
			fieldMappings.value.forEach(fm => {
				fm.mappings.forEach(m => {
					if (!m.labelId || m.labelId <= 0) {
						m.labelId = generateLabelId()
						console.log(`为字段 ${fm.field} 的映射生成新ID: ${m.labelId}`)
					}
				})
			})
			MessagePlugin.success(`成功加载 ${mappings.length} 个字段的映射配置`)
		} else {
			console.log('没有找到映射数据，初始化空映射')
			// 如果没有数据，初始化空映射
			fieldMappings.value = [
				{ field: 'validTime', mappings: [] },
				{ field: 'trafficConditions', mappings: [] },
				{ field: 'fixedEvents', mappings: [] },
				{ field: 'terminalType', mappings: [] },
				{ field: 'specialTimePeriods', mappings: [] },
			]
			MessagePlugin.info('暂无映射配置，您可以点击"导入默认映射"快速配置')
		}
	} catch (error) {
		console.error('Failed to load mappings:', error)
		MessagePlugin.error('映射加载失败: ' + (error instanceof Error ? error.message : '未知错误'))
		// 加载失败时初始化空映射
		fieldMappings.value = [
			{ field: 'validTime', mappings: [] },
			{ field: 'trafficConditions', mappings: [] },
			{ field: 'fixedEvents', mappings: [] },
			{ field: 'terminalType', mappings: [] },
			{ field: 'specialTimePeriods', mappings: [] },
		]
	} finally {
		mappingsLoading.value = false
	}
}

const saveMappings = async (): Promise<void> => {
	console.log('开始保存映射...')
	console.log('当前 fieldMappings 数据:', fieldMappings.value)
	
	// 为每个映射生成或确保有labelId
	fieldMappings.value.forEach(fm => {
		fm.mappings.forEach(m => {
			if (!m.labelId || m.labelId <= 0) {
				m.labelId = generateLabelId()
				console.log(`为字段 ${fm.field} 的映射生成新ID: ${m.labelId}`)
			}
		})
	})

	try {
		mappingsLoading.value = true
		
		// 转换为后端需要的格式，确保字段名正确映射
		const mappingsToSave: IFieldLabelMapping[] = fieldMappings.value
			.filter(fm => fm.mappings.length > 0) // 过滤没有映射的字段
			.map((fm: FieldMapping) => ({
				field: fm.field,
				mappings: fm.mappings
					.filter(m => m.placeholder.trim() !== '' && m.displayText.trim() !== '') // 过滤空映射
					.map(m => ({
						placeholder: m.placeholder.trim(),
						displayText: m.displayText.trim(),
						labelId: Number(m.labelId),
					})),
			}))
			.filter(fm => fm.mappings.length > 0) // 再次过滤，确保只保存有效映射

		console.log('准备保存的映射数据:', mappingsToSave)
		
		if (mappingsToSave.length === 0) {
			MessagePlugin.warning('没有有效的映射数据需要保存')
			showMappingEditor.value = false
			return
		}

		// 保存到数据库
		const result = await mappingService.bulkSaveMappings(mappingsToSave)
		console.log('服务端返回结果:', result)
		
		MessagePlugin.success(t('merchant.mapping.saved'))
		showMappingEditor.value = false
		
		// 保存成功后重新加载，确保数据同步
		await loadMappings()
	} catch (error) {
		console.error('Failed to save mappings:', error)
		const errorMessage = error instanceof Error ? error.message : '未知错误'
		MessagePlugin.error(t('merchant.mapping.saveError') + ': ' + errorMessage)
	} finally {
		mappingsLoading.value = false
	}
}

const getBusinessDistrictLabel = (value: string) => {
	const option = businessDistrictOptions.find(opt => opt.value === value)
	return option ? option.label : value
}

// 获取商圈标签 (typed function is defined above)

// 数据缓存机制
let merchantsCache: IMerchant[] | null = null
let cacheTimestamp: number = 0
const CACHE_DURATION = 5 * 60 * 1000 // 5分钟缓存

// 检查缓存是否有效
const isCacheValid = (): boolean => {
	return merchantsCache !== null && (Date.now() - cacheTimestamp) < CACHE_DURATION
}

// 模拟进度更新
const simulateProgress = (): Promise<void> => {
	return new Promise((resolve) => {
		loadingProgress.value = 0
		const interval = setInterval(() => {
			loadingProgress.value += Math.random() * 30
			if (loadingProgress.value >= 90) {
				clearInterval(interval)
				loadingProgress.value = 100
				setTimeout(resolve, 200)
			}
		}, 200)
	})
}

// 获取商户列表
const loadMerchants = async (useCache: boolean = true) => {
	// 检查缓存
	if (useCache && isCacheValid() && merchantsCache) {
		merchants.value = merchantsCache
		return
	}

	loading.value = true
	dataLoadError.value = null
	loadingProgress.value = 0

	try {
		// 开始进度模拟
		const progressPromise = simulateProgress()
		
		const service = new MerchantService()
		// 不使用分页，获取所有数据
		const result = await service.getAll(new MerchantModel(), {}, -1) // page: -1 表示获取所有数据

		// 等待进度模拟完成
		await progressPromise

		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const processedData = result.map((merchant: any) => ({
			...merchant,
			created: new Date(merchant.created),
			updated: new Date(merchant.updated),
		})) as IMerchant[]

		// 更新数据和缓存
		merchants.value = processedData
		merchantsCache = processedData
		cacheTimestamp = Date.now()

		// 数据加载成功提示
		if (processedData.length > 0) {
			MessagePlugin.success(t('merchant.messages.loadSuccess', { count: processedData.length }))
		}
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : t('merchant.messages.loadError')
		dataLoadError.value = errorMessage
		MessagePlugin.error(errorMessage)
		console.error('Failed to load merchants:', error)
	} finally {
		loading.value = false
		loadingProgress.value = 0
	}
}

// 重试加载数据
const retryLoadMerchants = async () => {
	isRetrying.value = true
	try {
		await loadMerchants(false) // 强制刷新，不使用缓存
	} finally {
		isRetrying.value = false
	}
}

// 清除缓存
const clearCache = () => {
	merchantsCache = null
	cacheTimestamp = 0
}

// 计算是否有活跃的筛选器
const hasActiveFilters = computed(() => {
	return Object.keys(filterValue.value).some(key => {
		const value = filterValue.value[key]
		return value !== undefined && value !== '' && value !== null && 
			(Array.isArray(value) ? value.length > 0 : true)
	})
})

// TDesign筛选事件处理
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const handleFilterChange = (filters: Record<string, any>) => {
	filterValue.value = filters
}

// 处理选择变化
const handleSelectChange = (value: number[]) => {
	selectedRowKeys.value = value
}

// 编辑商户
const editMerchant = (merchant: IMerchant) => {
	editingMerchant.value = merchant
	Object.assign(form, merchant)
	showCreateModal.value = true
}

// 删除商户
const deleteMerchant = async (merchant: IMerchant) => {
	try {
		const service = new MerchantService()
		await service.delete(merchant)
		MessagePlugin.success(t('merchant.messages.deleteSuccess'))
		// 清除缓存并重新加载
		clearCache()
		loadMerchants(false)
	} catch (error) {
		MessagePlugin.error(t('merchant.messages.deleteError'))
		console.error('Failed to delete merchant:', error)
	}
}

// 保存商户
const saveMerchant = async () => {
	if (!formRef.value) return
	
	const result = await formRef.value.validate()
	if (result !== true) return
	
	try {
		const service = new MerchantService()
		if (editingMerchant.value) {
			// 更新商户
			await service.update({ ...form } as IMerchant)
			MessagePlugin.success(t('merchant.messages.updateSuccess'))
		} else {
			// 创建商户
			await service.create({ ...form } as IMerchant)
			MessagePlugin.success(t('merchant.messages.createSuccess'))
		}

		resetForm()
		// 清除缓存并重新加载
		clearCache()
		loadMerchants(false)
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

// Excel导入相关函数
// 触发文件导入
const triggerFileImport = () => {
	fileInputRef.value?.click()
}

// 处理文件导入
const handleFileImport = async (event: Event) => {
	const target = event.target as HTMLInputElement
	const file = target.files?.[0]
	
	if (!file) {
		MessagePlugin.warning(t('merchant.import.noFileSelected'))
		return
	}

	// 检查文件类型
	const fileName = file.name.toLowerCase()
	if (!fileName.endsWith('.xlsx') && !fileName.endsWith('.xls')) {
		MessagePlugin.error(t('merchant.import.invalidFileType'))
		return
	}

	try {
		// 解析Excel文件
		await parseExcelFile(file)
	} catch (error) {
		console.error('Error parsing file:', error)
		MessagePlugin.error(t('merchant.import.parseError'))
	}

	// 清空文件输入
	target.value = ''
}

// 解析Excel文件
const parseExcelFile = async (file: File): Promise<void> => {
	const arrayBuffer = await file.arrayBuffer()
	const workbook = XLSX.read(arrayBuffer, { type: 'array' })
	
	// 获取第一个工作表
	const firstSheetName = workbook.SheetNames[0]
	if (!firstSheetName) {
		MessagePlugin.error(t('merchant.import.emptyFile'))
		return
	}
	
	const worksheet = workbook.Sheets[firstSheetName]
	if (!worksheet) {
		MessagePlugin.error(t('merchant.import.emptyFile'))
		return
	}
	
	// 转换为JSON数据
	const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 }) as unknown[][]
	
	if (jsonData.length < 2) {
		MessagePlugin.error(t('merchant.import.emptyFile'))
		return
	}

	// 解析表头
	const headers = jsonData[0] as unknown[]
	
	// 检查必需的列
	const requiredHeaders = ['有效时间', '交通情况', '固定事件', '终端类型', '特殊时段', '自定义筛选']
	const headerIndexMap: Record<string, number> = {}
	
	requiredHeaders.forEach(header => {
		const index = headers.findIndex(h => h?.toString().trim() === header)
		if (index === -1) {
			MessagePlugin.error(t('merchant.import.missingHeaders', { headers: header }))
			return
		}
		headerIndexMap[header] = index
	})

	// 如果有缺失的表头就返回
	if (Object.keys(headerIndexMap).length !== requiredHeaders.length) {
		return
	}

	// 解析数据行
	const data: ImportMerchantData[] = []
	for (let i = 1; i < jsonData.length; i++) {
		const row = jsonData[i] as unknown[]
		if (row && row.length > 0) {
			const getFieldValue = (fieldName: string): string => {
				const index = headerIndexMap[fieldName]
				return index !== undefined && row[index] !== undefined && row[index] !== null 
					? String(row[index]).trim() 
					: ''
			}
			
			const merchantData: ImportMerchantData = {
				validTime: getFieldValue('有效时间'),
				trafficConditions: getFieldValue('交通情况'),
				fixedEvents: getFieldValue('固定事件'),
				terminalType: getFieldValue('终端类型'),
				specialTimePeriods: getFieldValue('特殊时段'),
				customFilters: getFieldValue('自定义筛选'),
			}
			// 过滤掉完全空的行
			if (Object.values(merchantData).some(value => value !== '')) {
				data.push(merchantData)
			}
		}
	}

	if (data.length === 0) {
		MessagePlugin.error(t('merchant.import.noValidData'))
		return
	}

	// 在预览前应用映射
	console.log('Original import data:', data)
	const processedData = await applyMappingsToImportData(data)
	console.log('Processed import data after mapping:', processedData)

	// 设置导入数据并显示预览
	importData.value = processedData
	showImportPreview.value = true
}

// 确认导入
const confirmImport = async () => {
	if (importData.value.length === 0) {
		MessagePlugin.warning(t('merchant.import.noData'))
		return
	}

	importLoading.value = true
	try {
		const service = new MerchantService()
		let successCount = 0
		let errorCount = 0

		// 批量创建商户（后端会自动应用数据库映射）
		for (const merchantData of importData.value) {
			try {
				// 为每个商户设置必需的字段
				const merchant: Partial<IMerchant> = {
					title: `导入商户_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
					legalRepresentative: '待完善',
					businessAddress: '待完善',
					businessDistrict: 'other',
					...merchantData,
				}
				await service.create(merchant as IMerchant)
				successCount++
			} catch (error) {
				console.error('Error creating merchant:', error)
				errorCount++
			}
		}

		if (successCount > 0) {
			MessagePlugin.success(t('merchant.import.successMessage', { 
				successCount, 
				errorCount, 
				total: importData.value.length, 
			}))
			loadMerchants() // 重新加载商户列表
		} else {
			MessagePlugin.error(t('merchant.import.allFailed'))
		}

		cancelImport()
	} catch (error) {
		console.error('Import error:', error)
		MessagePlugin.error(t('merchant.import.importError'))
	} finally {
		importLoading.value = false
	}
}

// 应用映射到导入数据
const applyMappingsToImportData = async (data: ImportMerchantData[]): Promise<ImportMerchantData[]> => {
	console.log('开始应用映射到导入数据...')
	console.log('原始导入数据:', data)
	
	try {
		// 加载所有映射
		const mappings = await mappingService.loadAllMappings()
		console.log('加载的映射数据:', mappings)
		
		if (mappings.length === 0) {
			console.log('No mappings found, returning original data')
			MessagePlugin.info('没有找到映射配置，将使用原始数据')
			return data
		}
		
		// 构建映射查找表
		const mappingMap: Record<string, Record<string, string>> = {}
		mappings.forEach(fieldMapping => {
			if (!mappingMap[fieldMapping.field]) {
				mappingMap[fieldMapping.field] = {}
			}
			fieldMapping.mappings.forEach(mapping => {
				const fieldMap = mappingMap[fieldMapping.field]
				if (fieldMap) {
					// 确保占位符和显示文本都经过处理
					const cleanPlaceholder = mapping.placeholder.trim()
					const cleanDisplayText = mapping.displayText.trim()
					if (cleanPlaceholder && cleanDisplayText) {
						fieldMap[cleanPlaceholder] = cleanDisplayText
						console.log(`添加映射: ${fieldMapping.field}.${cleanPlaceholder} -> ${cleanDisplayText}`)
					}
				}
			})
		})
		
		console.log('Mapping lookup table:', mappingMap)
		
		// 应用映射到数据
		let totalApplied = 0
		const processedData = data.map((item, index) => {
			const processedItem = { ...item }
			let itemChanged = false
			
			// 对每个字段应用映射
			Object.keys(mappingMap).forEach(fieldName => {
				const originalValue = processedItem[fieldName as keyof ImportMerchantData]
				if (originalValue) {
					const cleanOriginalValue = originalValue.toString().trim()
					if (cleanOriginalValue && mappingMap[fieldName] && mappingMap[fieldName][cleanOriginalValue]) {
						const newValue = mappingMap[fieldName][cleanOriginalValue]
						// 更新字段值
						(processedItem as Record<string, string>)[fieldName] = newValue
						itemChanged = true
						totalApplied++
						console.log(`行 ${index + 1} 字段 ${fieldName}: "${cleanOriginalValue}" -> "${newValue}"`)
					}
				}
			})
			
			if (itemChanged) {
				console.log(`行 ${index + 1} 应用映射后:`, processedItem)
			}
			
			return processedItem
		})
		
		console.log(`映射应用完成，共应用 ${totalApplied} 个映射`)
		if (totalApplied > 0) {
			MessagePlugin.success(`成功应用 ${totalApplied} 个映射转换`)
		} else {
			MessagePlugin.info('没有找到匹配的映射，使用原始数据')
		}
		
		return processedData
	} catch (error) {
		console.error('Failed to apply mappings:', error)
		const errorMessage = error instanceof Error ? error.message : '未知错误'
		// 如果映射应用失败，返回原始数据并提示用户
		MessagePlugin.warning(t('merchant.import.mappingApplyWarning') + ': ' + errorMessage)
		return data
	}
}

// 取消导入
const cancelImport = () => {
	showImportPreview.value = false
	importData.value = []
	importLoading.value = false
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

.data-overview {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 20px;
	padding: 16px;
	background: var(--bg-color-container, #f5f5f5);
	border-radius: 8px;
	border: 1px solid var(--border-color, #e0e0e0);
}

.search-section {
	display: flex;
	gap: 12px;
	align-items: center;
}

.search-input {
	width: 300px;
}

.stats-section {
	display: flex;
	gap: 12px;
	align-items: center;
}

:deep(.t-table) {
	margin-top: 20px;
}

/* 修复多选框与单元格对齐问题 */
:deep(.t-table th[data-col-key="select"], .t-table td[data-col-key="select"]) {
    text-align: center !important;
    vertical-align: middle !important;
    padding: 8px !important;
}

:deep(.t-table th[data-col-key="select"] .t-table__cell-inner, 
       .t-table td[data-col-key="select"] .t-table__cell-inner) {
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
    height: 100% !important;
    min-height: 40px !important;
}

:deep(.t-table th[data-col-key="select"] .t-checkbox, 
       .t-table td[data-col-key="select"] .t-checkbox) {
    margin: 0 !important;
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
}

:deep(.t-table th[data-col-key="select"] .t-checkbox__input, 
       .t-table td[data-col-key="select"] .t-checkbox__input) {
    margin: 0 !important;
    vertical-align: middle !important;
}

/* 深夜模式下的多选框居中优化 */
:deep(.dark .t-table th[data-col-key="select"], 
       .dark .t-table td[data-col-key="select"]) {
    text-align: center !important;
    display: table-cell !important;
    vertical-align: middle !important;
}

:deep(.dark .t-table th[data-col-key="select"] .t-table__cell-inner, 
       .dark .t-table td[data-col-key="select"] .t-table__cell-inner) {
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
    width: 100% !important;
    height: 100% !important;
    position: relative !important;
}

:deep(.dark .t-table th[data-col-key="select"] .t-checkbox, 
       .dark .t-table td[data-col-key="select"] .t-checkbox) {
    margin: 0 auto !important;
    display: block !important;
    position: static !important;
    transform: none !important;
}

:deep(.dark .t-table th[data-col-key="select"] .t-checkbox .t-checkbox__input, 
       .dark .t-table td[data-col-key="select"] .t-checkbox .t-checkbox__input) {
    margin: 0 !important;
    position: static !important;
    transform: none !important;
    display: block !important;
}

/* 确保表格行高度一致 */
:deep(.t-table tbody tr) {
    height: 48px !important;
}

:deep(.t-table thead tr) {
    height: 48px !important;
}

/* 深夜模式下确保行高度一致性 */
:deep(.dark .t-table tbody tr) {
    height: 48px !important;
    line-height: 48px !important;
}

:deep(.dark .t-table thead tr) {
    height: 48px !important;
    line-height: 48px !important;
}

/* 空状态样式 */
.empty-state {
	text-align: center;
	padding: 60px 20px;
	color: var(--text-secondary);
}

.empty-state .empty-icon {
	color: var(--text-placeholder);
	margin-bottom: 16px;
}

.empty-state h3 {
	margin: 16px 0 8px 0;
	font-size: 18px;
	font-weight: 500;
	color: var(--text-primary);
}

.empty-state p {
	margin: 0 0 24px 0;
	font-size: 14px;
	line-height: 1.5;
}

/* 错误状态样式 */
.error-state {
	text-align: center;
	padding: 60px 20px;
	color: var(--text-secondary);
}

.error-state .error-icon {
	color: var(--error-color, #e34850);
	margin-bottom: 16px;
}

.error-state h3 {
	margin: 16px 0 8px 0;
	font-size: 18px;
	font-weight: 500;
	color: var(--text-primary);
}

.error-state p {
	margin: 0 0 24px 0;
	font-size: 14px;
	line-height: 1.5;
	max-width: 400px;
	margin-left: auto;
	margin-right: auto;
}

.error-actions {
	display: flex;
	gap: 12px;
	justify-content: center;
	align-items: center;
}
</style>
