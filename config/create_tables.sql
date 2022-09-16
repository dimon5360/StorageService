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
    bar_id int not null references bars(id) on delete cascade,
    ingredients_id bigint[],
    created_at timestamp not null,
    updated_at timestamp not null,
    on update cascade
);

--table ingredients
create table if not exists ingredients (
    id serial primary key,
    title text not null,
    amount int not null,
    drink_id int not null references drinks(id) on delete cascade,
    created_at timestamp not null,
    updated_at timestamp not null,
    on update cascade
);

-- update bars set title = '%s', address = '%s', description = '%s', drinks_id = '%s', "updated_at = '%s' where id = %s;
-- update drinks set title = '%s', price = '%s', type = '%s', description = '%s', "bar_id = '%s', updated_at = '%s' where id = %s;
-- update ingredients set title = '%s', amount = '%s', drink_id = '%s', updated_at = '%s' where id = %s;
-- returning *;"

commit;