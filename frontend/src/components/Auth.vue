<template>
  <div class="container">
    <form
      action="#"
      @keydown.enter="signin">
      <div class="form-group">
        <label for="email">Email</label>
        <input
          id="email"
          v-model="email"
          type="email"
          class="form-control"
          placeholder="Введите адрес эл почты">
      </div>
      <div class="form-group">
        <label for="password">Пароль</label>
        <input
          id="password"
          v-model="password"
          type="password"
          class="form-control"
          placeholder="Введите пароль">
      </div>
      <button
        type="submit"
        class="btn btn-primary"
        @click.stop.prevent="signin()">
        Вход
      </button>
      <button
        type="submit"
        action="signup"
        class="btn btn-primary"
        @click.stop.prevent="signup()">
        Регистрация
      </button>
    </form>
  </div>
</template>

<script>
import { signIn, signUp } from '@/api'

export default {
  name: 'Auth',
  data: function () {
    return {
      email: '',
      password: ''
    }
  },

  methods: {
    signin () {
      signIn(this.email, this.password).then(() => {
        this.$router.push({ name: 'MainPage' })
        this.$router.go()
      })
    },

    signup () {
      signUp(this.email, this.password).then(() => {
        this.$router.push({ name: 'MainPage' })
        this.$router.go()
      })
    }
  }
}
</script>
