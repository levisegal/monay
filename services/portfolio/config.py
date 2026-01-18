from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    listen_addr: str = ":8889"
    holdings_url: str = "http://localhost:8888"
    cache_path: str = "./cache.db"

    model_config = SettingsConfigDict(
        env_prefix="MONAY_PORTFOLIO_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    @property
    def host(self) -> str:
        if self.listen_addr.startswith(":"):
            return "0.0.0.0"
        return self.listen_addr.rsplit(":", 1)[0]

    @property
    def port(self) -> int:
        return int(self.listen_addr.rsplit(":", 1)[-1])


@lru_cache
def get_settings() -> Settings:
    return Settings()
