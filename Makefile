.phony: run_app

run_app: 
	docker compose up -d

stop:
	docker compose down