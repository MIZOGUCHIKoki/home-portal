import { useEffect, useState } from "react"
import { fetchCategories, fetchMethods } from "../api/masterApi"
import type { ApiCategory, ApiMethod } from "../types/api"

export function useMasters() {
	const [categories, setCategories] = useState<ApiCategory[]>([])
	const [methods, setMethods] = useState<ApiMethod[]>([])
	const [loading, setLoading] = useState(true)
	const [error, setError] = useState<string>("")

	useEffect(() => {
		let mounted = true

		async function load() {
			try {
				setLoading(true)
				setError("")

				const [categoriesData, methodsData] = await Promise.all([
					fetchCategories(),
					fetchMethods(),
				])

				if (!mounted) return

				setCategories(categoriesData)
				setMethods(methodsData)
			} catch (err) {
				if (!mounted) return
				console.error(err)
				setError("マスタ取得に失敗しました")
			} finally {
				if (!mounted) return
				setLoading(false)
			}
		}

		load()

		return () => {
			mounted = false
		}
	}, [])

	return {
		categories,
		methods,
		loading,
		error,
	}
}