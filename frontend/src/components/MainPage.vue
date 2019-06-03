<template>
  <div class="main-page">
    <div class="container">
      <h1>
        Кино в {{ city }} ({{ date }})
      </h1>
      <div class="row justify-content-center">
        <div
          v-for="movie in movies"
          :key="movie.movie_id"
          class="card col-md-4 col-lg-3 m-3 box-shadow">
          <!-- <img class="card-img-top" data-src="holder.js/100px225?theme=thumb&amp;bg=55595c&amp;fg=eceeef&amp;text=Thumbnail" alt="Thumbnail [100%x225]" src="data:image/svg+xml;charset=UTF-8,%3Csvg%20width%3D%22348%22%20height%3D%22225%22%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20viewBox%3D%220%200%20348%20225%22%20preserveAspectRatio%3D%22none%22%3E%3Cdefs%3E%3Cstyle%20type%3D%22text%2Fcss%22%3E%23holder_16ac0a6933e%20text%20%7B%20fill%3A%23eceeef%3Bfont-weight%3Abold%3Bfont-family%3AArial%2C%20Helvetica%2C%20Open%20Sans%2C%20sans-serif%2C%20monospace%3Bfont-size%3A17pt%20%7D%20%3C%2Fstyle%3E%3C%2Fdefs%3E%3Cg%20id%3D%22holder_16ac0a6933e%22%3E%3Crect%20width%3D%22348%22%20height%3D%22225%22%20fill%3D%22%2355595c%22%3E%3C%2Frect%3E%3Cg%3E%3Ctext%20x%3D%22117.125%22%20y%3D%22120.046875%22%3EThumbnail%3C%2Ftext%3E%3C%2Fg%3E%3C%2Fg%3E%3C%2Fsvg%3E" data-holder-rendered="true" style="height: 225px; width: 100%; display: block;"> -->
          <div class="card-body d-flex flex-column justify-content-between">
            <router-link
              :to="{
                name: 'Movie',
                params: {
                  id: movie.movie_id,
                  date: date
                }
              }"
              class="card-link"
            />
            <div>
              <h5 class="card-title mb-3">{{ movie.title }}</h5>
              <h6 class="card-subtitle mb-3 text-muted">
                {{ movie.session_count }} сеансов
                <span v-if="movie.min_price"> от {{ movie.min_price }} ₽ </span>
              </h6>
            </div>

            <div class="d-flex align-items-center justify-content-between">
              <div class="btn btn-primary">
                Билеты
              </div>

              <span class="text-muted ml-3">
                <span v-if="movie.rating">
                  {{ movie.rating / 100 }} 🌟
                </span>
              </span>
            </div>

          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { currentMovies } from '@/api'

export default {
  name: 'MainPage',
  data: function () {
    return {
      movies: [],
      city: 77,
      date: null
    }
  },
  created: function () {
    this.date = this.$route.params.date

    currentMovies(this.city, this.date)
      .then(response => {
        this.movies = response.data['movies']
      })
  }
}
</script>

<style>
.card:hover {
  transition: .15s box-shadow ease,.15s transform ease;
  transform: translateY(-0.25rem);
  box-shadow: 0 2.25rem 1.5rem -1.5rem rgba(33,37,41,.3),0 0 1.5rem .5rem rgba(33,37,41,.05)!important;
}
</style>
