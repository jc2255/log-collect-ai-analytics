<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button type="primary" @click="handleAdd"><el-icon><Plus /></el-icon>新增角色</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="角色名称" width="150" />
      <el-table-column prop="code" label="角色标识" width="150" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" fixed="right" width="240">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="warning" @click="handleAssignMenus(row)">分配权限</el-button>
          <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
            <template #reference><el-button size="small" type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- 编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑角色' : '新增角色'" width="450px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="角色名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="角色标识"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配权限对话框 -->
    <el-dialog v-model="menuDialogVisible" title="分配菜单权限" width="450px">
      <el-tree
        ref="menuTreeRef"
        :data="menuTree"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        :default-checked-keys="checkedMenuIds"
      />
      <template #footer>
        <el-button @click="menuDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitMenus">确定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { roleApi, menuApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ name: '', code: '', description: '', sort: 0, status: 1 })

const menuDialogVisible = ref(false)
const menuTree = ref<any[]>([])
const checkedMenuIds = ref<number[]>([])
const currentRoleId = ref<number>(0)
const menuTreeRef = ref<any>()

async function fetchData() {
  loading.value = true
  try {
    const res: any = await roleApi.list({ page: 1, page_size: 100 })
    tableData.value = res.data.list || res.data || []
  } finally { loading.value = false }
}

function handleAdd() {
  editingId.value = null
  Object.assign(form, { name: '', code: '', description: '', sort: 0, status: 1 })
  dialogVisible.value = true
}
function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, { name: row.name, code: row.code, description: row.description || '', sort: row.sort || 0, status: row.status ?? 1 })
  dialogVisible.value = true
}
async function handleSubmit() {
  try {
    if (editingId.value) {
      await roleApi.update(editingId.value, form)
    } else {
      await roleApi.create(form)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}
async function handleDelete(id: number) {
  await roleApi.delete(id)
  ElMessage.success('删除成功')
  fetchData()
}

async function handleAssignMenus(row: any) {
  currentRoleId.value = row.id
  const res: any = await menuApi.list()
  menuTree.value = res.data || []
  checkedMenuIds.value = (row.menus || []).map((m: any) => m.id)
  menuDialogVisible.value = true
}
async function handleSubmitMenus() {
  const ids = menuTreeRef.value?.getCheckedKeys(false) || []
  await roleApi.assignMenus(currentRoleId.value, ids)
  ElMessage.success('权限分配成功')
  menuDialogVisible.value = false
  fetchData()
}

onMounted(fetchData)
</script>


