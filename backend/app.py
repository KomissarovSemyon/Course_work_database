from flask import Flask, request
from json import dumps
from flask_jsonpify import jsonify
import psycopg2
from datetime import datetime

pg_url = 'postgres://kino:antman_and_thanos@localhost/kino?sslmode=disable'
app = Flask(__name__)
conn = psycopg2.connect(pg_url)

app.config['JSON_AS_ASCII'] = False

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
    columns = ('id', 'ya_id', 'date', 'price_min', 'price_max', 'cinema_name', 'cinema_address')
    cur.execute("""
    SELECT
        s.session_id,
        s.ya_id,
        s.date,
        s.price_min,
        s.price_max,
        c.name,
        c.address
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

    result = {
        'sessions': [dict(zip(columns, i)) for i in cur.fetchall()],
        'date': date_str
    }

    return jsonify(result)


if __name__ == '__main__':
    app.run()
