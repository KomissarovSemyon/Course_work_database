<template>
  <div class="container">
    <div
      v-for="cinema in schedule"
      :key="cinema.id">
      <h3>
        {{ cinema.name }}
      </h3>
      <h6 class="muted">
        {{ cinema.address }}
      </h6>

      <div
        v-for="session in cinema.sessions"
        :key="session.id"
        class="row">
        <pre>
          {{ JSON.stringify(session, null, 2) }}
        </pre>
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
      })
  }
}
</script>
