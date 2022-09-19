package main

import (
	"context"
	"time"

	v1 "github.com/MaciejTe/accountapiclient/pkg/client/v1"
	"github.com/MaciejTe/accountapiclient/pkg/config"
	"github.com/MaciejTe/accountapiclient/pkg/models"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
)

func main() {
	address := "http://localhost:8080"
	timeout := time.Duration(5) * time.Second
	config := config.NewConfig(&address, timeout, true)
	client := v1.NewAccountClient(*config)
	ctx := context.Background()

	// CREATE
	accountId := uuid.New()
	country := "GB"
	NewAccountRequest := models.AccountData{
		Attributes: &models.AccountAttributes{
			Name:    []string{"Johnny Bravo"},
			Country: &country,
		},
		Type:           "accounts",
		ID:             accountId.String(),
		OrganisationID: accountId.String(),
	}
	acc, err := client.Create(ctx, NewAccountRequest)
	spew.Dump("CREATED ACCOUNT: ", acc, "\n")
	spew.Dump("CREATE ERROR: ", err, "\n\n")

	// FETCH
	account, err := client.Fetch(ctx, accountId.String())
	spew.Dump("FETCHED ACCOUNT: ", account, "\n")
	spew.Dump("FETCH ERROR: ", err, "\n\n")

	// DELETE
	accVersion := 0
	err = client.Delete(ctx, accountId.String(), int64(accVersion))
	spew.Dump("DELETE RESULT: ", err, "\n\n")
}
