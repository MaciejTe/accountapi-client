package v1

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/MaciejTe/accountapiclient/pkg/models"

	"github.com/MaciejTe/accountapiclient/pkg/config"

	"github.com/pkg/errors"
)

var (
	apiVersion       = "v1"
	createAccountUrl = fmt.Sprintf("/%s/organisation/accounts", apiVersion)
	AccountUrlById   = fmt.Sprintf("/%s/organisation/accounts/:id", apiVersion)
)

// AccountClient implements Client interface
type AccountClient struct {
	config config.Config
	http   http.Client
}

// NewAccountClient creates pointer to AccountClient structure
func NewAccountClient(conf config.Config) *AccountClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipVerify},
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	return &AccountClient{
		config: conf,
		http: http.Client{
			Transport: tr,
			Timeout:   conf.Timeout,
		},
	}
}

// Create creates new account in Account API
func (ac *AccountClient) Create(ctx context.Context, accountData models.AccountData) (*models.AccountData, error) {
	log.Debugf("creating an account with data: %v", accountData)
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, ac.config.Timeout)
	defer cancel()

	var buf bytes.Buffer
	requestPayload := models.Account{
		Data: accountData,
	}
	err := json.NewEncoder(&buf).Encode(requestPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx, "POST",
		ac.config.Address+createAccountUrl,
		&buf,
	)
	if err != nil {
		log.Errorf("create: failed to create request: %v", err)
		return nil, err
	}

	req.Header.Add("Host", ac.config.Address)
	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Date", time.Now().String())
	req.Header.Add("Content-Type", "application/vnd.api+json")

	rsp, err := ac.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusCreated:
		account := models.Account{}
		if err := json.NewDecoder(rsp.Body).Decode(&account); err != nil {
			return nil, errors.Wrap(err, "error parsing created account data")
		}
		return &account.Data, nil
	case http.StatusBadRequest:
		err := models.ErrBadRequest{}
		if err := json.NewDecoder(rsp.Body).Decode(&err); err != nil {
			return nil, errors.Wrap(err, "payload parsing error")
		}
		respErr := errors.New(fmt.Sprintf("error code: %s, message: %s", err.ErrorCode, err.ErrorMessage))
		log.Errorf("create: %v", respErr.Error())
		return nil, respErr
	default:
		err := errors.Errorf("creating account resulted in unexpected HTTP error code: %v",
			rsp.StatusCode)
		log.Errorf("create: %v", err.Error())
		return nil, err
	}
}

// Fetch retrieves given account's data from the account API
func (ac *AccountClient) Fetch(ctx context.Context, id string) (*models.AccountData, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, ac.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx, "GET",
		ac.config.Address+strings.Replace(AccountUrlById, ":id", id, 1),
		nil,
	)
	if err != nil {
		log.Errorf("fetch: failed to create request: %v", err)
		return nil, err
	}

	req.Header.Add("Host", ac.config.Address)
	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Date", time.Now().String())

	rsp, err := ac.http.Do(req)
	if err != nil {
		log.Errorf("fetch: failed to send request: %v", err)
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNotFound:
		return nil, nil
	case http.StatusOK:
		account := models.Account{}
		if err := json.NewDecoder(rsp.Body).Decode(&account); err != nil {
			return nil, errors.Wrap(err, "error parsing account")
		}
		return &account.Data, nil
	case http.StatusBadRequest:
		err := models.ErrBadRequest{}
		if err := json.NewDecoder(rsp.Body).Decode(&err); err != nil {
			return nil, errors.Wrap(err, "payload parsing error")
		}
		requestErr := errors.New(fmt.Sprintf("error code: %s, message: %s", err.ErrorCode, err.ErrorMessage))
		log.Errorf("fetch: bad request: %v", requestErr)
		return nil, requestErr
	default:
		err := errors.Errorf("getting account data resulted in unexpected HTTP error code: %v",
			rsp.StatusCode)
		log.Errorf("fetch: %v", err.Error())
		return nil, err
	}
}

// Delete removes account with given ID and version from the system
func (ac *AccountClient) Delete(ctx context.Context, id string, version int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, ac.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx, "DELETE",
		ac.config.Address+strings.Replace(AccountUrlById, ":id", id, 1),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Add("Host", ac.config.Address)
	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Date", time.Now().String())

	query := req.URL.Query()
	query.Add("version", fmt.Sprint(version))
	req.URL.RawQuery = query.Encode()

	rsp, err := ac.http.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return errors.New("specified resource does not exist")
	case http.StatusBadRequest:
		err := models.ErrBadRequest{}
		if err := json.NewDecoder(rsp.Body).Decode(&err); err != nil {
			return errors.Wrap(err, "payload parsing error")
		}
		requestErr := errors.New(fmt.Sprintf("error code: %s, message: %s", err.ErrorCode, err.ErrorMessage))
		log.Errorf("delete: %v", requestErr.Error())
		return requestErr
	case http.StatusConflict:
		err := errors.New("specified version incorrect")
		log.Warnf("delete: %v", err.Error())
		return err
	default:
		err := errors.Errorf("deleting account resulted in unexpected HTTP error code: %v",
			rsp.StatusCode)
		log.Errorf("delete: %v", err.Error())
		return err
	}
}
