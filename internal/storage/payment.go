package storage

import (
	"bytes"
	"dbhose/internal/domain"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (sm *StorageManager) CreatePayment(payment domain.Payment) error {
	payment.ID = uuid.New().String()
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()

	key := fmt.Sprintf("payments/%s/%s.json", payment.Email, payment.ID)
	paymentBytes, err := json.Marshal(payment)
	if err != nil {
		return err
	}
	return sm.UploadToS3(key, bytes.NewReader(paymentBytes))
}

func (sm *StorageManager) GetPayment(userEmail, paymentID string) (domain.Payment, error) {
	key := fmt.Sprintf("payments/%s/%s.json", userEmail, paymentID)
	result, err := sm.DownloadFromS3(key)
	if err != nil {
		return domain.Payment{}, err
	}
	defer result.Body.Close()

	var payment domain.Payment
	if err := json.NewDecoder(result.Body).Decode(&payment); err != nil {
		return domain.Payment{}, err
	}
	return payment, nil
}

func (sm *StorageManager) UpdatePayment(payment domain.Payment) error {
	payment.UpdatedAt = time.Now()
	key := fmt.Sprintf("payments/%s/%s.json", payment.Email, payment.ID)
	paymentBytes, err := json.Marshal(payment)
	if err != nil {
		return err
	}
	return sm.UploadToS3(key, bytes.NewReader(paymentBytes))
}

func (sm *StorageManager) ListPayments(userEmail string) ([]domain.Payment, error) {
	prefix := fmt.Sprintf("payments/%s/", userEmail)
	files, err := sm.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	var payments []domain.Payment
	for _, file := range files {
		result, err := sm.DownloadFromS3(file)
		if err != nil {
			return nil, err
		}
		defer result.Body.Close()

		var payment domain.Payment
		if err := json.NewDecoder(result.Body).Decode(&payment); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, nil
}
