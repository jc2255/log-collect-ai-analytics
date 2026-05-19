import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '../api'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  dept_id: number
  post_id: number
  status: number
  roles: { id: number; name: string; code: string }[]
}

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('lca_token') || '')
  const userInfo = ref<UserInfo | null>(null)

  async function login(username: string, password: string) {
    const res: any = await authApi.login({ username, password })
    token.value = res.data.token
    userInfo.value = res.data.user_info
    localStorage.setItem('lca_token', res.data.token)
    return res
  }

  async function fetchUserInfo() {
    const res: any = await authApi.getUserInfo()
    userInfo.value = res.data
    return res.data
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('lca_token')
  }

  function isAdmin() {
    return userInfo.value?.username === 'admin'
  }

  return { token, userInfo, login, fetchUserInfo, logout, isAdmin }
})
