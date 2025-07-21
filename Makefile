start:
	docker compose -f docker-compose.yml up --build -d

stop:
	docker compose -f docker-compose.yml down

clean:
	docker compose -f docker-compose.yml down --volumes --remove-orphans
