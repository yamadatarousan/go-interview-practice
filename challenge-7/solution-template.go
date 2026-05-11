// Package challenge7 contains the solution for Challenge 7: Bank Account with Error Handling.
package challenge7

import (
	"sync"
	// Add any other necessary imports
	"fmt"
)

// BankAccount represents a bank account with balance management and minimum balance requirements.
type BankAccount struct {
	ID         string
	Owner      string
	Balance    float64
	MinBalance float64
	mu         sync.Mutex // For thread safety
}

// Constants for account operations
const (
	MaxTransactionAmount = 10000.0 // Example limit for deposits/withdrawals
)

// Custom error types

// AccountError is a general error type for bank account operations.
type AccountError struct {
	// Implement this error type
	Op  string
	ID  string
	Msg string
}

func (e *AccountError) Error() string {
	// Implement error message
	return fmt.Sprintf("account %s: %s failed: %s", e.ID, e.Op, e.Msg)
}

// InsufficientFundsError occurs when a withdrawal or transfer would bring the balance below minimum.
type InsufficientFundsError struct {
	// Implement this error type
	AccountID  string
	Balance    float64
	Amount     float64
	MinBalance float64
}

func (e *InsufficientFundsError) Error() string {
	// Implement error message
	return fmt.Sprintf("insufficient funds: account %s has %.2f, need %.2f (min %.2f)",
		e.AccountID, e.Balance, e.Amount, e.MinBalance)
}

// NegativeAmountError occurs when an amount for deposit, withdrawal, or transfer is negative.
type NegativeAmountError struct {
	// Implement this error type
	Amount float64
}

func (e *NegativeAmountError) Error() string {
	// Implement error message
	return fmt.Sprintf("negative amount: %.2f", e.Amount)
}

// ExceedsLimitError occurs when a deposit or withdrawal amount exceeds the defined limit.
type ExceedsLimitError struct {
	// Implement this error type
	Amount float64
	Limit  float64
}

func (e *ExceedsLimitError) Error() string {
	// Implement error message
	return fmt.Sprintf("amount %.2f exceeds limit %.2f", e.Amount, e.Limit)
}

// NewBankAccount creates a new bank account with the given parameters.
// It returns an error if any of the parameters are invalid.
func NewBankAccount(id, owner string, initialBalance, minBalance float64) (*BankAccount, error) {
	// Implement account creation with validation
	if id == "" {
		return nil, &AccountError{Op: "create", ID: id, Msg: "id required"}
	}
	if owner == "" {
		return nil, &AccountError{Op: "create", ID: id, Msg: "owner required"}
	}
	if initialBalance < 0 {
		return nil, &NegativeAmountError{Amount: initialBalance}
	}
	if minBalance < 0 {
		return nil, &NegativeAmountError{Amount: minBalance}
	}
	if initialBalance < minBalance {
		return nil, &InsufficientFundsError{
			AccountID: id, Balance: initialBalance, Amount: 0, MinBalance: minBalance,
		}
	}
	return &BankAccount{
		ID: id, Owner: owner, Balance: initialBalance, MinBalance: minBalance,
	}, nil
}

// Deposit adds the specified amount to the account balance.
// It returns an error if the amount is invalid or exceeds the transaction limit.
func (a *BankAccount) Deposit(amount float64) error {
	// Implement deposit functionality with proper error handling
	a.mu.Lock()
	defer a.mu.Unlock()

	if amount < 0 {
		return &NegativeAmountError{Amount: amount}
	}
	if amount > MaxTransactionAmount {
		return &ExceedsLimitError{Amount: amount, Limit: MaxTransactionAmount}
	}
	a.Balance += amount
	return nil
}

// Withdraw removes the specified amount from the account balance.
// It returns an error if the amount is invalid, exceeds the transaction limit,
// or would bring the balance below the minimum required balance.
func (a *BankAccount) Withdraw(amount float64) error {
	// Implement withdrawal functionality with proper error handling
	a.mu.Lock()
	defer a.mu.Unlock()

	if amount < 0 {
		return &NegativeAmountError{Amount: amount}
	}
	if amount > MaxTransactionAmount {
		return &ExceedsLimitError{Amount: amount, Limit: MaxTransactionAmount}
	}
	if a.Balance-amount < a.MinBalance {
		return &InsufficientFundsError{
			AccountID: a.ID, Balance: a.Balance, Amount: amount, MinBalance: a.MinBalance,
		}
	}
	a.Balance -= amount
	return nil
}

// Transfer moves the specified amount from this account to the target account.
// It returns an error if the amount is invalid, exceeds the transaction limit,
// or would bring the balance below the minimum required balance.
func (a *BankAccount) Transfer(amount float64, target *BankAccount) error {
	// Implement transfer functionality with proper error handling
	if target == nil {
		return &AccountError{Op: "transfer", ID: a.ID, Msg: "target account required"}
	}
	if amount < 0 {
		return &NegativeAmountError{Amount: amount}
	}
	if amount > MaxTransactionAmount {
		return &ExceedsLimitError{Amount: amount, Limit: MaxTransactionAmount}
	}

	// デッドロック回避のため ID 順にロック
	first, second := a, target
	if a.ID > target.ID {
		first, second = target, a
	}
	first.mu.Lock()
	second.mu.Lock()
	defer first.mu.Unlock()
	defer second.mu.Unlock()

	if a.Balance-amount < a.MinBalance {
		return &InsufficientFundsError{
			AccountID: a.ID, Balance: a.Balance, Amount: amount, MinBalance: a.MinBalance,
		}
	}
	a.Balance -= amount
	target.Balance += amount
	return nil
}
