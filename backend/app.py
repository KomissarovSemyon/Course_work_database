from flask import Flask, request, jsonify
from flask_jwt_extended import (
    JWTManager, jwt_required, jwt_optional,
    create_access_token, get_jwt_identity
)
from flask_bcrypt import Bcrypt
import psycopg2
from datetime import datetime
from flask_cors import CORS
import base64
import jwt
import sys
import os.path


def install_secret_key(app, filename='secret_key'):
    """Configure the SECRET_KEY from a file
    in the instance directory.

    If the file does not exist, print instructions
    to create it from a shell with a random key,
    then exit.
    """
    filename = os.path.join(app.instance_path, filename)
    try:
        app.config['SECRET_KEY'] = open(filename, 'rb').read()
    except IOError:
        print('Error: No secret key. Create it with:')
        if not os.path.isdir(os.path.dirname(filename)):
            print('mkdir -p', os.path.dirname(filename))
        print('head -c 24 /dev/urandom >', filename)
        sys.exit(1)


pg_url = 'postgres://kino:antman_and_thanos@localhost/kino?sslmode=disable'
app = Flask(__name__)
install_secret_key(app)
conn = psycopg2.connect(pg_url)

app.config['JSON_AS_ASCII'] = False
CORS(app)
bcrypt = Bcrypt(app)
jwt = JWTManager(app)


@app.route('/api/current_movies/<city_id>')
@app.route('/api/current_movies/<city_id>/<date_str>')
def get_current_movies(city_id, date_str=None):
    if date_str is None:
        date_str = datetime.today().strftime('%Y-%m-%d')

    cur = conn.cursor()
    columns = ('movie_id', 'title', 'rating', 'session_count', 'min_price')
    cur.execute("""
    SELECT DISTINCT
        MAX(m.movie_id),
        MAX(m.title_ru),
        MAX(m.kp_rating),
        COUNT(s.session_id),
        MIN(NULLIF(s.price_min, 0))
    FROM sessions s
    JOIN movies m ON m.movie_id = s.movie_id
    JOIN cinemas c ON s.cinema_id = c.cinema_id
    WHERE s.date::date = %(date)s
        AND c.city_id = %(city_id)s
    GROUP BY s.movie_id
    ORDER BY COUNT(s.session_id) DESC
    """, {
        'date': date_str,
        'city_id': int(city_id)
    })

    result = {
        'movies': [dict(zip(columns, i)) for i in cur.fetchall()],
        'date': date_str,
    }

    cur.close()

    return jsonify(result)


@app.route('/api/movie_schedule/<city_id>/<movie_id>')
@app.route('/api/movie_schedule/<city_id>/<movie_id>/<date_str>')
@jwt_optional
def get_movie_schedule(city_id, movie_id, date_str=None):
    if date_str is None:
        date_str = datetime.today().strftime('%Y-%m-%d')

    cur = conn.cursor()
    cur.execute("""
    SELECT
        s.session_id,
        s.ya_id,
        s.date,
        s.price_min,
        s.price_max,
        c.cinema_id,
        c.name,
        c.address,
        s.hall_name,
        (ufc.cinema_id is not null) as is_favorite
    FROM sessions s
    JOIN cinemas c on s.cinema_id = c.cinema_id
    LEFT JOIN user_favorite_cinemas ufc on ufc.user_id = %(user_id)s
        AND ufc.cinema_id = s.cinema_id
    WHERE s.date::date = %(date)s AND
        s.movie_id = %(movie_id)s AND
        c.city_id = %(city_id)s
    ORDER BY is_favorite desc
    """, {
        'date': date_str,
        'movie_id': movie_id,
        'city_id': city_id,
        'user_id': get_jwt_identity(),
    })

    cinemas = list()
    cinemap = dict()
    for row in cur.fetchall():
        session_id = row[0]
        ya_id = row[1]
        date = row[2]
        price_min = row[3]
        price_max = row[4]
        cinema_id = row[5]
        cinema_name = row[6]
        cinema_address = row[7]
        hall_name = row[8]
        is_favorite = row[9]

        if cinema_id not in cinemap:
            d = {
                'id': cinema_id,
                'name': cinema_name,
                'address': cinema_address,
                'is_favorite': is_favorite,
                'sessions': list()
            }
            cinemas.append(d)
            cinemap[cinema_id] = d

        ticket_url = None
        if ya_id:
            ya_id = ya_id.encode('utf-8').strip()
            ya_id = base64.b64encode(ya_id).decode('ascii')
            ticket_url = 'http://widget.afisha.yandex.ru/w/sessions/' + ya_id

        sesslist = cinemap[cinema_id]['sessions']
        sesslist.append({
            'id': session_id,
            'ticket_url': ticket_url,
            'date': date.strftime('%Y-%m-%dT%H:%M:%SZ'),
            'price_min': price_min,
            'price_max': price_max,
            'hall': hall_name,
        })

    cur.close()

    return jsonify({
        'schedule': cinemas,
        'date': date_str,
    })


@app.route('/api/cinema_schedule/<cinema_id>')
@app.route('/api/cinema_schedule/<cinema_id>/<date_str>')
@jwt_optional
def get_cinema_schedule(cinema_id, date_str=None):
    if date_str is None:
        date_str = datetime.today().strftime('%Y-%m-%d')
    
    user_id = get_jwt_identity()
    cur = conn.cursor()

    cur.execute("""
    SELECT
        s.session_id,
        s.ya_id,
        s.date,
        s.price_min,
        s.price_max,
        m.movie_id,
        m.title_ru,
        s.hall_name,
        (usm.movie_id is not null) as is_starred
    FROM sessions s
    JOIN cinemas c on s.cinema_id = c.cinema_id
    JOIN movies m on s.movie_id = m.movie_id
    LEFT JOIN user_starred_movies usm on usm.user_id = %(user_id)s
        AND usm.movie_id = s.movie_id
    WHERE s.date::date = %(date)s AND
        c.cinema_id = %(cinema_id)s
    
    ORDER BY is_starred desc,
        -- (c.loc <@> point(0, 0)) asc,
        c.name asc,
        s.type desc,
        s.date asc
    """, {
        'date': date_str,
        'cinema_id': cinema_id,
        'user_id': user_id,
    })

    movies = list()
    moviemap = dict()
    for row in cur.fetchall():
        session_id = row[0]
        ya_id = row[1]
        date = row[2]
        price_min = row[3]
        price_max = row[4]
        movie_id = row[5]
        movie_name = row[6]
        hall_name = row[7]
        is_starred = row[8]

        if movie_id not in moviemap:
            d = {
                'id': movie_id,
                'name': movie_name,
                'is_starred': is_starred,
                'sessions': list()
            }
            movies.append(d)
            moviemap[movie_id] = d

        ticket_url = None
        if ya_id:
            ya_id = ya_id.encode('utf-8').strip()
            ya_id = base64.b64encode(ya_id).decode('ascii')
            ticket_url = 'http://widget.afisha.yandex.ru/w/sessions/' + ya_id

        sesslist = moviemap[movie_id]['sessions']
        sesslist.append({
            'id': session_id,
            'ticket_url': ticket_url,
            'date': date.strftime('%Y-%m-%dT%H:%M:%SZ'),
            'price_min': price_min,
            'price_max': price_max,
            'hall': hall_name,
        })

    cur.close()

    return jsonify({
        'schedule': movies,
        'date': date_str,
    })

@app.route('/api/top_cinemas/<city_id>')
@app.route('/api/top_cinemas/<city_id>/<date_str>')
def get_top_cinemas(city_id, date_str=None):
    if date_str is None:
        date_str = datetime.today().strftime('%Y-%m-%d')

    cur = conn.cursor()
    columns = ('id', 'name', 'address', 'count')
    cur.execute("""
    SELECT
        c.cinema_id,
        c.name,
        c.address,
        COUNT(s.session_id)
    FROM sessions s
    JOIN cinemas c on s.cinema_id = c.cinema_id
    WHERE s.date::date = %(date)s AND
        c.city_id = %(city_id)s
    GROUP BY c.cinema_id
    ORDER BY COUNT(s.session_id) DESC
    """, {
        'city_id': city_id,
        'date': date_str
    })

    result = {
        'cinemas': [dict(zip(columns, i)) for i in cur.fetchall()],
        'date': date_str,
        'city_id': city_id
    }

    cur.close()

    return jsonify(result)


@app.route('/api/movie/<movie_id>')
@jwt_optional
def get_movie_info(movie_id):
    cur = conn.cursor()
    columns = (
        'title_ru', 'title_or', 'year', 'duration', 'release',
        'kp_id', 'kp_rating', 'rating', 'rating_count',
        'country_name_ru', 'country_name_en', 'is_starred'
    )
    cur.execute("""
    SELECT
        m.title_ru,
        m.title_or,
        m.year,
        m.duration,
        m.release,
        m.kp_id,
        m.kp_rating,
        m.rating,
        m.rating_count,
        c.name_ru,
        c.name_en,
        (usm.movie_id is not null) as is_starred
    FROM movies m
    LEFT JOIN countries c on m.country_code = c.country_code
    LEFT JOIN user_starred_movies usm on usm.movie_id = m.movie_id AND usm.user_id = %(user_id)s
    WHERE m.movie_id = %(movie_id)s
    """, {
        'movie_id': movie_id,
        'user_id': get_jwt_identity(),
    })

    result = dict(zip(columns, cur.fetchone()))
    result['kp_link'] = 'https://www.kinopoisk.ru/film/' + str(result['kp_id'])
    result['movie_id'] = int(movie_id)
    del result['kp_id']

    cur.close()

    return jsonify(result)


@app.route('/api/cinema/<cinema_id>')
@jwt_optional
def get_cinema_info(cinema_id):
    cur = conn.cursor()
    columns = ('name', 'address', 'location', 'city_name', 'is_favorite')
    cur.execute("""
    SELECT c.name,
        c.address,
        c.loc,
        ct.name,
        (ufc.cinema_id is not null) as is_favorite
    FROM cinemas c
    JOIN cities ct on c.city_id = ct.city_id
    LEFT JOIN user_favorite_cinemas ufc ON c.cinema_id = ufc.cinema_id AND ufc.user_id = %(user_id)s
    WHERE c.cinema_id = %(cinema_id)s
    """, {
        'cinema_id': cinema_id,
        'user_id': get_jwt_identity()
    })

    result = dict(zip(columns, cur.fetchone()))
    result['cinema_id'] = int(cinema_id)

    cur.close()

    return jsonify(result)


@app.route('/auth/register', methods=['POST'])
def register():
    if not request.is_json:
        return jsonify({"msg": "Missing JSON in request"}), 400
    
    email = request.json['email']
    password = request.json['password']
    city_id = request.json.get('city_id')

    cur = conn.cursor()
    pass_hash = bcrypt.generate_password_hash(password).decode('ascii')

    try:
        cur.execute("""
        INSERT into users (email, password_hash, city_id)
        VALUES (%(email)s, %(pass_hash)s, %(city_id)s)
        RETURNING user_id
        """, {
            'email': email,
            'pass_hash': pass_hash,
            'city_id': city_id,
        })
        user_id = cur.fetchone()[0]
        result = {
            'ok': True,
            'access_token': create_access_token(identity=user_id)
        }
        conn.commit()
    except psycopg2.errors.UniqueViolation:
        # This email already in database
        result = {'ok': False}
        conn.rollback()

    cur.close()
    return jsonify(result)

@app.route('/api/star_movie/<movie_id>', methods=['POST'])
@jwt_required
def star_movie(movie_id):
    if not request.is_json:
        return jsonify({"msg": "Missing JSON in request"}), 400
    
    star = request.json.get('star')

    cur = conn.cursor()
    if star:
        code = """
        INSERT into user_starred_movies (user_id, movie_id)
        VALUES (%(user_id)s, %(movie_id)s)
        ON CONFLICT ON CONSTRAINT user_starred_movies_pk DO NOTHING
        """
    else:
        code = """
        DELETE FROM user_starred_movies
        WHERE user_id = %(user_id)s AND movie_id = %(movie_id)s
        """

    cur.execute(code, {
        'user_id': get_jwt_identity(),
        'movie_id': movie_id,
    })
    conn.commit()
    cur.close()

    return jsonify({'star': star})

@app.route('/api/favorite_cinema/<cinema_id>', methods=['POST'])
@jwt_required
def favorite_cinema(cinema_id):
    if not request.is_json:
        return jsonify({"msg": "Missing JSON in request"}), 400
    
    fav = request.json.get('favorite')

    cur = conn.cursor()
    if fav:
        code = """
        INSERT into user_favorite_cinemas (user_id, cinema_id)
        VALUES (%(user_id)s, %(cinema_id)s)
        ON CONFLICT ON CONSTRAINT user_favorite_cinemas_pk DO NOTHING
        """
    else:
        code = """
        DELETE FROM user_favorite_cinemas
        WHERE user_id = %(user_id)s AND cinema_id = %(cinema_id)s
        """

    cur.execute(code, {
        'user_id': get_jwt_identity(),
        'cinema_id': cinema_id,
    })
    conn.commit()
    cur.close()

    return jsonify({'favorite': fav})

@app.route('/auth/login', methods=['POST'])
def login():
    if not request.is_json:
        return jsonify({"msg": "Missing JSON in request"}), 400

    email = request.json['email']
    password = request.json['password']

    cur = conn.cursor()

    cur.execute("""
    SELECT u.user_id, u.password_hash
    FROM users u
    WHERE u.email = %(email)s
    """, {
        'email': email
    })

    fetched = cur.fetchone()
    cur.close()
    if fetched is None:
        return jsonify({
            'ok': False,
            'status': 'email not registered',
        })

    user_id, password_hash = fetched

    if bcrypt.check_password_hash(password_hash, password):
        result = {
            'ok': True,
            'access_token': create_access_token(identity=user_id)
        }
    else:
        result = {
            'ok': False
        }

    return jsonify(result)

@app.route('/auth/me')
@jwt_required
def me():
    user_id = int(get_jwt_identity())

    cur = conn.cursor()
    cur.execute("""
    SELECT u.email, u.city_id
    FROM users u
    WHERE u.user_id = %(user_id)s
    """, {
        'user_id': user_id
    })
    email, city_id = cur.fetchone()
    cur.close()

    cur = conn.cursor()
    cur.execute("""
    SELECT u.email, u.city_id
    FROM users u
    WHERE u.user_id = %(user_id)s
    """, {
        'user_id': user_id,
    })
    email, city_id = cur.fetchone()
    cur.close()

    movie_columns = (
        'title_ru', 'title_or','year',
        'kp_rating', 'rating',
    )
    cur = conn.cursor()
    cur.execute("""
    SELECT
        m.title_ru,
        m.title_or,
        m.year,
        m.kp_rating,
        m.rating
    FROM user_starred_movies usm
    JOIN movies m on usm.movie_id = m.movie_id
    WHERE usm.user_id = %(user_id)s
    """, {
        'user_id': user_id,
    })
    fav_movies = [dict(zip(movie_columns, v)) for v in cur.fetchall()]
    cur.close()

    cinema_columns = (
        'name', 'address', 'location', 'city_name'
    )
    cur = conn.cursor()
    cur.execute("""
    SELECT c.name,
        c.address,
        c.loc,
        ct.name
    FROM user_favorite_cinemas ufc
    LEFT JOIN cinemas c ON c.cinema_id = ufc.cinema_id
    JOIN cities ct on c.city_id = ct.city_id
    WHERE ufc.user_id = %(user_id)s
    """, {
        'user_id': user_id,
    })
    fav_cinemas = [dict(zip(cinema_columns, v)) for v in cur.fetchall()]
    cur.close()

    result = {
        'email': email,
        'city_id': city_id,
        'favorite_cinemas': fav_cinemas,
        'favorite_movies': fav_movies,
    }

    return jsonify(result)


if __name__ == '__main__':
    app.run()
