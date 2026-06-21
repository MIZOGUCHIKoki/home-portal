import type {
	ApiCreateTransactionRequest,
	ApiTransactionResponse,
	ApiUpdateTransactionRequest,
	ApiUpdateAdvanceRequest,
} from "../types/api"

const BASE_URL = "http://127.0.0.1:8080"

async function safeReadResponse(res: Response) {
	const text = await res.text()

	try {
		return text ? JSON.parse(text) : {}
	} catch {
		return { error: text || `HTTP ${res.status}` }
	}
}

export async function fetchTransactions(): Promise<ApiTransactionResponse[]> {
	const res = await fetch(`${BASE_URL}/transactions`)
	const data = await safeReadResponse(res)

	if (!res.ok) {
		throw new Error(data?.error ?? "transaction一覧の取得に失敗しました")
	}

	return Array.isArray(data) ? data : []
}

export async function createTransaction(
	body: ApiCreateTransactionRequest
): Promise<{ status: string }> {
	const res = await fetch(`${BASE_URL}/transactions`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(body),
	})

	const data = await safeReadResponse(res)

	if (!res.ok) {
		throw new Error(data?.error ?? "transaction登録に失敗しました")
	}

	return data
}

export async function updateTransaction(
	body: ApiUpdateTransactionRequest
): Promise<{ status: string }> {
	const res = await fetch(`${BASE_URL}/transactions`, {
		method: "PUT",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(body),
	})

	const data = await safeReadResponse(res)

	if (!res.ok) {
		throw new Error(data?.error ?? "transaction更新に失敗しました")
	}

	return data
}

export async function updateAdvance(
	body: ApiUpdateAdvanceRequest
): Promise<{ status: string }> {
	const res = await fetch(`${BASE_URL}/advances`, {
		method: "PUT",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(body),
	})

	const data = await safeReadResponse(res)

	if (!res.ok) {
		throw new Error(data?.error ?? "advance更新に失敗しました")
	}

	return data
}

export async function deleteTransaction(id: number): Promise<{ status: string }> {
	const res = await fetch(`${BASE_URL}/transactions?id=${id}`, {
		method: "DELETE",
	})

	const data = await safeReadResponse(res)

	if (!res.ok) {
		throw new Error(data?.error ?? "transaction削除に失敗しました")
	}

	return data
}