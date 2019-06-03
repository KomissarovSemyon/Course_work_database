<template>
  <div class="container">
    <div
      v-for="cinema in schedule"
      :key="cinema.id">
      <div>
        <div class="col-md-5 col-lg-4">
          <h3>
            {{ cinema.name }}
          </h3>
          <h6 class="muted">
            {{ cinema.address }}
          </h6>
        </div>
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
import { movieSchedule } from '@/api'

export default {
  name: 'Movie',
  data: function () {
    return {
      movieID: 0,
      schedule: [],
      date: null
    }
  },

  created: function () {
    this.movieID = this.$route.params.id
    this.date = this.$route.params.date

    movieSchedule(77, this.movieID, this.date)
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
  }
}
</script>
