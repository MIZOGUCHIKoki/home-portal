import { useEffect, useState } from "react"
import {
	fetchTransactions,
	updateTransaction,
	updateAdvance,
	deleteTransaction,
} from "../api/transactionApi"
import {
	toTransaction,
	toUpdateTransactionRequest,
	toUpdateAdvanceRequest,
} from "../mappers/transactionMapper"
import type { Transaction } from "../types/view"
import { useMasters } from "../../masters/hooks/useMasters"

type Props = {
	reloadKey: number
}

const colWidths = [
	"160px",  // 日付
	"140px",  // 金額
	"150px",  // 実質金額
	"90px",   // 収支
	"90px",   // 振替
	"150px",  // カテゴリ
	"200px",  // 決済方法
	"500px",  // 場所
	"320px",  // メモ
	"360px",  // 立替一覧
	"120px",  // 立替不足
	"160px",  // 操作
]

export default function TransactionList({ reloadKey }: Props) {
	const { categories, methods } = useMasters()
	const [transactions, setTransactions] = useState<Transaction[]>([])
	const [loading, setLoading] = useState(true)
	const [error, setError] = useState("")

	const [editingId, setEditingId] = useState<number | null>(null)
	const [editRow, setEditRow] = useState<Transaction | null>(null)

	useEffect(() => {
		let mounted = true

		async function load() {
			try {
				setLoading(true)
				setError("")

				const data = await fetchTransactions()
				if (!mounted) return

				setTransactions(data.map(toTransaction))
			} catch (err) {
				if (!mounted) return
				console.error(err)
				setError("transaction一覧の取得に失敗しました")
			} finally {
				if (!mounted) return
				setLoading(false)
			}
		}

		load()

		return () => {
			mounted = false
		}
	}, [reloadKey])

	const categoryName = (id: number | null) => {
		if (!id) return "-"
		return categories.find((c) => c.id === id)?.name ?? "-"
	}

	const methodName = (id: number) => {
		return methods.find((m) => m.id === id)?.name ?? "-"
	}

	const startEdit = (t: Transaction) => {
		setEditingId(t.id)
		setEditRow({
			...t,
			advances: t.advances.map((a) => ({ ...a })),
		})
	}

	const cancelEdit = () => {
		setEditingId(null)
		setEditRow(null)
	}

	// ✅ 追加: 削除
	const startDelete = async (t: Transaction) => {
		const ok = window.confirm(
			`transaction ${t.id} を削除しますか？\n紐づく立替も削除されます。`
		)
		if (!ok) return

		try {
			await deleteTransaction(t.id)

			setTransactions((prev) => prev.filter((row) => row.id !== t.id))

			if (editingId === t.id) {
				setEditingId(null)
				setEditRow(null)
			}
		} catch (err) {
			console.error(err)
			alert("削除に失敗しました")
		}
	}

	const saveEdit = async () => {
		if (!editRow) return

		try {
			const transactionBody = toUpdateTransactionRequest(editRow)
			await updateTransaction(transactionBody)

			await Promise.all(
				editRow.advances.map((a) => updateAdvance(toUpdateAdvanceRequest(a)))
			)

			setTransactions((prev) =>
				prev.map((t) => (t.id === editRow.id ? editRow : t))
			)

			setEditingId(null)
			setEditRow(null)
		} catch (err) {
			console.error(err)
			alert("更新に失敗しました")
		}
	}

	const handleEditField = <K extends keyof Transaction>(
		key: K,
		value: Transaction[K]
	) => {
		if (!editRow) return
		setEditRow({
			...editRow,
			[key]: value,
		})
	}

	const handleEditAdvanceField = (
		advanceIndex: number,
		key: "name" | "amount",
		value: string | number
	) => {
		if (!editRow) return

		const nextAdvances = [...editRow.advances]
		const target = { ...nextAdvances[advanceIndex] }

		if (key === "name") {
			target.name = String(value)
		}
		if (key === "amount") {
			target.amount = Number(value)
		}

		nextAdvances[advanceIndex] = target

		setEditRow({
			...editRow,
			advances: nextAdvances,
		})
	}

	if (loading) {
		return <p>読み込み中...</p>
	}

	if (error) {
		return <p style={{ color: "red" }}>{error}</p>
	}

	return (
		<div>
			<h2>Transaction 一覧</h2>

			{transactions.length === 0 ? (
				<p>データがありません</p>
			) : (
				<div style={{ overflowX: "auto" }}>
					<table
						style={{
							width: "100%",
							minWidth: "2440px",
							borderCollapse: "collapse",
							tableLayout: "fixed",
						}}
					>
						<colgroup>
							{colWidths.map((width, index) => (
								<col key={index} style={{ width }} />
							))}
						</colgroup>

						<thead>
							<tr>
								<th style={thStyle}>日付</th>
								<th style={thStyle}>金額</th>
								<th style={thStyle}>実質金額</th>
								<th style={thStyle}>収支</th>
								<th style={thStyle}>振替</th>
								<th style={thStyle}>カテゴリ</th>
								<th style={thStyle}>決済方法</th>
								<th style={thStyle}>場所</th>
								<th style={thStyle}>メモ</th>
								<th style={thStyle}>立替一覧</th>
								<th style={thStyle}>立替不足</th>
								<th style={thStyle}>操作</th>
							</tr>
						</thead>

						<tbody>
							{transactions.map((t) => {
								const isEditing = editingId === t.id
								const row = isEditing && editRow ? editRow : t

								const advanceSum = row.advances.reduce((sum, a) => sum + a.amount, 0)
								const advanceDiff = row.amount - advanceSum - row.netAmount

								return (
									<tr key={t.id}>
										<td style={tdStyleNoWrap}>
											{isEditing ? (
												<input
													type="date"
													value={row.date}
													onChange={(e) => handleEditField("date", e.target.value)}
													style={inputStyle}
												/>
											) : (
												row.date
											)}
										</td>

										<td style={tdStyleNumber}>
											{isEditing ? (
												<input
													type="number"
													value={row.amount}
													onChange={(e) =>
														handleEditField("amount", Number(e.target.value))
													}
													style={inputStyle}
												/>
											) : (
												row.amount
											)}
										</td>

										<td style={tdStyleNumber}>
											{isEditing ? (
												<input
													type="number"
													value={row.netAmount}
													onChange={(e) =>
														handleEditField("netAmount", Number(e.target.value))
													}
													style={inputStyle}
												/>
											) : (
												row.netAmount
											)}
										</td>

										<td style={tdStyleNoWrap}>
											{isEditing ? (
												<select
													value={row.type ? "true" : "false"}
													onChange={(e) =>
														handleEditField("type", e.target.value === "true")
													}
													style={inputStyle}
												>
													<option value="false">支出</option>
													<option value="true">収入</option>
												</select>
											) : (
												row.type ? "収入" : "支出"
											)}
										</td>

										<td style={tdStyleNoWrap}>
											{isEditing ? (
												<input
													type="checkbox"
													checked={row.isTransfer}
													onChange={(e) =>
														handleEditField("isTransfer", e.target.checked)
													}
												/>
											) : (
												row.isTransfer ? "✅" : "☑️"
											)}
										</td>

										<td style={tdStyleNoWrap}>
											{isEditing ? (
												<select
													value={row.categoryId ?? ""}
													onChange={(e) =>
														handleEditField(
															"categoryId",
															e.target.value ? Number(e.target.value) : null
														)
													}
													style={inputStyle}
												>
													<option value="">未設定</option>
													{categories.map((c) => (
														<option key={c.id} value={c.id}>
															{c.name}
														</option>
													))}
												</select>
											) : (
												categoryName(row.categoryId)
											)}
										</td>

										<td style={tdStyleNoWrap}>
											{isEditing ? (
												<select
													value={row.methodId}
													onChange={(e) =>
														handleEditField("methodId", Number(e.target.value))
													}
													style={inputStyle}
												>
													{methods.map((m) => (
														<option key={m.id} value={m.id}>
															{m.name}
														</option>
													))}
												</select>
											) : (
												methodName(row.methodId)
											)}
										</td>

										<td style={tdStyleWrap}>
											{isEditing ? (
												<input
													value={row.place}
													onChange={(e) => handleEditField("place", e.target.value)}
													style={inputStyle}
												/>
											) : (
												row.place || "-"
											)}
										</td>

										<td style={tdStyleWrap}>
											{isEditing ? (
												<input
													value={row.note}
													onChange={(e) => handleEditField("note", e.target.value)}
													style={inputStyle}
												/>
											) : (
												row.note || "-"
											)}
										</td>

										<td style={tdStyle}>
											{row.advances.length === 0 ? (
												<span>なし</span>
											) : (
												<div style={{ minWidth: 260 }}>
													{row.advances.map((a, idx) => (
														<div
															key={a.id}
															style={{
																padding: "6px 0",
																borderBottom: "1px solid #eee",
																whiteSpace: "nowrap",
															}}
														>
															<div>立替ID: {a.id}</div>

															{isEditing ? (
																<>
																	<div>
																		<input
																			value={a.name}
																			onChange={(e) =>
																				handleEditAdvanceField(idx, "name", e.target.value)
																			}
																			style={inputStyle}
																		/>
																	</div>
																	<div>
																		<input
																			type="number"
																			value={a.amount}
																			onChange={(e) =>
																				handleEditAdvanceField(
																					idx,
																					"amount",
																					Number(e.target.value)
																				)
																			}
																			style={inputStyle}
																		/>
																	</div>
																	<div>返済済み: {a.returnedAmount}</div>
																	<div>状態: {a.status ? "完済" : "未完済"}</div>
																</>
															) : (
																<>
																	<div><strong>{a.name}</strong></div>
																	<div>金額: {a.amount}</div>
																	<div>返済済み: {a.returnedAmount}</div>
																	<div>状態: {a.status ? "完済" : "未完済"}</div>
																</>
															)}
														</div>
													))}
												</div>
											)}

											{row.advances.length === 0 ? null : (
												<div style={{ textAlign: "right", marginTop: 8, whiteSpace: "nowrap" }}>
													合計: {advanceSum}
												</div>
											)}
										</td>

										<td style={tdStyleNumber}>
											{row.advances.length === 0 ? (
												<span>0</span>
											) : (
												<span>{advanceDiff}</span>
											)}
										</td>

										<td style={tdStyleNoWrap}>
											{isEditing ? (
												<>
													<div style={{ marginBottom: 8 }}>
														<button onClick={saveEdit}>DONE</button>
													</div>
													<div>
														<button onClick={cancelEdit}>CANCEL</button>
													</div>
												</>
											) : (
												<>
													<div style={{ marginBottom: 8 }}>
														<button onClick={() => startEdit(t)}>EDIT</button>
													</div>
													<div>
														<button onClick={() => startDelete(t)}>DELETE</button>
													</div>
												</>
											)}
										</td>
									</tr>
								)
							})}
						</tbody>
					</table>
				</div>
			)}
		</div>
	)
}

const thBase: React.CSSProperties = {
	border: "1px solid #ccc",
	padding: "8px",
	verticalAlign: "top",
	fontWeight: 600,
}

const tdBase: React.CSSProperties = {
	border: "1px solid #ccc",
	padding: "8px",
	verticalAlign: "top",
}

const thStyle: React.CSSProperties = {
	...thBase,
	textAlign: "left",
	whiteSpace: "nowrap",
}

const tdStyle: React.CSSProperties = {
	...tdBase,
	whiteSpace: "nowrap",
}

const tdStyleNumber: React.CSSProperties = {
	...tdBase,
	textAlign: "right",
	whiteSpace: "nowrap",
}

const tdStyleNoWrap: React.CSSProperties = {
	...tdBase,
	whiteSpace: "nowrap",
}

const inputStyle: React.CSSProperties = {
	width: "100%",
	boxSizing: "border-box",
}

const tdStyleWrap: React.CSSProperties = {
	...tdBase,
	whiteSpace: "normal",
	wordBreak: "break-word",
	lineHeight: 1.4,
}