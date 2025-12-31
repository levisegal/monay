package plaid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/plaid/plaid-go/v29/plaid"
)

func (c *Client) CreateLinkToken(ctx context.Context, userID string) (string, error) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userID,
	}

	req := plaid.NewLinkTokenCreateRequest(
		"Monay",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		user,
	)
	req.SetProducts([]plaid.Products{plaid.PRODUCTS_INVESTMENTS})
	// OAuth institutions (Schwab, etc.) require redirect URI in production
	if c.redirectURI != "" {
		req.SetRedirectUri(c.redirectURI)
	}

	resp, _, err := c.api.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", plaidError(err)
	}

	return resp.GetLinkToken(), nil
}

func ServeLinkAndWait(linkToken string, timeout time.Duration) (string, error) {
	publicTokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, linkHTML, linkToken)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		publicToken := r.URL.Query().Get("public_token")
		if publicToken == "" {
			http.Error(w, "missing public_token", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Success! You can close this window.</h2></body></html>")
		publicTokenCh <- publicToken
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case token := <-publicTokenCh:
		server.Shutdown(context.Background())
		return token, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		server.Shutdown(context.Background())
		return "", fmt.Errorf("timeout waiting for callback")
	}
}

const linkHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Connect Account - Monay</title>
  <script src="https://cdn.plaid.com/link/v2/stable/link-initialize.js"></script>
  <style>
    body { font-family: system-ui; max-width: 600px; margin: 100px auto; text-align: center; }
    button { padding: 16px 32px; font-size: 18px; cursor: pointer; }
  </style>
</head>
<body>
  <h1>Connect Your Account</h1>
  <p>Click below to securely connect your brokerage account via Plaid.</p>
  <button id="link-btn">Connect Account</button>
  <p id="status"></p>
  <script>
    const handler = Plaid.create({
      token: '%s',
      onSuccess: (public_token, metadata) => {
        document.getElementById('status').innerText = 'Connected! Redirecting...';
        window.location.href = '/callback?public_token=' + encodeURIComponent(public_token);
      },
      onExit: (err, metadata) => {
        if (err) {
          document.getElementById('status').innerText = 'Error: ' + err.display_message;
        }
      },
    });
    document.getElementById('link-btn').addEventListener('click', () => handler.open());
  </script>
</body>
</html>`
