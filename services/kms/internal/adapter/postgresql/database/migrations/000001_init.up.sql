create or replace function set_updated_at()
    returns trigger as
$$
begin
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;

create table keys
(
    id               uuid primary key                     default gen_random_uuid(),
    seed_fingerprint text                        not null,
    alias            varchar(250)                not null,
    curve            varchar(50)                 not null,
    algorithm        varchar(50)                 not null,
    derivation_path  text                        not null,
    status           varchar(50)                 not null,

    created_at       timestamp                   not null default current_timestamp,
    updated_at       timestamp                   not null default current_timestamp,
    expires_at       timestamp without time zone null,

    unique (seed_fingerprint, derivation_path),
    unique (alias)
);

create trigger trigger_set_updated_at
    before update
    on keys
    for each row
execute function set_updated_at();


create table chains
(
    coin       varchar(150) not null,
    symbol     varchar(10)  not null,
    coin_type  int          not null,
    is_enabled boolean      not null default false,

    created_at timestamp    not null default current_timestamp,
    updated_at timestamp    not null default current_timestamp,

    primary key (coin, symbol, coin_type)
);

create trigger trigger_set_updated_at
    before update
    on chains
    for each row
execute function set_updated_at();


create table event_log
(
    id         uuid primary key                     default gen_random_uuid(),
    payload    jsonb                       not null,
    created_at timestamp without time zone not null default current_timestamp
);
