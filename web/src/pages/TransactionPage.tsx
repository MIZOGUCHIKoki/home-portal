import { useState } from "react"
import TransactionForm from "../features/transactions/components/TransactionForm"
import TransactionList from "../features/transactions/components/TransactionList"

export default function TransactionsPage() {
	const [reloadKey, setReloadKey] = useState(0)

	const handleCreated = () => {
		setReloadKey((prev) => prev + 1)
	}

	return (
		<div>
			<div>
				<h1>家計簿</h1>
			</div>

			<div>
				<div>
					<TransactionForm onCreated={handleCreated} />
				</div>

				<div>
					<TransactionList reloadKey={reloadKey} />
				</div>
			</div>
		</div>
	)
}