import axios from 'axios'

const baseUrl = 'http://127.0.0.1:5000/'

const authStore = {
  get token () {
    return localStorage.authToken || null
  },

  set token (newTok) {
    localStorage.authToken = newTok
  },

  clear () {
    localStorage.removeItem('authToken')
  },

  get config () {
    return (this.token !== null)
      ? {
        headers: {'Authorization': 'Bearer ' + this.token}
      } : {}
  }
}

const handleAuth = function (response) {
  const data = response['data']
  if (data.ok) {
    authStore.token = data.access_token
  }
}

const get = function (path) {
  return axios
    .get(baseUrl + path, authStore.config)
    .catch(function (error) {
      if (error.response.status === 401) {
        authStore.clear()
      }
    })
}

const post = function (path, data) {
  return axios
    .post(baseUrl + path, data, authStore.config)
    .catch(function (error) {
      if (error.response.status === 401) {
        authStore.clear()
      }
    })
}

const movieSchedule = function (cityID, movieID, date) {
  let path = `api/movie_schedule/${cityID}/${movieID}`
  if (date) {
    path += '/' + date
  }

  return get(path)
}

const cinemaSchedule = function (cinemaID, date) {
  let path = `api/cinema_schedule/${cinemaID}`
  if (date) {
    path += '/' + date
  }

  return get(path)
}

const currentMovies = function (cityID, date) {
  let path = `api/current_movies/${cityID}`
  if (date) {
    path += '/' + date
  }

  return get(path)
}

const cinemaInfo = function (cinemaID) {
  let path = `api/cinema/${cinemaID}`
  return get(path)
}

const movieInfo = function (cinemaID) {
  let path = `api/movie/${cinemaID}`
  return get(path)
}

const signUp = function (email, password) {
  let path = `auth/register`

  return post(path, {
    'email': email,
    'password': password
  }).then(handleAuth)
}

const signIn = function (email, password) {
  let path = `auth/login`

  return post(path, {
    'email': email,
    'password': password
  }).then(handleAuth)
}

let meCached = null

const signOut = function () {
  meCached = null
  return authStore.clear()
}

const setCity = function (cityID) {
  const path = `auth/set_city`
  return post(path, {
    city_id: cityID
  })
}

const getMe = function (noCache = false) {
  return new Promise((resolve, reject) => {
    if (meCached !== null && !noCache) {
      resolve(meCached)
      return
    }

    let path = `auth/me`

    get(path).then(resp => {
      if (resp && resp.data) {
        meCached = resp.data
        resolve(meCached)
      } else {
        reject(new Error('cant get myself'))
      }
    })
  })
}

let citiesCached = null

const cities = function (noCache = false) {
  return new Promise((resolve, reject) => {
    if (citiesCached !== null && !noCache) {
      resolve(citiesCached)
      return
    }

    let path = `api/cities`

    get(path).then(resp => {
      if (resp && resp.data) {
        citiesCached = {}
        resp.data.forEach(c => { citiesCached[c.id] = c })
        resolve(citiesCached)
      } else {
        reject(new Error('cant get myself'))
      }
    })
  })
}

const starMovie = function (movieID, starred) {
  const path = `api/star_movie/${movieID}`

  return post(path, {
    star: starred
  })
}

const starCinema = function (cinemaID, starred) {
  const path = `api/favorite_cinema/${cinemaID}`

  return post(path, {
    favorite: starred
  })
}

export {
  movieSchedule,
  cinemaSchedule,

  currentMovies,
  cinemaInfo,
  movieInfo,

  get,
  post,

  signUp,
  signIn,
  signOut,
  getMe,
  setCity,

  cities,

  starMovie,
  starCinema
}
