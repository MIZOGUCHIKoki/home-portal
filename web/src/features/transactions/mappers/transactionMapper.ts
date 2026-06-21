import type {
	ApiAdvanceResponse,
	ApiCreateTransactionRequest,
	ApiTransactionResponse,
	ApiUpdateTransactionRequest,
	ApiUpdateAdvanceRequest,
} from "../types/api"
import type {
	Advance,
	Transaction,
	TransactionFormValues,
} from "../types/view"

export function toAdvance(api: ApiAdvanceResponse): Advance {
	return {
		id: api.id,
		name: api.name,
		amount: api.amount,
		returnedAmount: api.returned_amount,
		status: api.status,
	}
}

export function toTransaction(api: ApiTransactionResponse): Transaction {
	return {
		id: api.id,
		userId: api.user_id,
		date: api.date,
		amount: api.amount,
		netAmount: api.net_amount,
		type: api.type,
		isTransfer: api.is_transfer,
		methodId: api.method_id,
		categoryId: api.category_id,
		place: api.place,
		note: api.note,
		advances: (api.advances ?? []).map(toAdvance),
	}
}

export function toCreateTransactionRequest(
	values: TransactionFormValues
): ApiCreateTransactionRequest {
	const amount = Number(values.amount)
	const netAmount =
		values.mode === "refund"
			? amount
			: Number(values.netAmount || values.amount)

	return {
		user_id: 1,
		date: values.date,

		amount,
		net_amount: netAmount,

		type: values.mode === "refund" ? true : values.type,
		is_transfer: values.mode === "normal" ? values.isTransfer : false,

		place: values.mode === "refund" ? "" : values.place,
		note: values.note,

		method_id: Number(values.methodId),
		category_id:
			values.mode === "refund"
				? null
				: values.categoryId
					? Number(values.categoryId)
					: null,

		advances:
			values.mode === "advance"
				? values.advances
					.filter((a) => a.name && a.amount)
					.map((a) => ({
						name: a.name,
						amount: Number(a.amount),
					}))
				: [],

		refund_advance_id:
			values.mode === "refund" && values.refundAdvanceId
				? Number(values.refundAdvanceId)
				: null,
	}
}

export function toUpdateTransactionRequest(t: Transaction): ApiUpdateTransactionRequest {
	return {
		id: t.id,
		date: t.date,
		amount: t.amount,
		net_amount: t.netAmount,
		type: t.type,
		is_transfer: t.isTransfer,
		place: t.place,
		note: t.note,
		method_id: t.methodId,
		category_id: t.categoryId,
	}
}

export function toUpdateAdvanceRequest(a: Advance): ApiUpdateAdvanceRequest {
	return {
		id: a.id,
		name: a.name,
		amount: a.amount,
	}
}