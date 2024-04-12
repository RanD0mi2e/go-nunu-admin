CREATE TABLE tree_data (
    id INTEGER PRIMARY KEY,
    label TEXT NOT NULL,
    level INTEGER NOT NULL,
    parent_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    roles TEXT
);

CREATE TABLE user_roles (
    user_id INTEGER,
    tree_data_id INTEGER,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(tree_data_id) REFERENCES tree_data(id)
);

-- auto-generated definition
create table users
(
    id         integer
        primary key autoincrement,
    user_id    text not null
        constraint uni_users_user_id
            unique,
    nickname   text not null,
    password   text not null,
    email      text not null,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime
);

create index idx_users_deleted_at
    on users (deleted_at);

