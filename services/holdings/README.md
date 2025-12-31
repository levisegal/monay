# Holdings Service

Investment holdings aggregation via Plaid.

## Quick Start

```bash
# Copy env template
cp env.example .env
# Edit .env with your Plaid credentials

# Start the service (via Docker)
cd ../../build && make up

# Or run locally
make run
```

## Local Development with Plaid OAuth

OAuth institutions (Schwab, Fidelity, etc.) require HTTPS redirect URIs in production.

### Setup ngrok

1. **Get authtoken** from [dashboard.ngrok.com](https://dashboard.ngrok.com/get-started/your-authtoken)

2. **Add to `.env`:**
   ```
   NGROK_AUTHTOKEN=your_authtoken
   MONAY_HOLDINGS_PLAID_REDIRECT_URI=https://your-domain.ngrok-free.dev/
   ```

3. **Add redirect URI to Plaid Dashboard:**
   - Go to Developers → API → Allowed redirect URIs
   - Add your ngrok domain

4. **Start ngrok:**
   ```bash
   make ngrok
   ```

5. **Test at** `https://your-domain.ngrok-free.dev/`

## Plaid Institution Access

Some institutions require explicit registration:
- **Schwab**: Requires approval via Plaid Integrations
- **Fidelity, Vanguard**: Generally available

Check Plaid Dashboard → Integrations for institution-specific requirements.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `MONAY_HOLDINGS_PLAID_CLIENT_ID` | Plaid client ID |
| `MONAY_HOLDINGS_PLAID_SECRET` | Plaid secret |
| `MONAY_HOLDINGS_PLAID_ENV` | `sandbox` or `production` |
| `MONAY_HOLDINGS_PLAID_REDIRECT_URI` | HTTPS URL for OAuth callback |
| `MONAY_HOLDINGS_LISTEN_ADDR` | Server listen address (default `:8888`) |
| `NGROK_AUTHTOKEN` | ngrok authtoken for local HTTPS |

## Make Targets

```bash
make build          # Build binary
make run            # Run server locally
make ngrok          # Start ngrok tunnel
make proto.generate # Regenerate protobuf code
make test           # Run tests
```

