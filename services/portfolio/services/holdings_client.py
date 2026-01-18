import httpx

from config import get_settings


class HoldingsClient:
    def __init__(self, base_url: str | None = None):
        settings = get_settings()
        self.base_url = (base_url or settings.holdings_url).rstrip("/")
        self._client = httpx.AsyncClient(base_url=self.base_url, timeout=30.0)

    async def close(self):
        await self._client.aclose()

    async def get_accounts(self) -> list[dict]:
        resp = await self._client.get("/api/v1/accounts")
        resp.raise_for_status()
        data = resp.json()
        return data.get("accounts", [])

    async def get_holdings(self, account_id: str | None = None) -> list[dict]:
        params = {}
        if account_id:
            params["account_id"] = account_id
        resp = await self._client.get("/api/v1/holdings", params=params)
        resp.raise_for_status()
        data = resp.json()
        return data.get("holdings", [])
