<template>
  <div id="app">
    <header
      class="d-flex flex-row align-items-center justify-content-between p-3 px-md-4 mb-3 bg-white border-bottom box-shadow"
    >
      <h5 class="my-0 mr-md-auto font-weight-normal">
        <router-link
          id="link-main"
          :to="{ name: 'MainPage'}">
          Кукаремуви
        </router-link>
      </h5>

      <router-link
        v-if="email"
        :to="{ name: 'User' }"
        class="mr-2 btn btn-outline-primary"
      >
        {{ email }}
      </router-link>

      <a
        v-if="email"
        class="btn btn-outline-danger"
        href="#"
        @click="signOut">
        Выйти
      </a>
      <router-link
        v-else
        :to="{ name: 'Auth' }"
        class="btn btn-outline-primary"
      >
        Войти
      </router-link>
    </header>
    <keep-alive include="MainPage">
      <router-view/>
    </keep-alive>
  </div>
</template>

<script>
import { getMe, signOut } from '@/api'

export default {
  name: 'App',
  data: function () {
    return {
      email: '',
      signOut: () => {
        signOut()
        this.email = null
      }
    }
  },

  created: function () {
    getMe().then(me => { this.email = me.email }).catch(() => 0)
  }
}
</script>

<style>
#app {
  font-family: Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

.box-shadow {
  box-shadow: 0 .25rem .75rem rgba(0, 0, 0, .05);
}

#link-main, #link-main:hover, #link-main:focus {
    text-decoration: none;
}

.card-link {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 1;
  cursor: pointer;
}
</style>
