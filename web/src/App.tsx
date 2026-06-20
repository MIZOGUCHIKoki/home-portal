import { useState, useEffect } from "react"

type Category = {
  id: number
  name: string
}

type Method = {
  id: number
  name: string
}

type AdvanceResponse = {
  id: number
  name: string
  amount: number
  returned_amount: number
  status: boolean
}

type TransactionResponse = {
  id: number
  user_id: number
  date: string
  amount: number
  net_amount: number
  type: boolean
  is_transfer: boolean
  method_id: number
  category_id: number | null
  place?: string
  note?: string
  advances?: AdvanceResponse[] | null
}

function App() {
  const [amount, setAmount] = useState("")
  const [netAmount, setNetAmount] = useState("")
  const [place, setPlace] = useState("")
  const [note, setNote] = useState("")
  const [date, setDate] = useState(() =>
    new Date().toISOString().slice(0, 10)
  )

  const [categoryId, setCategoryId] = useState("")
  const [methodId, setMethodId] = useState("")

  const [type, setType] = useState(false)
  const [isTransfer, setIsTransfer] = useState(false)

  const [mode, setMode] = useState<"normal" | "advance" | "refund">("normal")

  const [advances, setAdvances] = useState([{ name: "", amount: "" }])
  const [refundAdvanceId, setRefundAdvanceId] = useState("")

  const [categories, setCategories] = useState<Category[]>([])
  const [methods, setMethods] = useState<Method[]>([])
  const [transactions, setTransactions] = useState<TransactionResponse[]>([])

  const loadMasters = async () => {
    try {
      const categoriesRes = await fetch("http://127.0.0.1:8080/categories")
      const categoriesData = await categoriesRes.json()
      setCategories(Array.isArray(categoriesData) ? categoriesData : [])

      const methodsRes = await fetch("http://127.0.0.1:8080/methods")
      const methodsData = await methodsRes.json()
      setMethods(Array.isArray(methodsData) ? methodsData : [])
    } catch (err) {
      console.error("loadMasters error:", err)
      setCategories([])
      setMethods([])
    }
  }

  const loadTransactions = async () => {
    try {
      const res = await fetch("http://127.0.0.1:8080/transactions")
      const data = await res.json()
      setTransactions(Array.isArray(data) ? data : [])
    } catch (err) {
      console.error("loadTransactions error:", err)
      setTransactions([])
    }
  }

  useEffect(() => {
    loadMasters()
    loadTransactions()
  }, [])

  const resetForm = () => {
    setAmount("")
    setNetAmount("")
    setPlace("")
    setNote("")
    setCategoryId("")
    setMethodId("")
    setType(false)
    setIsTransfer(false)
    setMode("normal")
    setAdvances([{ name: "", amount: "" }])
    setRefundAdvanceId("")
  }

  const handleSubmit = async () => {
    if (!methodId) {
      alert("決済方法を選択してください")
      return
    }

    const body = {
      user_id: 1,
      date,

      amount: Number(amount),
      net_amount:
        mode === "refund" ? Number(amount) : Number(netAmount || amount),

      type: mode === "refund" ? true : type,
      is_transfer: isTransfer,

      place: mode === "refund" ? "" : place,
      note,

      category_id: mode === "refund"
        ? null
        : (categoryId ? Number(categoryId) : null),

      method_id: Number(methodId),

      advances:
        mode === "advance"
          ? advances
            .filter((a) => a.name && a.amount)
            .map((a) => ({
              name: a.name,
              amount: Number(a.amount),
            }))
          : [],

      refund_advance_id:
        mode === "refund" && refundAdvanceId
          ? Number(refundAdvanceId)
          : null,
    }

    console.log("送信データ", body)

    try {
      const res = await fetch("http://127.0.0.1:8080/transactions", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      })

      const data = await res.json()
      console.log("レスポンス", data)

      if (res.ok) {
        alert("登録成功 🎉")
        resetForm()
        loadTransactions()
      } else {
        alert("エラー")
      }
    } catch (err) {
      console.error(err)
      alert("通信エラー")
    }
  }

  const categoryName = (id: number | null) => {
    if (!id) return "-"
    return categories.find((c) => c.id === id)?.name ?? "-"
  }

  const methodName = (id: number) => {
    return methods.find((m) => m.id === id)?.name ?? "-"
  }

  return (
    <div style={{ padding: 20 }}>
      <h2>Transaction 登録</h2>

      <div style={{ marginBottom: 12 }}>
        <label>
          <input
            type="radio"
            checked={mode === "normal"}
            onChange={() => setMode("normal")}
          />
          通常
        </label>
        <label style={{ marginLeft: 12 }}>
          <input
            type="radio"
            checked={mode === "advance"}
            onChange={() => setMode("advance")}
          />
          立替
        </label>
        <label style={{ marginLeft: 12 }}>
          <input
            type="radio"
            checked={mode === "refund"}
            onChange={() => setMode("refund")}
          />
          返済
        </label>
      </div>

      <div>
        <label>日付 </label>
        <input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
      </div>

      <div>
        <input
          placeholder="金額"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
        />
      </div>

      {mode !== "refund" && (
        <div>
          <input
            placeholder="実質金額"
            value={netAmount}
            onChange={(e) => setNetAmount(e.target.value)}
          />
        </div>
      )}

      {mode !== "refund" && (
        <div>
          <input
            placeholder="場所"
            value={place}
            onChange={(e) => setPlace(e.target.value)}
          />
        </div>
      )}

      <div>
        <input
          placeholder="メモ"
          value={note}
          onChange={(e) => setNote(e.target.value)}
        />
      </div>

      {mode !== "refund" && (
        <div>
          <select value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
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
        <select value={methodId} onChange={(e) => setMethodId(e.target.value)}>
          <option value="">決済方法選択</option>
          {methods.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name}
            </option>
          ))}
        </select>
      </div>

      {mode !== "refund" && (
        <div>
          <select
            value={type ? "true" : "false"}
            onChange={(e) => setType(e.target.value === "true")}
          >
            <option value="false">支出</option>
            <option value="true">収入</option>
          </select>
        </div>
      )}

      <div>
        <label>
          <input
            type="checkbox"
            checked={isTransfer}
            onChange={(e) => setIsTransfer(e.target.checked)}
          />
          振替
        </label>
      </div>

      {mode === "advance" && (
        <div style={{ marginTop: 12 }}>
          <h3>立替入力</h3>
          {advances.map((a, i) => (
            <div key={i}>
              <input
                placeholder="名前"
                value={a.name}
                onChange={(e) => {
                  const copy = [...advances]
                  copy[i].name = e.target.value
                  setAdvances(copy)
                }}
              />
              <input
                placeholder="金額"
                value={a.amount}
                onChange={(e) => {
                  const copy = [...advances]
                  copy[i].amount = e.target.value
                  setAdvances(copy)
                }}
              />
            </div>
          ))}
          <button
            onClick={() =>
              setAdvances([...advances, { name: "", amount: "" }])
            }
          >
            + 立替追加
          </button>
        </div>
      )}

      {mode === "refund" && (
        <div style={{ marginTop: 12 }}>
          <h3>返済入力</h3>
          <input
            placeholder="返済対象 advance_id"
            value={refundAdvanceId}
            onChange={(e) => setRefundAdvanceId(e.target.value)}
          />
        </div>
      )}

      <div style={{ marginTop: 12 }}>
        <button onClick={handleSubmit}>登録</button>
      </div>

      <hr style={{ margin: "24px 0" }} />

      <h2>Transaction 一覧</h2>

      {(transactions ?? []).map((t) => (
        <div
          key={t.id}
          style={{
            border: "1px solid #ccc",
            padding: 12,
            marginBottom: 12,
            borderRadius: 8,
          }}
        >
          <div><strong>ID:</strong> {t.id}</div>
          <div><strong>日付:</strong> {t.date}</div>
          <div><strong>金額:</strong> {t.amount}</div>
          <div><strong>実質金額:</strong> {t.net_amount}</div>
          <div><strong>収支:</strong> {t.type ? "収入" : "支出"}</div>
          <div><strong>振替:</strong> {t.is_transfer ? "あり" : "なし"}</div>
          <div><strong>カテゴリ:</strong> {categoryName(t.category_id)}</div>
          <div><strong>決済方法:</strong> {methodName(t.method_id)}</div>
          <div><strong>場所:</strong> {t.place || "-"}</div>
          <div><strong>メモ:</strong> {t.note || "-"}</div>

          <div style={{ marginTop: 8 }}>
            <strong>Advance 一覧</strong>
            {(t.advances ?? []).length === 0 ? (
              <div>なし</div>
            ) : (
              <ul>
                {(t.advances ?? []).map((a) => (
                  <li key={a.id}>
                    {a.name} / 金額: {a.amount} / 返済済み: {a.returned_amount} / 状態: {a.status ? "完済" : "未完済"}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export default App