import AbstractModel from './abstractModel'
import UserModel from '@/models/user'

import type {IMerchant} from '@/modelTypes/IMerchant'
import type {IUser} from '@/modelTypes/IUser'

export default class MerchantModel extends AbstractModel<IMerchant> implements IMerchant {
	id = 0
	title = ''
	legalRepresentative = ''
	businessAddress = ''
	businessDistrict = ''
	validTime = ''
	trafficConditions = ''
	fixedEvents = ''
	terminalType = ''
	specialTimePeriods = ''
	customFilters = ''

	owner: IUser = new UserModel()
	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<IMerchant> = {}) {
		super()
		this.assignData(data)

		this.owner = new UserModel(this.owner)

		if (this.created) {
			this.created = new Date(this.created)
		}
		if (this.updated) {
			this.updated = new Date(this.updated)
		}
	}
}