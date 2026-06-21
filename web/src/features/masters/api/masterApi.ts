import type { ApiCategory, ApiMethod } from "../types/api"

const BASE_URL = "http://127.0.0.1:8080"

export async function fetchCategories(): Promise<ApiCategory[]> {
	const res = await fetch(`${BASE_URL}/categories`)
	if (!res.ok) {
		throw new Error("カテゴリ取得に失敗しました")
	}
	const data = await res.json()
	return Array.isArray(data) ? data : []
}

export async function fetchMethods(): Promise<ApiMethod[]> {
	const res = await fetch(`${BASE_URL}/methods`)
	if (!res.ok) {
		throw new Error("決済方法取得に失敗しました")
	}
	const data = await res.json()
	return Array.isArray(data) ? data : []
}