.PHONY: up stop down logs db-migrate db-studio dev build reset

# Instala dependências, compila todos os packages e roda migrations.
build:
	npm install
	npx turbo run build
	docker compose up -d --wait
	npx turbo run db:migrate --filter=@memory-card/db
	docker compose down

# Para processos locais e derruba serviços Docker.
stop:
	-pkill -f "tsx watch" || true
	-pkill -f "vite" || true
	docker compose down

# Sobe banco, API e frontend juntos.
up: stop
	docker compose up -d
	npx turbo run dev

# Derruba os serviços Docker.
down:
	docker compose down

# Acompanha os logs dos serviços Docker.
logs:
	docker compose logs -f

# Executa as migrations do banco.
db-migrate:
	npx turbo run db:migrate --filter=@memory-card/db

# Abre o Drizzle Studio.
db-studio:
	npx turbo run db:studio --filter=@memory-card/db

# Recria os serviços Docker removendo volumes.
reset:
	docker compose down -v && docker compose up -d
