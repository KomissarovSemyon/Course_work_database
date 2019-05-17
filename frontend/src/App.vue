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

      <a
        v-if="userID"
        class="btn btn-outline-danger"
        href="#"
        @click="userID = 0">
        Выйти
      </a>
      <a
        v-else
        class="btn btn-outline-primary"
        href="#"
        @click="userID = 1"
      >
        Войти
      </a>
    </header>
    <router-view />
  </div>
</template>

<script>
export default {
  name: 'App',
  data: function () {
    return {
      userID: 1
    }
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
</style>
