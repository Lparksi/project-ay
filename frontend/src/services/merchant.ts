import AbstractService from './abstractService'
import MerchantModel from '@/models/merchant'
import type {IMerchant} from '@/modelTypes/IMerchant'
import axios from 'axios'
import {getToken} from '@/helpers/auth'

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

	async importFromXlsx(file: File, headerMapping: Record<string, string> = {}): Promise<IMerchant[]> {
		const formData = new FormData()
		formData.append('file', file)

		// Convert header mapping to JSON and append it to the form data
		formData.append('headerMapping', JSON.stringify(headerMapping))

		// Use the parent's uploadFormData method which properly handles multipart/form-data
		const response = await this.uploadFormData('/merchants/import', formData)

		// Backend responds with { message, count, merchants }
		const data = response && response.merchants ? response.merchants : response
		return (Array.isArray(data) ? data : []).map((merchant: IMerchant) => this.modelFactory(merchant))
	}
}