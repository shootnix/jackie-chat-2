create table telegram_bots (
    id serial primary key,
    name text not null unique,
    token text not null unique,
    is_active boolean,
    info json
);

create table users (
    id serial primary key,
    name text not null unique,
    password text not null,
    default_bot integer references telegram_bots(id),
    is_active boolean default true,
    is_deleted boolean default false
);

create table messages (
    id serial primary key,
    message text not null,
    parse_mode text not null default 'html',
    chat_id integer not null,
    ctime timestamp not null default now(),
    is_success integer not null default -1,
    err text,
    user_id integer references users(id) on delete cascade,
    bot_id integer references telegram_bots(id) on delete set null
);

create table journal (
    id serial primary key,
    worker_name text not null,
    ctime timestamp not null default now(),
    message_id integer references messages(id) on delete cascade
);