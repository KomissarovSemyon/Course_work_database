<template>
  <div class="container">
    <div>
      <h3>
        {{ info.title_ru }}
      </h3>
      <h6 class="text-muted">
        {{ info.title_or }}
        <span :v-if="info.year">
          ({{ info.year }})
        </span>
      </h6>
      <h6
        :v-if="info.duration"
        class="text-muted small">
        {{ info.duration }} мин
      </h6>
      <div>
        <button
          v-if="info.is_starred"
          class="btn btn-primary"
          @click="star(false)">
          💘
        </button>
        <button
          v-else
          class="btn btn-primary"
          @click="star(true)">
          ❤️
        </button>
        <a
          v-if="info.kp_rating"
          :href="info.kp_link"
          class="btn btn-yellow">
          <strong>
            {{ info.kp_rating / 100 }}
          </strong>
        </a>
      </div>
    </div>
    <hr>
    <div
      v-for="cinema in schedule"
      :key="cinema.id">
      <div>

        <router-link
          :to="{
            name: 'Cinema',
            params: {
              id: cinema.id,
              date: date
            }
          }"
          class="col-md-5 col-lg-4"
        >
          <h3>
            {{ cinema.is_favorite ? '❤️' : '' }}
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

<style>
.btn-yellow {
  background-color: #ffdb4d;
}

.btn-yellow:hover {
  color: red;
}
</style>

<script>
import { movieSchedule, movieInfo, starMovie } from '@/api'

export default {
  name: 'Movie',
  data: function () {
    return {
      movieID: 0,
      cityID: 0,
      schedule: [],
      date: null,
      info: {}
    }
  },

  created: function () {
    this.movieID = this.$route.params.id
    this.cityID = this.$route.params.city_id
    this.date = this.$route.params.date

    movieSchedule(this.cityID, this.movieID, this.date)
      .then(response => {
        this.schedule = response.data['schedule']
          .map(c => {
            c.sessions = c.sessions.map(s => {
              s.date = new Date(s.date)
              return s
            }).sort((a, b) => a.date - b.date)
            return c
          })
      })

    movieInfo(this.movieID)
      .then(response => {
        let data = response.data
        this.info = data
      })
  },

  methods: {
    star: async function (newV) {
      this.info.is_starred = (await starMovie(this.movieID, newV)).data.star
    }
  }
}
</script>
