<template>
  <div class="auth-container">
    <h2>Регистрация</h2>
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
      <button type="submit" class="btn btn-primary">Зарегистрироваться</button>
      <button type="button" class="btn btn-secondary" @click="$emit('toggle')">
        Уже есть аккаунт?
      </button>
    </form>
  </div>
</template>

<script>
import { ref } from 'vue'
import axios from 'axios'

export default {
  name: 'Register',
  emits: ['register', 'toggle'],
  setup(props, { emit }) {
    const username = ref('')
    const password = ref('')
    const error = ref('')

    const handleSubmit = async () => {
      try {
        error.value = ''
        const response = await axios.post('http://localhost:8080/register', {
          username: username.value,
          password: password.value
        })
        emit('register', response.data)
      } catch (err) {
        error.value = err.response?.data || 'Ошибка регистрации'
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
