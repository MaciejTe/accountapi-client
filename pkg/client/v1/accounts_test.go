package v1

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/MaciejTe/accountapiclient/pkg/config"
	"github.com/MaciejTe/accountapiclient/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNewAccountClient(t *testing.T) {
	t.Parallel()

	trVerify := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	trSkipVerify := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	testCases := map[string]struct {
		Config         config.Config
		ExpectedResult *AccountClient
	}{
		"ok": {
			Config: config.Config{
				Timeout:    20000000,
				SkipVerify: true,
			},
			ExpectedResult: &AccountClient{
				config: config.Config{
					Timeout:    20000000,
					SkipVerify: true,
				},
				http: http.Client{
					Transport: trSkipVerify,
					Timeout:   20000000,
				},
			},
		},
		"ok, default skipVerify": {
			ExpectedResult: &AccountClient{
				config: config.Config{
					Timeout: 0,
				},
				http: http.Client{
					Transport: trVerify,
					Timeout:   0,
				},
			},
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(fmt.Sprintf("tc %s", name), func(t *testing.T) {
			result := NewAccountClient(tc.Config)
			assert.Equal(t, tc.ExpectedResult, result)
		})
	}
}

type AccountAPIClientTestSuite struct {
	suite.Suite
	client            *AccountClient
	createdAccountIds []uuid.UUID
	country           string
	name              string
	accountData       models.AccountData
	id                uuid.UUID
}

func (suite *AccountAPIClientTestSuite) SetupSuite() {
	conf := config.Config{
		Address:    config.DefaultAddress,
		Timeout:    config.DefaultReqTimeout,
		SkipVerify: true,
	}
	suite.client = NewAccountClient(conf)
	suite.country = "GB"
	suite.name = "Johnny Bravo"
}

func (suite *AccountAPIClientTestSuite) SetupTest() {
	suite.id = uuid.New()
	suite.accountData = models.AccountData{
		ID:             suite.id.String(),
		OrganisationID: suite.id.String(),
		Type:           "accounts",
		Attributes: &models.AccountAttributes{
			Country: &suite.country,
			Name:    []string{suite.name},
		},
	}
	suite.client.config.Address = config.DefaultAddress
}

func (suite *AccountAPIClientTestSuite) TestCreateBadRequest() {
	improperAccountData := models.AccountData{
		ID:             "1234",
		OrganisationID: uuid.New().String(),
		Type:           "accounts",
		Attributes: &models.AccountAttributes{
			Country: &suite.country,
			Name:    []string{suite.name},
		},
	}
	var expectedResult *models.AccountData
	expectedErrMsg := "error code: , message: validation failure list:\nvalidation failure list:\nid in body must be of type uuid: \"1234\""
	result, err := suite.client.Create(context.Background(), improperAccountData)
	assert.Equal(suite.T(), expectedResult, result)
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TestCreateAccount() {
	expectedVersion := new(int64)
	*expectedVersion = 0
	expectedResult := models.AccountData{
		ID:             suite.id.String(),
		OrganisationID: suite.id.String(),
		Type:           "accounts",
		Version:        expectedVersion,
		Attributes: &models.AccountAttributes{
			Country: &suite.country,
			Name:    []string{suite.name},
		},
	}
	result, err := suite.client.Create(context.Background(), suite.accountData)
	assert.Equal(suite.T(), &expectedResult, result)
	assert.Nil(suite.T(), err)
	suite.createdAccountIds = append(suite.createdAccountIds, suite.id)
}

func (suite *AccountAPIClientTestSuite) TestCreateCtxNil() {
	expectedVersion := new(int64)
	*expectedVersion = 0
	expectedResult := models.AccountData{
		ID:             suite.id.String(),
		OrganisationID: suite.id.String(),
		Type:           "accounts",
		Version:        expectedVersion,
		Attributes: &models.AccountAttributes{
			Country: &suite.country,
			Name:    []string{suite.name},
		},
	}
	result, err := suite.client.Create(nil, suite.accountData)
	assert.Equal(suite.T(), &expectedResult, result)
	assert.Nil(suite.T(), err)
	suite.createdAccountIds = append(suite.createdAccountIds, suite.id)
}

func (suite *AccountAPIClientTestSuite) TestCreateImproperAddress() {
	suite.client.config.Address = "http://improper_address:9000"
	var expectedResult *models.AccountData
	result, err := suite.client.Create(context.Background(), suite.accountData)
	assert.Equal(suite.T(), expectedResult, result)
	expectedErrMsg := "Post \"http://improper_address:9000/v1/organisation/accounts\": dial tcp: lookup improper_address: Temporary failure in name resolution"
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TestFetchCtxNil() {
	account, err := suite.client.Create(context.Background(), suite.accountData)
	assert.Nil(suite.T(), err)
	suite.createdAccountIds = append(suite.createdAccountIds, suite.id)

	expectedVersion := new(int64)
	*expectedVersion = 0
	expectedResult := models.AccountData{
		ID:             account.ID,
		OrganisationID: account.ID,
		Type:           "accounts",
		Version:        expectedVersion,
		Attributes: &models.AccountAttributes{
			Country: &suite.country,
			Name:    []string{suite.name},
		},
	}
	result, err := suite.client.Fetch(nil, account.ID)
	assert.Equal(suite.T(), &expectedResult, result)
	assert.Nil(suite.T(), err)
}

func (suite *AccountAPIClientTestSuite) TestFetchBadRequest() {
	result, err := suite.client.Fetch(nil, "1234")
	expectedErrMsg := "error code: , message: id is not a valid uuid"
	assert.Nil(suite.T(), result)
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TestFetchImproperAddress() {
	suite.client.config.Address = "http://improper_address:9000"
	result, err := suite.client.Fetch(nil, "1234")
	assert.Nil(suite.T(), result)
	expectedErrMsg := "Get \"http://improper_address:9000/v1/organisation/accounts/1234\": dial tcp: lookup improper_address: Temporary failure in name resolution"
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TestDeleteImproperAddress() {
	suite.client.config.Address = "http://improper_address:9000"
	err := suite.client.Delete(context.Background(), "1234", 0)
	expectedErrMsg := "Delete \"http://improper_address:9000/v1/organisation/accounts/1234?version=0\": dial tcp: lookup improper_address: Temporary failure in name resolution"
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TestDeleteCtxNil() {
	_, err := suite.client.Create(context.Background(), suite.accountData)
	assert.Nil(suite.T(), err)
	suite.createdAccountIds = append(suite.createdAccountIds, suite.id)

	err = suite.client.Delete(nil, suite.id.String(), 0)
	assert.Nil(suite.T(), err)
}

func (suite *AccountAPIClientTestSuite) TestDeleteBadRequest() {
	err := suite.client.Delete(context.Background(), "1234", 0)
	expectedErrMsg := "error code: , message: id is not a valid uuid"
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TestDeleteResourceDoesNotExist() {
	err := suite.client.Delete(context.Background(), uuid.NewString(), 0)
	expectedErrMsg := "specified resource does not exist"
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TestDeleteConflict() {
	_, err := suite.client.Create(context.Background(), suite.accountData)
	assert.Nil(suite.T(), err)
	suite.createdAccountIds = append(suite.createdAccountIds, suite.id)

	improperVersion := 4
	err = suite.client.Delete(context.Background(), suite.id.String(), int64(improperVersion))
	expectedErrMsg := "specified version incorrect"
	assert.EqualError(suite.T(), err, expectedErrMsg)
}

func (suite *AccountAPIClientTestSuite) TearDownSuite() {
	suite.client.config.Address = config.DefaultAddress
	for _, accountId := range suite.createdAccountIds {
		_ = suite.client.Delete(context.Background(), fmt.Sprint(accountId), 0)
	}
}

func TestAccountAPIClientTestSuite(t *testing.T) {
	suite.Run(t, new(AccountAPIClientTestSuite))
}
