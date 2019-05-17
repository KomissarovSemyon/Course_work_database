import Vue from 'vue'
import Router from 'vue-router'
import MainPage from '@/components/MainPage'
import Movie from '@/components/Movie'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/:date?',
      name: 'MainPage',
      component: MainPage
    },
    {
      path: '/movie/:id/:date?',
      name: 'Movie',
      component: Movie
    }
  ]
})
