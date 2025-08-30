import AbstractService from './abstractService'
import type { IMerchantMapping, IFieldLabelMapping } from '@/modelTypes/IMerchant'

export default class MerchantMappingService extends AbstractService<IMerchantMapping> {
	constructor() {
		super({
			create: '/merchant-mappings',
			getAll: '/merchant-mappings',
			get: '/merchant-mappings/{id}',
			update: '/merchant-mappings/{id}',
			delete: '/merchant-mappings/{id}',
		})
	}

	modelFactory(data: Partial<IMerchantMapping>): IMerchantMapping {
		return {
			id: 0,
			fieldName: '',
			placeholder: '',
			displayText: '',
			labelId: 0,
			isActive: true,
			created: new Date(),
			updated: new Date(),
			maxPermission: 2, // 添加缺失的属性
			...data,
		} as IMerchantMapping
	}

	beforeUpdate(mapping: IMerchantMapping): IMerchantMapping {
		return mapping
	}

	beforeCreate(mapping: IMerchantMapping): IMerchantMapping {
		return mapping
	}

	/**
	 * Bulk save field mappings
	 */
	async bulkSaveMappings(fieldMappings: IFieldLabelMapping[]) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post('/merchant-mappings/bulk_save', { 
				fieldMappings 
			})
			return response.data
		} finally {
			cancel()
		}
	}

	/**
	 * Bulk delete mappings by field names
	 */
	async bulkDeleteByFields(fields: string[]) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.delete('/merchant-mappings/bulk_delete', { 
				data: { fields } 
			})
			return response.data
		} finally {
			cancel()
		}
	}

	/**
	 * Load all mappings and convert to field mapping format
	 */
	async loadAllMappings(): Promise<IFieldLabelMapping[]> {
		const cancel = this.setLoading()
		try {
			const mappings = await this.getAll(this.modelFactory({}), {}, -1) as IMerchantMapping[]
			
			// Group by field name
			const fieldMap: Record<string, IFieldLabelMapping> = {}
			
			mappings.forEach(mapping => {
				if (!fieldMap[mapping.fieldName]) {
					fieldMap[mapping.fieldName] = {
						field: mapping.fieldName,
						mappings: []
					}
				}
				
				fieldMap[mapping.fieldName].mappings.push({
					placeholder: mapping.placeholder,
					displayText: mapping.displayText,
					labelId: mapping.labelId
				})
			})
			
			return Object.values(fieldMap)
		} finally {
			cancel()
		}
	}
}
