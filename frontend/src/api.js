import axios from 'axios'

const apiBase = 'http://127.0.0.1:5000/api/'

const get = function (path) {
  return axios
    .get(apiBase + path)
    .catch(function (error) {
      console.log(error)
    })
}

const movieSchedule = function (cityID, movieID, date) {
  let path = `movie_schedule/${cityID}/${movieID}`
  if (date) {
    path += '/' + date
  }

  return get(path)
}

const cinemaSchedule = function (cinemaID, date) {
  let path = `cinema_schedule/${cinemaID}`
  if (date) {
    path += '/' + date
  }

  return get(path)
}

const currentMovies = function (cityID, date) {
  let path = `current_movies/${cityID}`
  if (date) {
    path += '/' + date
  }

  return get(path)
}

export {
  movieSchedule,
  cinemaSchedule,
  currentMovies,
  get
}
