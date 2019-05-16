<template>
  <div id="app">
    <header class="d-flex flex-column flex-md-row align-items-center p-3 px-md-4 mb-3 bg-white border-bottom box-shadow">
      <h5 class="my-0 mr-md-auto font-weight-normal">
        <router-link :to="{ name: 'MainPage'}">
          Кукаремуви
        </router-link>
      </h5>

      <a class="btn btn-outline-danger" href="#" v-on:click="userID=0" v-if="userID">
        Выйти
      </a>
      <a class="btn btn-outline-primary" href="#" v-on:click="userID=1" v-else>
        Войти
      </a>
    </header>
    <router-view/>
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
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

.box-shadow {
  box-shadow: 0 .25rem .75rem rgba(0, 0, 0, .05);
}
</style>
