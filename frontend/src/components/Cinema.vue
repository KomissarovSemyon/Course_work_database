<template>
  <div class="container">
    <div>
      <h3>
        {{ info.name }}
      </h3>
      <a :href="info.location_href">
        <h6 class="muted">
          {{ info.address }}
        </h6>
        <h6 class="small muted">
          {{ info.city_name }}
        </h6>
      </a>
    </div>
    <hr>
    <div
      v-for="movie in schedule"
      :key="movie.id">
      <div>
        <router-link
          :to="{
            name: 'Movie',
            params: {
              id: movie.id,
              date: date
            }
          }"
          class="col-md-5 col-lg-4"
        >
          <h3>
            {{ movie.name }}
          </h3>
        </router-link>
        <div class="d-flex flex-wrap justify-content-left">
          <a
            v-for="session in movie.sessions"
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
import { cinemaSchedule, cinemaInfo } from '@/api'

export default {
  name: 'Movie',
  data: function () {
    return {
      cinemaID: 0,
      schedule: [],
      date: null,
      info: {
        name: 'hih'
      }
    }
  },

  created: function () {
    this.cinemaID = this.$route.params.id
    this.date = this.$route.params.date

    cinemaSchedule(this.cinemaID, this.date)
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

    cinemaInfo(this.cinemaID)
      .then(response => {
        let data = response.data
        data['location_href'] = `https://maps.yandex.ru/?text=` + data['location'].slice(1, -1)
        this.info = data
      })
  }
}
</script>
