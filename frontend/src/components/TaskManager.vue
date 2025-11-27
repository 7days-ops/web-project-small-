<template>
  <div class="tasks-container">
    <div class="header">
      <h2>Мои задачи</h2>
      <div class="user-info">
        <span>{{ user.username }}</span>
        <button class="btn-logout" @click="$emit('logout')">Выйти</button>
      </div>
    </div>

    <div class="task-form">
      <input v-model="newTask.title" placeholder="Название задачи" />
      <textarea v-model="newTask.description" placeholder="Описание задачи"></textarea>
      <button class="btn-add" @click="createTask">Добавить задачу</button>
    </div>

    <div v-if="tasks.length === 0" class="empty-state">
      <p>У вас пока нет задач</p>
    </div>

    <div v-else class="tasks-list">
      <div v-for="task in tasks" :key="task.id" class="task-item">
        <div class="task-header">
          <h3 class="task-title">{{ task.title }}</h3>
          <div class="task-actions">
            <button 
              v-if="task.status !== 'completed'" 
              class="btn-complete" 
              @click="completeTask(task.id)"
            >
              ✓ Завершить
            </button>
            <button class="btn-delete" @click="deleteTask(task.id)">✕ Удалить</button>
          </div>
        </div>
        <p class="task-description">{{ task.description }}</p>
        <span class="task-status" :class="`status-${task.status}`">
          {{ task.status === 'completed' ? 'Завершена' : 'В процессе' }}
        </span>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import axios from 'axios'

export default {
  name: 'TaskManager',
  props: {
    user: Object
  },
  emits: ['logout'],
  setup() {
    const tasks = ref([])
    const newTask = ref({
      title: '',
      description: ''
    })

    const getAuthHeader = () => {
      return {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`
        }
      }
    }

    const loadTasks = async () => {
      try {
        const response = await axios.get('http://localhost:8082/tasks', getAuthHeader())
        tasks.value = response.data
      } catch (err) {
        console.error('Ошибка загрузки задач:', err)
      }
    }

    const createTask = async () => {
      if (!newTask.value.title.trim()) return

      try {
        await axios.post('http://localhost:8082/tasks', newTask.value, getAuthHeader())
        newTask.value = { title: '', description: '' }
        await loadTasks()
      } catch (err) {
        console.error('Ошибка создания задачи:', err)
      }
    }

    const completeTask = async (id) => {
      try {
        await axios.put(`http://localhost:8082/tasks/${id}`, { status: 'completed' }, getAuthHeader())
        await loadTasks()
      } catch (err) {
        console.error('Ошибка обновления задачи:', err)
      }
    }

    const deleteTask = async (id) => {
      try {
        await axios.delete(`http://localhost:8082/tasks/${id}`, getAuthHeader())
        await loadTasks()
      } catch (err) {
        console.error('Ошибка удаления задачи:', err)
      }
    }

    onMounted(() => {
      loadTasks()
    })

    return {
      tasks,
      newTask,
      createTask,
      completeTask,
      deleteTask
    }
  }
}
</script>
