import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

/**
 * Merchant interface for business management
 */
export interface IMerchant extends IAbstract {
	id: number
	title: string
	legalRepresentative: string
	businessAddress: string
	businessDistrict: string
	validTime: string
	trafficConditions: string
	fixedEvents: string
	terminalType: string
	specialTimePeriods: string
	customFilters: string

	owner: IUser
	created: Date
	updated: Date
}