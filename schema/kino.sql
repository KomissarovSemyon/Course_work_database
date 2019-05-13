-- see pg_timezone_names
create table if not exists timezones
(
    timezone_id int         not null
        generated always as identity
        primary key,
    -- like Europe/Moscow
    name        varchar(40) not null unique,

    -- current offset, considering dst & stuff
    utc_offset  interval    not null
);

create unique index if not exists timezones_name_index
    on timezones (name);

create table if not exists countries
(
    -- alpha-2 country code
    country_code char(2)     not null
        primary key,
    -- pretty name in russian
    name_ru      varchar(50) not null,
    -- pretty name in english
    name_en      varchar(50) not null
);

create table if not exists cities
(
    city_id      int         not null
        generated always as identity
        primary key,
    country_code char(2)     not null references countries (country_code) on delete restrict,

    -- pretty name in russian
    name_ru      varchar(50) not null,

    -- pretty name in english
    name_en      varchar(50) not null,

    -- city "id" (name) used by yandex afisha
    ya_name      varchar(30),

    timezone_id  int         not null references timezones (timezone_id) on delete restrict
);

create table if not exists movies
(
    movie_id     int         not null
        generated always as identity
        primary key,

    -- Movie title in Russian
    title_ru     varchar(30) not null,

    -- Movie title in English
    title_en     varchar(30) not null,

    -- Release Year
    year         smallint,
    -- Duration in seconds
    duration     smallint    not null,
    -- Release Date in Russia
    release_ru   date,
    -- kinopoisk.ru ID
    kp_id        int,
    -- rating on kinopoisk.ru
    kp_rating    smallint,

    -- last successful kinopoisk.ru sync
    kp_last_sync timestamp,

    -- yandex afisha event id
    ya_event_id  varchar(30),

    -- last successful yandex afisha sync
    ya_last_sync timestamp,

    -- our rating: current rating
    rating       smallint,
    -- and number of votes
    rating_count int,

    country_code char(2)     references countries (country_code) on delete set null
);

create index if not exists movies_rating_index
    on movies (rating desc, kp_rating desc);

create table if not exists genres
(
    genre_id int         not null
        generated always as identity
        primary key,
    -- genre name in russian
    name_ru  varchar(20) not null,
    -- kinopoisk.ru genre id
    kp_id    smallint
);

create table if not exists movie_genres
(
    movie_id int references movies (movie_id) on delete cascade,
    genre_id int references genres (genre_id) on delete cascade,
    CONSTRAINT movie_genres_pk PRIMARY KEY (movie_id, genre_id)
);

create table if not exists cinemas
(
    cinema_id int         not null
        generated always as identity
        primary key,
    -- cinema id in russian
    name      varchar(30) not null,
    -- full address in russian
    address   varchar(80) not null,
    -- coordinates
    loc       point       not null,

    -- the city
    city_id   int         not null references cities (city_id) on delete restrict,

    -- place id on yandex afisha
    ya_id     varchar(30)
);

create table if not exists sessions
(
    session_id int       not null
        generated always as identity
        primary key,

    -- cinema specific hall name
    hall_name  varchar(20),

    cinema_id  int references cinemas (cinema_id) on delete cascade,
    movie_id   int references movies (movie_id) on delete cascade,

    -- session type enum
    type       int,

    -- datetime of event
    date       timestamp not null,

    -- price range in rub
    price_min  smallint,
    price_max  smallint,

    -- yandex afisha "ticket id"
    ya_id      varchar(50)
);

create index if not exists sessions_movie_index
    on sessions (movie_id, date desc);

create index if not exists sessions_cinema_index
    on sessions (cinema_id, date desc);

create table if not exists users
(
    user_id       int         not null
        generated always as identity
        primary key,

    email         varchar(30) not null,
    -- bcrypt
    password_hash char(60)    not null,

    city_id       int         not null references cities (city_id) on delete restrict
);

create unique index if not exists users_email_index
    on users (lower(email));

create table if not exists user_movie_ratings
(
    movie_id int references movies (movie_id) on delete cascade,
    user_id  int references users (user_id) on delete restrict,
    rating   smallint not null,
    CONSTRAINT user_movie_ratings_pk PRIMARY KEY (movie_id, user_id)
);

create index if not exists user_movie_ratings_user_index
    on user_movie_ratings (user_id);

create table if not exists user_favorite_cinemas
(
    cinema_id int references cinemas (cinema_id) on delete cascade,
    user_id   int references users (user_id) on delete cascade,
    CONSTRAINT user_favorite_cinemas_pk PRIMARY KEY (cinema_id, user_id)
);

create table if not exists user_starred_movies
(
    movie_id int references movies (movie_id) on delete cascade,
    user_id  int references users (user_id) on delete cascade,
    CONSTRAINT user_starred_movies_pk PRIMARY KEY (movie_id, user_id)
);

