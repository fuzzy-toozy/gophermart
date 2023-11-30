package services

import (
	"encoding/json"
	"io"
	"net/http"

	serviceErrs "github.com/fuzzy-toozy/gophermart/internal/errors"
)

type AccrualService struct {
	client *http.Client
	addr   string
}

type AccrualOrderInfo struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

var (
	ErrAccuralTooManyReq = "accural service too many requests"
	ErrAccuralInternal   = "accural service internal error"
)

func NewAccrualService(client *http.Client, addr string) *AccrualService {
	return &AccrualService{
		client: client,
		addr:   addr,
	}
}

func (s *AccrualService) GetOrderInfo(orderNumber string) (*AccrualOrderInfo, error) {
	client := http.Client{}
	reqURL := s.addr + orderNumber

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusTooManyRequests:
		return nil, serviceErrs.NewServiceError(res.StatusCode, ErrAccuralInternal)
	case http.StatusInternalServerError:
		return nil, serviceErrs.NewServiceError(res.StatusCode, ErrAccuralInternal)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, serviceErrs.NewServiceError(http.StatusInternalServerError, err.Error())
	}

	orderInfo := new(AccrualOrderInfo)

	err = json.Unmarshal(data, &orderInfo)
	if err != nil {
		return nil, serviceErrs.NewServiceError(http.StatusInternalServerError, err.Error())
	}

	return orderInfo, nil
}
