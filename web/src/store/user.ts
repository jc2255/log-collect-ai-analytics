import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '../api'

export interface UserInfo {
  id: number
  tenant_id: number
  tenant_name: string
  username: string
  nickname: string
  email: string
  phone: string
  avatar: string
  is_super_admin: boolean
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

  function isSuperAdmin() {
    return userInfo.value?.is_super_admin || false
  }

  return { token, userInfo, login, fetchUserInfo, logout, isSuperAdmin }
})
