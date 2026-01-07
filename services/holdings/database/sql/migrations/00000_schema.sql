-- +goose Up
create schema if not exists monay;

-- +goose Down
drop schema if exists monay cascade;


