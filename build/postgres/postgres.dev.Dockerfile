FROM postgres:17

WORKDIR /

COPY init_postgres.sh /docker-entrypoint-initdb.d/000_init.sh
RUN chmod +x /docker-entrypoint-initdb.d/000_init.sh

EXPOSE 5432

# Default entrypoint/cmd from base image will run postgres
