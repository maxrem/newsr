.PHONY: migrate-create migrate-up truncate-table

MIGRATE_SEQ?=

migrate-create:
	docker-compose run migrate create -ext sql -dir /migrations -seq $(MIGRATE_SEQ)

migrate-up:
	docker-compose run migrate -path=/migrations -database mysql://newsr_user:passwd@tcp\(db:3306\)/newsr up

NEWSR_TABLE?=article
truncate-table:
	docker-compose exec db mysql -ppasswd newsr -e 'TRUNCATE TABLE $(NEWSR_TABLE)'
