package inttest

import (
	"os"
	"testing"

	"github.com/kozlov-ma/sesc-backend/apiclient/client/authentication"
	"github.com/kozlov-ma/sesc-backend/apiclient/models"
	"github.com/stretchr/testify/require"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

func alogin(t *testing.T) (host, basePath, scheme, bearerToken string) {
	t.Helper()

	host = os.Getenv("TEST_API_HOST")
	basePath = os.Getenv("TEST_API_BASE_PATH")
	scheme = "http"

	username := os.Getenv("TEST_ADMIN_USERNAME")
	password := os.Getenv("TEST_ADMIN_PASSWORD")

	transport := httptransport.New(host, basePath, []string{scheme})

	ctx := t.Context()
	params := authentication.NewPostAuthAdminLoginParamsWithContext(ctx).WithRequest(&models.APICredentialsRequest{
		Password: &username,
		Username: &password,
	})

	res, err := authentication.New(transport, strfmt.Default).PostAuthAdminLogin(params)
	require.NoError(t, err, "couldn't log in as admin")

	return host, basePath, scheme, *res.Payload.Token
}

func ulogin(t *testing.T, username, password string) (host, basePath, scheme, bearerToken string) {
	host = os.Getenv("TEST_API_HOST")
	basePath = os.Getenv("TEST_API_BASE_PATH")
	scheme = "http"

	transport := httptransport.New(host, basePath, []string{scheme})

	ctx := t.Context()
	params := authentication.NewPostAuthLoginParamsWithContext(ctx).WithRequest(&models.APICredentialsRequest{
		Password: &username,
		Username: &password,
	})

	res, err := authentication.New(transport, strfmt.Default).PostAuthLogin(params)
	require.NoError(t, err, "couldn't log in as user")

	return host, basePath, scheme, *res.Payload.Token
}
