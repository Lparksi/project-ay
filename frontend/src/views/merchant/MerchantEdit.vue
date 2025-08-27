<template>
	<div class="content merchant-edit">
		<header class="merchant-header">
			<div class="fancylists-title">
				<h1>{{ isEdit ? $t('merchant.edit.title') : $t('merchant.create.title') }}</h1>
				<p>{{ isEdit ? $t('merchant.edit.description') : $t('merchant.create.description') }}</p>
			</div>
			<div class="filter-container">
				<div class="items">
					<XButton
						icon="arrow-left"
						:to="{name: 'merchants.index'}"
						variant="secondary"
					>
						{{ $t('misc.back') }}
					</XButton>
				</div>
			</div>
		</header>

		<Card>
			<form @submit.prevent="saveMerchant">
				<div class="field">
					<label class="label">{{ $t('merchant.attributes.title') }}</label>
					<div class="control">
						<input
							v-model="merchant.title"
							class="input"
							type="text"
							:placeholder="$t('merchant.attributes.title')"
							required
						/>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.legalRepresentative') }}</label>
					<div class="control">
						<input
							v-model="merchant.legalRepresentative"
							class="input"
							type="text"
							:placeholder="$t('merchant.attributes.legalRepresentative')"
						/>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.businessAddress') }}</label>
					<div class="control">
						<textarea
							v-model="merchant.businessAddress"
							class="textarea"
							:placeholder="$t('merchant.attributes.businessAddress')"
						></textarea>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.businessDistrict') }}</label>
					<div class="control">
						<input
							v-model="merchant.businessDistrict"
							class="input"
							type="text"
							:placeholder="$t('merchant.attributes.businessDistrict')"
						/>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.validTime') }}</label>
					<div class="control">
						<input
							v-model="merchant.validTime"
							class="input"
							type="text"
							:placeholder="$t('merchant.attributes.validTime')"
						/>
						<p class="help">{{ $t('merchant.attributes.validTimeHelp') }}</p>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.trafficConditions') }}</label>
					<div class="control">
						<textarea
							v-model="merchant.trafficConditions"
							class="textarea"
							:placeholder="$t('merchant.attributes.trafficConditions')"
						></textarea>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.fixedEvents') }}</label>
					<div class="control">
						<textarea
							v-model="merchant.fixedEvents"
							class="textarea"
							:placeholder="$t('merchant.attributes.fixedEvents')"
						></textarea>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.terminalType') }}</label>
					<div class="control">
						<input
							v-model="merchant.terminalType"
							class="input"
							type="text"
							:placeholder="$t('merchant.attributes.terminalType')"
						/>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.specialTimePeriods') }}</label>
					<div class="control">
						<textarea
							v-model="merchant.specialTimePeriods"
							class="textarea"
							:placeholder="$t('merchant.attributes.specialTimePeriods')"
						></textarea>
					</div>
				</div>

				<div class="field">
					<label class="label">{{ $t('merchant.attributes.customFilters') }}</label>
					<div class="control">
						<textarea
							v-model="merchant.customFilters"
							class="textarea"
							:placeholder="$t('merchant.attributes.customFilters')"
						></textarea>
						<p class="help">{{ $t('merchant.attributes.customFiltersHelp') }}</p>
					</div>
				</div>

				<!-- Label Replacement Section -->
				<div class="field">
					<label class="label">{{ $t('merchant.labelReplacement.title') }}</label>
					<div class="control">
						<div class="label-replacement-container">
							<div v-for="(mapping, index) in labelMappings" :key="index" class="label-mapping">
								<div class="columns">
									<div class="column">
										<input
											v-model="mapping.placeholder"
											class="input"
											type="text"
											:placeholder="$t('merchant.labelReplacement.placeholder')"
										/>
									</div>
									<div class="column">
										<div class="select">
											<select v-model="mapping.labelId">
												<option value="">{{ $t('merchant.labelReplacement.selectLabel') }}</option>
												<option
													v-for="label in availableLabels"
													:key="label.id"
													:value="label.id"
												>
													{{ label.title }}
												</option>
											</select>
										</div>
									</div>
									<div class="column is-narrow">
										<XButton
											icon="trash"
											@click="removeLabelMapping(index)"
											variant="danger"
											size="small"
										>
											{{ $t('misc.remove') }}
										</XButton>
									</div>
								</div>
							</div>
							<XButton
								icon="plus"
								@click="addLabelMapping"
								variant="secondary"
								size="small"
							>
								{{ $t('merchant.labelReplacement.add') }}
							</XButton>
						</div>
						<p class="help">{{ $t('merchant.labelReplacement.help') }}</p>
					</div>
				</div>

				<div class="field">
					<div class="control">
						<XButton
							type="submit"
							:loading="loading"
							variant="primary"
						>
							{{ isEdit ? $t('misc.save') : $t('misc.create') }}
						</XButton>
					</div>
				</div>
			</form>
		</Card>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter, useRoute} from 'vue-router'

import MerchantService from '@/services/merchant'
import LabelService from '@/services/label'
import MerchantModel from '@/models/merchant'
import LabelModel from '@/models/label'
import {success, error} from '@/message'

import Card from '@/components/misc/Card.vue'
import XButton from '@/components/input/Button.vue'

interface LabelMapping {
	placeholder: string
	labelId: number | string
}

const {t} = useI18n()
const router = useRouter()
const route = useRoute()

const merchantService = new MerchantService()
const labelService = new LabelService()

const merchant = ref(new MerchantModel())
const availableLabels = ref<LabelModel[]>([])
const labelMappings = ref<LabelMapping[]>([])
const loading = ref(false)

const isEdit = computed(() => !!route.params.id)

onMounted(async () => {
	await loadLabels()
	
	if (isEdit.value) {
		await loadMerchant()
	}
})

async function loadLabels() {
	try {
		const response = await labelService.getAll()
		availableLabels.value = response.data
	} catch (e) {
		error(e)
	}
}

async function loadMerchant() {
	if (!route.params.id) return

	try {
		const merchantData = await merchantService.get(Number(route.params.id))
		merchant.value = merchantData

		// Parse existing label mappings from customFilters
		if (merchant.value.customFilters) {
			try {
				const parsedMappings = JSON.parse(merchant.value.customFilters)
				if (Array.isArray(parsedMappings)) {
					labelMappings.value = parsedMappings
				}
			} catch (e) {
				// If parsing fails, start with empty mappings
				labelMappings.value = []
			}
		}
	} catch (e) {
		error(e)
	}
}

async function saveMerchant() {
	loading.value = true
	
	try {
		// Save label mappings to customFilters
		merchant.value.customFilters = JSON.stringify(labelMappings.value)

		let savedMerchant
		if (isEdit.value) {
			savedMerchant = await merchantService.update(merchant.value)
		} else {
			savedMerchant = await merchantService.create(merchant.value)
		}

		success({
			message: isEdit.value 
				? t('merchant.edit.success') 
				: t('merchant.create.success')
		})

		router.push({name: 'merchants.index'})
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
}

function addLabelMapping() {
	labelMappings.value.push({
		placeholder: '',
		labelId: ''
	})
}

function removeLabelMapping(index: number) {
	labelMappings.value.splice(index, 1)
}
</script>

<style scoped lang="scss">
.merchant-edit {
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

.label-replacement-container {
	.label-mapping {
		margin-bottom: 1rem;
	}
}
</style>