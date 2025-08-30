import AbstractService from './abstractService'
import MerchantModel from '@/models/merchant'
import type {IMerchant} from '@/modelTypes/IMerchant'

export default class MerchantService extends AbstractService<IMerchant> {
	constructor() {
		super({
			create: '/merchants',
			getAll: '/merchants',
			get: '/merchants/{id}',
			update: '/merchants/{id}',
			delete: '/merchants/{id}',
		})
	}

	modelFactory(data) {
		return new MerchantModel(data)
	}

	beforeUpdate(merchant) {
		return merchant
	}

	beforeCreate(merchant) {
		return merchant
	}

	/**
	 * Bulk delete merchants by ids
	 */
	async bulkDelete(ids: number[]) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post('/merchants/bulk_delete', { ids })
			return response.data
		} finally {
			cancel()
		}
	}
}
