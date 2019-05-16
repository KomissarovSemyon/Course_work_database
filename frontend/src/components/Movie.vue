<template>
  <div class="container">
    <div v-for="cinema in schedule" v-bind:key="cinema.id">
      <div class="lead">
        {{cinema.name}}
      </div>
      <div class="lead">
        {{cinema.address}}
      </div>

      <div class="row" v-for="session in cinema.sessions" v-bind:key="session.id">
        <pre>
          {{JSON.stringify(session, null, 2)}}
        </pre>
      </div>
      <hr>
    </div>
  </div>
</template>

<script>
import {movieSchedule} from '@/api'

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
