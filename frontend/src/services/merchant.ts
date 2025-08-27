import AbstractService from './abstractService'
import MerchantModel from '@/models/merchant'
import type {IMerchant} from '@/modelTypes/IMerchant'

export default class MerchantService extends AbstractService<IMerchant> {
	constructor() {
		super({
			getAll: '/merchants',
			create: '/merchants',
			get: '/merchants/{id}',
			update: '/merchants/{id}',
			delete: '/merchants/{id}',
		})
	}

	modelFactory(data: Partial<IMerchant>) {
		return new MerchantModel(data)
	}

	async importFromXlsx(file: File): Promise<IMerchant[]> {
		const formData = new FormData()
		formData.append('file', file)

		const response = await this.http.put('/merchants/import', formData, {
			headers: {
				'Content-Type': 'multipart/form-data',
			},
		})

		return response.data.map((merchant: IMerchant) => this.modelFactory(merchant))
	}
}