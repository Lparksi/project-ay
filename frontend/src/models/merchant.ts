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

	owner: IUser = UserModel
	created: Date = null
	updated: Date = null

	constructor(data: Partial<IMerchant> = {}) {
		super()
		this.assignData(data)

		this.owner = new UserModel(this.owner)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}