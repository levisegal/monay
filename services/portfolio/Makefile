.PHONY: dev run test lint format check

dev:
	uv run uvicorn main:app --reload --port 8889

run:
	uv run uvicorn main:app --port 8889

test:
	uv run pytest -v

lint:
	uv run ruff check .

format:
	uv run ruff format .

check: lint test
