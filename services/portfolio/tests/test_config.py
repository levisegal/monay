import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from config import Settings


def test_default_settings():
    settings = Settings()
    assert settings.listen_addr == ":8889"
    assert settings.holdings_url == "http://localhost:8888"
    assert settings.cache_path == "./cache.db"


def test_host_port_properties():
    settings = Settings(listen_addr=":8889")
    assert settings.host == "0.0.0.0"
    assert settings.port == 8889


def test_explicit_host():
    settings = Settings(listen_addr="127.0.0.1:9000")
    assert settings.host == "127.0.0.1"
    assert settings.port == 9000


def test_env_override(monkeypatch):
    monkeypatch.setenv("MONAY_PORTFOLIO_LISTEN_ADDR", ":9999")
    monkeypatch.setenv("MONAY_PORTFOLIO_HOLDINGS_URL", "http://holdings:8888")
    monkeypatch.setenv("MONAY_PORTFOLIO_CACHE_PATH", "/data/cache.db")

    settings = Settings()
    assert settings.listen_addr == ":9999"
    assert settings.holdings_url == "http://holdings:8888"
    assert settings.cache_path == "/data/cache.db"
