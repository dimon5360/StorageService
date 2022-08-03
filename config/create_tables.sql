begin;

-- create schema
create schema if not exists barmapschema; 

-- table drinks
create table if not exists drinks (
    id serial primary key,
    title text not null,
    price int not null,
    type int not null,
    description text not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    ingredient_id int[]
);

--table items
create table if not exists ingredients (
    id serial primary key,
    title text not null,
    amount int not null,
    created_at timestamp not null,
    updated_at timestamp not null
);

commit;