from flask import Flask, request
from flask_restful import Resource, Api
from sqlalchemy import create_engine
from json import dumps
from flask_jsonpify import jsonify


db_connect = create_engine('postgres://kino:antman_and_thanos@localhost/kino?sslmode=disable')
app = Flask(__name__)
api = Api(app)


class CurrentMovies(Resource):
    def get(self, city_id):
        columns = ['id', 'title', 'rating', 'session_count']
        conn = db_connect.connect()
        query = conn.execute("SELECT DISTINCT MAX(m.movie_id), MAX(m.title_ru), MAX(m.kp_rating), COUNT(s.session_id)\
                              FROM sessions s\
                              JOIN movies m on m.movie_id = s.movie_id\
                              JOIN cinemas c on s.cinema_id = c.cinema_id\
                              WHERE DATE(s.date) = '2019-05-17' AND\
                                    c.city_id = %d\
                              GROUP BY s.movie_id\
                              ORDER BY COUNT(s.session_id) DESC" % int(city_id))
        result = {'data': [dict(zip(columns, i)) for i in query.cursor.fetchall()]}
        return jsonify(result)


api.add_resource(CurrentMovies, '/current_movies/<city_id>')

if __name__ == '__main__':
    app.run()
