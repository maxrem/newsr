MIGRATE_SEQ?=

migrate-create:
	docker-compose run migrate create -ext sql -dir /migrations -seq $(MIGRATE_SEQ)

migrate-up:
	docker-compose run migrate -path=/migrations -database mysql://newsr_user:passwd@tcp\(db:3306\)/newsr up
