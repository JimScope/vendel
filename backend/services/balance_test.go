package services_test

import (
	"testing"

	"vendel/services"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// TestProcessPaymentCreditIdempotency verifies that redelivering the same
// provider transaction id credits the balance exactly once. The idempotency
// marker in payment_idempotency is what makes the second delivery a no-op.
func TestProcessPaymentCreditIdempotency(t *testing.T) {
	app := setupServicesTestApp(t)
	defer app.Cleanup()

	userID := seededUserID(t, app)
	const txID = "tx-idem-001"
	const amount = 25.0

	// First delivery credits the balance.
	res1, err := services.ProcessPaymentCredit(app, userID, txID, amount)
	if err != nil {
		t.Fatalf("first credit failed: %v", err)
	}
	if res1["status"] != "balance_credited" {
		t.Fatalf("first delivery status = %v, want balance_credited", res1["status"])
	}
	if got := res1["new_balance"]; got != amount {
		t.Fatalf("first new_balance = %v, want %v", got, amount)
	}

	// Redelivery of the same tx id must be a no-op.
	res2, err := services.ProcessPaymentCredit(app, userID, txID, amount)
	if err != nil {
		t.Fatalf("redelivery failed: %v", err)
	}
	if res2["status"] != "already_processed" {
		t.Fatalf("redelivery status = %v, want already_processed", res2["status"])
	}

	// Balance must reflect a single credit, not two.
	bal, err := services.GetBalance(app, userID)
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}
	if bal != amount {
		t.Fatalf("balance after redelivery = %v, want %v (double-credit detected)", bal, amount)
	}

	// Exactly one idempotency marker should exist for this tx.
	assertMarkerCount(t, app, txID, 1)
}

// TestProcessDepositIdempotency verifies the same guard for wallet deposits
// routed through ProcessDeposit (wallet-address lookup + creditAndActivate).
func TestProcessDepositIdempotency(t *testing.T) {
	app := setupServicesTestApp(t)
	defer app.Cleanup()

	userID := seededUserID(t, app)
	const walletAddr = "0xDEPOSITWALLET"
	const txHash = "0xhash-idem-001"
	const amount = 40.0

	if err := services.SetWalletInfo(app, userID, walletAddr, "wallet-123"); err != nil {
		t.Fatalf("SetWalletInfo failed: %v", err)
	}

	res1, err := services.ProcessDeposit(app, walletAddr, txHash, amount, "USDT")
	if err != nil {
		t.Fatalf("first deposit failed: %v", err)
	}
	if res1["status"] != "balance_credited" {
		t.Fatalf("first deposit status = %v, want balance_credited", res1["status"])
	}

	res2, err := services.ProcessDeposit(app, walletAddr, txHash, amount, "USDT")
	if err != nil {
		t.Fatalf("redelivered deposit failed: %v", err)
	}
	if res2["status"] != "already_processed" {
		t.Fatalf("redelivered deposit status = %v, want already_processed", res2["status"])
	}

	bal, err := services.GetBalance(app, userID)
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}
	if bal != amount {
		t.Fatalf("balance after redelivered deposit = %v, want %v (double-credit detected)", bal, amount)
	}

	assertMarkerCount(t, app, txHash, 1)
}

// TestCreditBalanceAndDebitAccounting sanity-checks the credit/debit math and
// the insufficient-funds guard, independent of provider webhooks.
func TestCreditBalanceAndDebitAccounting(t *testing.T) {
	app := setupServicesTestApp(t)
	defer app.Cleanup()

	userID := seededUserID(t, app)

	if _, err := services.CreditBalance(app, userID, 100); err != nil {
		t.Fatalf("credit failed: %v", err)
	}
	if bal, _ := services.GetBalance(app, userID); bal != 100 {
		t.Fatalf("balance = %v, want 100", bal)
	}

	newBal, err := services.DebitBalance(app, userID, 30)
	if err != nil {
		t.Fatalf("debit failed: %v", err)
	}
	if newBal != 70 {
		t.Fatalf("balance after debit = %v, want 70", newBal)
	}

	// Overdraw must fail and leave the balance untouched.
	if _, err := services.DebitBalance(app, userID, 1000); err == nil {
		t.Fatal("expected insufficient-funds error, got nil")
	}
	if bal, _ := services.GetBalance(app, userID); bal != 70 {
		t.Fatalf("balance after failed overdraw = %v, want 70", bal)
	}

	// Non-positive amounts are rejected.
	if _, err := services.CreditBalance(app, userID, 0); err == nil {
		t.Fatal("expected error crediting non-positive amount")
	}
	if _, err := services.DebitBalance(app, userID, -5); err == nil {
		t.Fatal("expected error debiting non-positive amount")
	}
}

// assertMarkerCount asserts how many idempotency markers exist for a tx id.
func assertMarkerCount(t testing.TB, app core.App, txID string, want int) {
	t.Helper()
	records, err := app.FindRecordsByFilter(
		"payment_idempotency",
		"provider_transaction_id = {:txId}",
		"", 100, 0,
		dbx.Params{"txId": txID},
	)
	if err != nil {
		t.Fatalf("failed to query idempotency markers: %v", err)
	}
	if len(records) != want {
		t.Fatalf("idempotency markers for %s = %d, want %d", txID, len(records), want)
	}
}
