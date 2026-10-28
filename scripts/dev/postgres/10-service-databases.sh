#!/bin/sh
# Creates a role and an owned database for every entry of SERVICE_DATABASES,
# a space separated list of "database:user:password" (no ':' in the password).
#
# Mounted into /docker-entrypoint-initdb.d, so it only runs when the data
# volume is empty. To apply it to an existing volume it is safe to re-run:
#   docker compose exec postgres sh /docker-entrypoint-initdb.d/10-service-databases.sh
# Schemas are not created here, each service migrates its own on start.
set -eu

for entry in ${SERVICE_DATABASES:-}; do
	db=${entry%%:*}
	rest=${entry#*:}
	user=${rest%%:*}
	pass=${rest#*:}

	echo "init-db: ensuring database '$db' owned by '$user'"

	psql -v ON_ERROR_STOP=1 -U "${POSTGRES_USER:-postgres}" -d postgres \
		-v db="$db" -v usr="$user" -v pass="$pass" <<'SQL'
select format('create role %I login password %L', :'usr', :'pass')
where not exists (select from pg_roles where rolname = :'usr') \gexec
select format('create database %I owner %I', :'db', :'usr')
where not exists (select from pg_database where datname = :'db') \gexec
SQL
done
