import { useMemo, useState } from "react"
import { useMasters } from "../../masters/hooks/useMasters"
import { createTransaction } from "../api/transactionApi"
import { toCreateTransactionRequest } from "../mappers/transactionMapper"
import type { TransactionFormValues } from "../types/view"

type Props = {
	onCreated: () => void
}

function createInitialValues(): TransactionFormValues {
	return {
		date: new Date().toISOString().slice(0, 10),
		amount: "",
		netAmount: "",
		place: "",
		note: "",

		categoryId: "",
		methodId: "",

		type: false,
		isTransfer: false,

		mode: "normal",

		advances: [{ name: "", amount: "" }],
		refundAdvanceId: "",
	}
}

export default function TransactionForm({ onCreated }: Props) {
	const { categories, methods, loading, error } = useMasters()

	const [values, setValues] = useState<TransactionFormValues>(createInitialValues())
	const [submitting, setSubmitting] = useState(false)

	const canSubmit = useMemo(() => {
		if (!values.methodId) return false
		if (!values.amount) return false

		if (values.mode === "refund" && !values.refundAdvanceId) return false

		return true
	}, [values])

	const updateField = <K extends keyof TransactionFormValues>(
		key: K,
		value: TransactionFormValues[K]
	) => {
		setValues((prev) => ({
			...prev,
			[key]: value,
		}))
	}

	const updateAdvanceRow = (
		index: number,
		key: "name" | "amount",
		value: string
	) => {
		setValues((prev) => {
			const next = [...prev.advances]
			next[index] = {
				...next[index],
				[key]: value,
			}
			return {
				...prev,
				advances: next,
			}
		})
	}

	const addAdvanceRow = () => {
		setValues((prev) => ({
			...prev,
			advances: [...prev.advances, { name: "", amount: "" }],
		}))
	}

	const resetForm = () => {
		setValues(createInitialValues())
	}

	const handleSubmit = async () => {
		if (!canSubmit) {
			alert("必須項目を入力してください")
			return
		}

		try {
			setSubmitting(true)

			const body = toCreateTransactionRequest(values)
			console.log("送信データ", body)

			await createTransaction(body)

			alert("登録成功 🎉")
			resetForm()
			onCreated()
		} catch (err) {
			console.error(err)
			alert("登録に失敗しました")
		} finally {
			setSubmitting(false)
		}
	}

	return (
		<div>
			<h2>Transaction 登録</h2>

			{loading && <p>マスタ読込中...</p>}
			{error && <p>{error}</p>}

			<div>
				<label>
					<input
						type="radio"
						checked={values.mode === "normal"}
						onChange={() => updateField("mode", "normal")}
					/>
					通常
				</label>

				<label style={{ marginLeft: 12 }}>
					<input
						type="radio"
						checked={values.mode === "advance"}
						onChange={() => updateField("mode", "advance")}
					/>
					立替
				</label>

				<label style={{ marginLeft: 12 }}>
					<input
						type="radio"
						checked={values.mode === "refund"}
						onChange={() => updateField("mode", "refund")}
					/>
					立替精算
				</label>
			</div>

			<div>
				<label>日付 </label>
				<input
					type="date"
					value={values.date}
					onChange={(e) => updateField("date", e.target.value)}
				/>
			</div>

			<div>
				<input
					placeholder="金額"
					value={values.amount}
					onChange={(e) => updateField("amount", e.target.value)}
				/>
			</div>

			{values.mode !== "refund" && (
				<div>
					<input
						placeholder="実質金額"
						value={values.netAmount}
						onChange={(e) => updateField("netAmount", e.target.value)}
					/>
				</div>
			)}

			{values.mode !== "refund" && (
				<div>
					<input
						placeholder="場所"
						value={values.place}
						onChange={(e) => updateField("place", e.target.value)}
					/>
				</div>
			)}

			<div>
				<input
					placeholder="メモ"
					value={values.note}
					onChange={(e) => updateField("note", e.target.value)}
				/>
			</div>

			{values.mode !== "refund" && (
				<div>
					<select
						value={values.categoryId}
						onChange={(e) => updateField("categoryId", e.target.value)}
					>
						<option value="">カテゴリ選択</option>
						{categories.map((c) => (
							<option key={c.id} value={c.id}>
								{c.name}
							</option>
						))}
					</select>
				</div>
			)}

			<div>
				<select
					value={values.methodId}
					onChange={(e) => updateField("methodId", e.target.value)}
				>
					<option value="">決済方法選択</option>
					{methods.map((m) => (
						<option key={m.id} value={m.id}>
							{m.name}
						</option>
					))}
				</select>
			</div>

			{values.mode === "normal" && (
				<>
					<div>
						<select
							value={values.type ? "true" : "false"}
							onChange={(e) => updateField("type", e.target.value === "true")}
						>
							<option value="false">支出</option>
							<option value="true">収入</option>
						</select>
					</div>

					<div>
						<label>
							<input
								type="checkbox"
								checked={values.isTransfer}
								onChange={(e) => updateField("isTransfer", e.target.checked)}
							/>
							振替
						</label>
					</div>
				</>
			)}

			{values.mode === "advance" && (
				<div style={{ marginTop: 12 }}>
					<h3>立替入力</h3>

					{values.advances.map((a, i) => (
						<div key={i} style={{ display: "flex", gap: 8, marginBottom: 8 }}>
							<input
								placeholder="名前"
								value={a.name}
								onChange={(e) => updateAdvanceRow(i, "name", e.target.value)}
							/>
							<input
								placeholder="金額"
								value={a.amount}
								onChange={(e) => updateAdvanceRow(i, "amount", e.target.value)}
							/>
						</div>
					))}

					<button onClick={addAdvanceRow}>
						+ 立替追加
					</button>
				</div>
			)}

			{values.mode === "refund" && (
				<div style={{ marginTop: 12 }}>
					<h3>返済入力</h3>
					<input
						placeholder="返済対象 advance_id"
						value={values.refundAdvanceId}
						onChange={(e) => updateField("refundAdvanceId", e.target.value)}
					/>
				</div>
			)}

			<div style={{ marginTop: 16 }}>
				<button onClick={handleSubmit} disabled={!canSubmit || submitting}>
					{submitting ? "送信中..." : "登録"}
				</button>
			</div>
		</div>
	)
}