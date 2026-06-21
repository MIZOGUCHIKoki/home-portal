export type ApiAdvanceInput = {
	name: string
	amount: number
}

export type ApiCreateTransactionRequest = {
	user_id: number
	date: string

	amount: number
	net_amount: number

	type: boolean
	is_transfer: boolean

	place: string
	note: string

	method_id: number
	category_id: number | null

	advances: ApiAdvanceInput[]

	refund_advance_id: number | null
}

export type ApiUpdateTransactionRequest = {
	id: number
	date: string

	amount: number
	net_amount: number

	type: boolean
	is_transfer: boolean

	place: string
	note: string

	method_id: number
	category_id: number | null
}

// ✅ 追加
export type ApiUpdateAdvanceRequest = {
	id: number
	name: string
	amount: number
}

export type ApiAdvanceResponse = {
	id: number
	name: string
	amount: number
	returned_amount: number
	status: boolean
}

export type ApiTransactionResponse = {
	id: number
	user_id: number
	date: string

	amount: number
	net_amount: number

	type: boolean
	is_transfer: boolean

	method_id: number
	category_id: number | null

	place: string
	note: string

	advances: ApiAdvanceResponse[] | null
}