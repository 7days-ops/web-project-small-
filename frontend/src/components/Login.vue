<template>
  <div class="auth-container">
    <h2>Вход</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <label>Имя пользователя</label>
        <input v-model="username" type="text" required />
      </div>
      <div class="form-group">
        <label>Пароль</label>
        <input v-model="password" type="password" required />
      </div>
      <button type="submit" class="btn btn-primary">Войти</button>
      <button type="button" class="btn btn-secondary" @click="$emit('toggle')">
        Регистрация
      </button>
    </form>
  </div>
</template>

<script>
import { ref } from 'vue'
import axios from 'axios'

export default {
  name: 'Login',
  emits: ['login', 'toggle'],
  setup(props, { emit }) {
    const username = ref('')
    const password = ref('')
    const error = ref('')

    const handleSubmit = async () => {
      try {
        error.value = ''
        const response = await axios.post('http://localhost:8080/login', {
          username: username.value,
          password: password.value
        })
        emit('login', response.data)
      } catch (err) {
        error.value = err.response?.data || 'Ошибка входа'
      }
    }

    return {
      username,
      password,
      error,
      handleSubmit
    }
  }
}
</script>
