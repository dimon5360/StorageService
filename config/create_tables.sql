begin;

-- create schema
create schema if not exists barmapschema; 

-- table bars
create table if not exists bars (
    id serial primary key,
    title text not null,
    address text not null,
    description text not null,
    drinks_id bigint[],
    created_at timestamp not null,
    updated_at timestamp not null
);

-- table drinks
create table if not exists drinks (
    id serial primary key,
    title text not null,
    price int not null,
    type int not null,
    description text not null,
    bar_id bigint not null references bars(id) on delete cascade,
    ingredients_id bigint[],
    created_at timestamp not null,
    updated_at timestamp not null
);

--table ingredients
create table if not exists ingredients (
    id serial primary key,
    title text not null,
    amount int not null,
    drink_id bigint not null references drinks(id) on delete cascade,
    created_at timestamp not null,
    updated_at timestamp not null
);

commit;