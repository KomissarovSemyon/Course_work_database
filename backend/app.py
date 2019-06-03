from flask import Flask, request
from json import dumps
from flask_jsonpify import jsonify
import psycopg2
from datetime import datetime
from flask_cors import CORS
import base64
import hashlib

pg_url = 'postgres://kino:antman_and_thanos@localhost/kino?sslmode=disable'
app = Flask(__name__)
conn = psycopg2.connect(pg_url)

app.config['JSON_AS_ASCII'] = False
CORS(app)


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
    WHERE DATE(s.date) = %(date)s
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
        s.hall_name
    FROM sessions s
    JOIN cinemas c on s.cinema_id = c.cinema_id
    WHERE DATE(s.date) = %(date)s AND
        s.movie_id = %(movie_id)s AND
        c.city_id = %(city_id)s
    """, {
        'date': date_str,
        'movie_id': movie_id,
        'city_id': city_id
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

        if cinema_id not in cinemap:
            d = {
                'id': cinema_id,
                'name': cinema_name,
                'address': cinema_address,
                'sessions': list()
            }
            cinemas.append(d)
            cinemap[cinema_id] = d

        ticket_url = None
        if ya_id:
            ya_id = ya_id.encode('utf-8')
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
def get_cinema_schedule(cinema_id, date_str=None):
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
        m.movie_id,
        m.title_ru,
        s.hall_name
    FROM sessions s
    JOIN cinemas c on s.cinema_id = c.cinema_id
    JOIN movies m on s.movie_id = m.movie_id
    WHERE DATE(s.date) = %(date)s AND
        c.cinema_id = %(cinema_id)s
    """, {
        'date': date_str,
        'cinema_id': cinema_id
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

        if movie_id not in moviemap:
            d = {
                'id': movie_id,
                'name': movie_name,
                'sessions': list()
            }
            movies.append(d)
            moviemap[movie_id] = d

        ticket_url = None
        if ya_id:
            ya_id = ya_id.encode('utf-8')
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
    WHERE DATE(s.date) = %(date)s AND
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
def get_movie_info(movie_id):
    cur = conn.cursor()
    columns = (
        'title_ru', 'title_or', 'year', 'duration', 'release',
        'kp_id', 'kp_rating', 'rating', 'rating_count',
        'country_name_ru', 'country_name_en'
    )
    cur.execute("""
    SELECT m.title_ru,
           m.title_or,
           m.year,
           m.duration,
           m.release,
           m.kp_id,
           m.kp_rating,
           m.rating,
           m.rating_count,
           c.name_ru,
           c.name_en
    FROM movies m
    LEFT JOIN countries c on m.country_code = c.country_code
    WHERE movie_id = %(movie_id)s
    """, {
        'movie_id': movie_id
    })

    result = dict(zip(columns, cur.fetchone()))
    result['kp_link'] = 'https://www.kinopoisk.ru/film/' + str(result['kp_id'])
    result['movie_id'] = int(movie_id)
    del result['kp_id']

    cur.close()

    return jsonify(result)


@app.route('/api/cinema/<cinema_id>')
def get_cinema_info(cinema_id):
    cur = conn.cursor()
    columns = ('name', 'address', 'location', 'city_name')
    cur.execute("""
    SELECT c.name,
        c.address,
        c.loc,
        ct.name
    FROM cinemas c
    JOIN cities ct on c.city_id = ct.city_id
    WHERE c.cinema_id = %(cinema_id)s
    """, {
        'cinema_id': cinema_id
    })

    result = dict(zip(columns, cur.fetchone()))
    result['cinema_id'] = int(cinema_id)

    cur.close()

    return jsonify(result)


@app.route('/api/register', methods=['POST'])
def register():
    email = request.form['email']
    city_id = request.form['city_id']
    password = request.form['password']

    cur = conn.cursor()
    pass_hash = hashlib.md5(password.encode('utf-8')).hexdigest()

    try:
        cur.execute("""
        INSERT into users (email, password_hash, city_id)
        VALUES (%(email)s, %(pass_hash)s, %(city_id)s)
        """, {
            'email': email,
            'pass_hash': pass_hash,
            'city_id': city_id,
        })
        status_code = 200
    except psycopg2.errors.UniqueViolation:
        # This email already in database
        status_code = 300

    cur.close()
    return jsonify({'status': status_code})


if __name__ == '__main__':
    app.run()
