package utils

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
	"gorm.io/gorm"
)

var (
	ErrInvalidTxManager              = fmt.Errorf("invalid transaction manager: both DB and Firestore client are nil")
	ErrInvalidTxFunction             = fmt.Errorf("invalid transaction function: function cannot be nil")
	ErrFirestoreClientNotInitialized = fmt.Errorf("Firestore client is not initialized")
	ErrGormDBNotInitialized          = fmt.Errorf("GORM DB is not initialized")
)

// TxManager is a struct that manages database transactions using GORM. It provides a method to execute a function within a transaction context, handling the commit and rollback logic based on the success or failure of the function execution.
type TxManager struct {
	DB              *gorm.DB
	firestoreClient *firestore.Client
}

// NewTxManager creates a new instance of TxManager with the provided GORM database connection and Firestore client. This allows the application to manage transactions in a consistent way across different parts of the codebase by using this transaction manager to execute functions that require transactional behavior.
func NewTxManager(db *gorm.DB, firestoreClient *firestore.Client) *TxManager {
	return &TxManager{DB: db, firestoreClient: firestoreClient}
}

// Do executes the provided function within a transaction context. It begins a new transaction, calls the function with the transaction context, and handles committing or rolling back the transaction based on whether an error occurred during the function execution. The function returns a boolean indicating if there was a server error and any error that occurred during the process.
func (txm *TxManager) Do(ctx context.Context, fn func(ctx context.Context, txGorm *gorm.DB, txFirestore *firestore.Transaction) (bool, error)) (bool, error) {
	var serverError bool
	var err error
	var txGorm *gorm.DB
	if txm.DB == nil || txm.firestoreClient == nil {
		return true, ErrInvalidTxManager
	}
	if fn == nil {
		return true, ErrInvalidTxFunction
	}
	txGorm = txm.DB.WithContext(ctx).Begin()
	err = txm.firestoreClient.RunTransaction(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) error {
		serverError, err = fn(ctx, txGorm, txFirestore)
		return err
	})
	if serverError || err != nil {
		txGorm.Rollback()
		return serverError, err
	}
	if err := txGorm.Commit().Error; err != nil {
		serverError = true
		return serverError, err
	}
	return false, nil
}

func (txm *TxManager) DoFirestore(ctx context.Context, fn func(ctx context.Context, txFirestore *firestore.Transaction) (bool, error)) (bool, error) {
	var serverError bool
	var err error
	if txm.firestoreClient == nil {
		return true, ErrFirestoreClientNotInitialized
	}
	if fn == nil {
		return true, ErrInvalidTxFunction
	}
	err = txm.firestoreClient.RunTransaction(ctx, func(ctx context.Context, txFirestore *firestore.Transaction) error {
		serverError, err = fn(ctx, txFirestore)
		return err
	})
	return serverError, err
}

func (txm *TxManager) DoGorm(ctx context.Context, fn func(ctx context.Context, txGorm *gorm.DB) (bool, error)) (bool, error) {
	var serverError bool
	var err error
	var txGorm *gorm.DB
	if txm.DB == nil {
		return true, ErrGormDBNotInitialized
	}
	if fn == nil {
		return true, ErrInvalidTxFunction
	}
	txGorm = txm.DB.WithContext(ctx).Begin()
	serverError, err = fn(ctx, txGorm)
	if err != nil {
		txGorm.Rollback()
		return serverError, err
	}
	if err := txGorm.Commit().Error; err != nil {
		serverError = true
		return serverError, err
	}
	return false, nil
}
