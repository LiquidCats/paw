create table state (
    key varchar primary key,
    value jsonb default null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);