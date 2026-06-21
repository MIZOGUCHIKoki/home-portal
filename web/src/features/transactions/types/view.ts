export type Advance = {
	id: number
	name: string
	amount: number
	returnedAmount: number
	status: boolean
}

export type Transaction = {
	id: number
	userId: number
	date: string

	amount: number
	netAmount: number

	type: boolean
	isTransfer: boolean

	methodId: number
	categoryId: number | null

	place: string
	note: string

	advances: Advance[]
}

export type TransactionMode = "normal" | "advance" | "refund"

export type AdvanceFormRow = {
	name: string
	amount: string
}

export type TransactionFormValues = {
	date: string
	amount: string
	netAmount: string
	place: string
	note: string

	categoryId: string
	methodId: string

	type: boolean
	isTransfer: boolean

	mode: TransactionMode

	advances: AdvanceFormRow[]
	refundAdvanceId: string
}