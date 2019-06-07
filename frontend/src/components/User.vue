<template>
  <div class="container">
    <div>
      <h3>
        {{ email }}
      </h3>
      <form>
        <select
          v-model="city_id"
          class="form-control"
          @change="setCity">
          <option
            v-for="city in cities"
            :label="city.name"
            :key="city.id">
            {{ city.id }}
          </option>
        </select>
      </form>
    </div>
    <hr>

    <h3
      v-if="favorite_movies.length > 0">
      Любимые фильмы
    </h3>
    <div
      v-for="movie in favorite_movies"
      :key="movie.id">
      <div>
        <router-link
          :to="{
            name: 'Movie',
            params: {
              id: movie.id,
              city_id: city_id
            }
          }"
          class="col-md-5 col-lg-4"
        >
          <h3>
            {{ movie.title_ru }}
          </h3>
        </router-link>
      </div>
      <hr>
    </div>

    <h3
      v-if="favorite_cinemas.length">
      Любимые кинотеатры
    </h3>
    <div
      v-for="cinema in favorite_cinemas"
      :key="cinema.id">
      <div>

        <router-link
          :to="{
            name: 'Cinema',
            params: {
              id: cinema.id,
            }
          }"
          class="col-md-5 col-lg-4"
        >
          <h3>
            {{ cinema.name }}
          </h3>
          <h6 class="muted">
            {{ cinema.address }}
          </h6>
        </router-link>
        <div class="d-flex flex-wrap justify-content-left">
          <a
            v-for="session in cinema.sessions"
            :key="session.id"
            :href="session.ticket_url"
            :class="'m-2 btn btn-' + (session.ticket_url ? 'primary' : 'outline-secondary')">
            <div>
              {{ ((session.date.getHours() > 10) ? '' : '0') + session.date.toLocaleTimeString('ru-ru', {
                hour: 'numeric', minute: '2-digit',
              }) }}
            </div>
            <div
              v-if="session.price_min"
              class="small">
              {{ session.price_min }}&nbsp;₽
            </div>
          </a>
        </div>
      </div>
      <hr>
    </div>
  </div>
</template>

<script>
import { cities, getMe, setCity } from '@/api'

export default {
  name: 'User',

  data: function () {
    return {
      email: '',
      city_id: 0,
      favorite_cinemas: [],
      favorite_movies: [],

      cities: {}
    }
  },

  created: async function () {
    this.cities = await cities()

    let me = await getMe()

    this.email = me.email
    this.city_id = me.city_id
    this.favorite_cinemas = me.favorite_cinemas
    this.favorite_movies = me.favorite_movies
  },

  methods: {
    setCity: function () {
      setCity(this.city_id)
    }
  }
}
</script>
