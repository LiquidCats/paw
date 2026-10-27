create table if not exists gaps (
    id uuid default uuidv7() primary key,
    chain varchar(255) not null,
    from_block bigint not null,
    to_block bigint not null,
    current_block bigint not null,
    status varchar(255) not null default 'unprocessed',
    created_at timestamp default current_timestamp
);

create index if not exists idx_gaps_current_block on gaps(chain, from_block);

create table if not exists canonical_chain (
    height bigint not null primary key,
    chain varchar(255) not null,
    hash varchar(1000) not null,
    previous_hash varchar(1000) not null,
    created_at timestamp default current_timestamp
);

create index if not exists idx_canonical_chain_height on canonical_chain(chain, hash);

