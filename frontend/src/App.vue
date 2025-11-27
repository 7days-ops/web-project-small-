<template>
  <div>
    <div v-if="!isAuthenticated">
      <Login v-if="showLogin" @login="handleLogin" @toggle="showLogin = false" />
      <Register v-else @register="handleLogin" @toggle="showLogin = true" />
    </div>
    <TaskManager v-else :user="user" @logout="handleLogout" />
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import Login from './components/Login.vue'
import Register from './components/Register.vue'
import TaskManager from './components/TaskManager.vue'

export default {
  name: 'App',
  components: {
    Login,
    Register,
    TaskManager
  },
  setup() {
    const isAuthenticated = ref(false)
    const showLogin = ref(true)
    const user = ref(null)

    onMounted(() => {
      const token = localStorage.getItem('token')
      const userData = localStorage.getItem('user')
      if (token && userData) {
        isAuthenticated.value = true
        user.value = JSON.parse(userData)
      }
    })

    const handleLogin = (data) => {
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
      isAuthenticated.value = true
      user.value = data.user
    }

    const handleLogout = () => {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      isAuthenticated.value = false
      user.value = null
    }

    return {
      isAuthenticated,
      showLogin,
      user,
      handleLogin,
      handleLogout
    }
  }
}
</script>
