package service

import (
	"time"

	"github.com/mhsanaei/3x-ui/v2/database"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/util/random"
)

type ApiKeyService struct{}

func (s *ApiKeyService) CreateApiKey(description string, expiryTime int64) (*model.ApiKey, error) {
	db := database.GetDB()
	apiKey := &model.ApiKey{
		Key:         "sk-" + random.Seq(45),
		Description: description,
		ExpiryTime:  expiryTime,
		Enabled:     true,
		CreatedAt:   time.Now().UnixMilli(),
	}
	err := db.Create(apiKey).Error
	if err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *ApiKeyService) ListApiKeys() ([]*model.ApiKey, error) {
	db := database.GetDB()
	var apiKeys []*model.ApiKey
	err := db.Order("created_at DESC").Find(&apiKeys).Error
	return apiKeys, err
}

func (s *ApiKeyService) DeleteApiKey(id int) error {
	db := database.GetDB()
	return db.Delete(&model.ApiKey{}, id).Error
}

func (s *ApiKeyService) ToggleApiKey(id int) error {
	db := database.GetDB()
	var apiKey model.ApiKey
	if err := db.First(&apiKey, id).Error; err != nil {
		return err
	}
	return db.Model(&apiKey).Update("enabled", !apiKey.Enabled).Error
}

func (s *ApiKeyService) ValidateApiKey(key string) bool {
	if key == "" {
		return false
	}
	db := database.GetDB()
	var apiKey model.ApiKey
	if err := db.Where("key = ? AND enabled = ?", key, true).First(&apiKey).Error; err != nil {
		return false
	}
	if apiKey.ExpiryTime > 0 && apiKey.ExpiryTime < time.Now().UnixMilli() {
		return false
	}
	return true
}
