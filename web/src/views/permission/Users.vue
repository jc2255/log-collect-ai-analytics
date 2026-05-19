<template>
  <div class="users-page">
    <el-row :gutter="20">
      <!-- 左侧部门树 -->
      <el-col :span="5">
        <el-card shadow="never" class="dept-card">
          <template #header>
            <div class="card-header">
              <el-icon><OfficeBuilding /></el-icon>
              <span>部门</span>
            </div>
          </template>
          <el-tree
            :data="deptTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            default-expand-all
            highlight-current
            @node-click="handleDeptClick"
          />
        </el-card>
      </el-col>
      <!-- 右侧用户列表 -->
      <el-col :span="19">
        <el-card shadow="never">
          <div class="page-toolbar">
            <el-input v-model="query.keyword" placeholder="搜索用户名/昵称" style="width: 220px" clearable @clear="fetchData" @keyup.enter="fetchData" prefix-icon="Search" />
            <el-select v-model="query.status" placeholder="状态" clearable style="width: 130px" @change="fetchData">
              <el-option label="正常" :value="1" />
              <el-option label="停用" :value="0" />
            </el-select>
            <el-button type="primary" @click="fetchData">
              <el-icon><Search /></el-icon>查询
            </el-button>
            <div style="flex:1"></div>
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>新增
            </el-button>
          </div>
          <el-table :data="tableData" v-loading="loading" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="username" label="用户名" width="120" />
            <el-table-column prop="nickname" label="昵称" width="120" />
            <el-table-column prop="dept_id" label="部门" width="100" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-switch :model-value="row.status === 1" @change="(val: boolean) => handleStatusChange(row, val)" />
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" fixed="right" width="220">
              <template #default="{ row }">
                <el-button size="small" @click="handleEdit(row)">编辑</el-button>
                <el-button size="small" @click="handleResetPwd(row)">重置密码</el-button>
                <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
                  <template #reference><el-button size="small" type="danger">删除</el-button></template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="query.page"
            v-model:page-size="query.page_size"
            :total="total"
            layout="total, prev, pager, next"
            style="margin-top: 16px"
            @current-change="fetchData"
          />
        </el-card>
      </el-col>
    </el-row>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑用户' : '新增用户'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名"><el-input v-model="form.username" :disabled="!!editingId" /></el-form-item>
        <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item label="密码" v-if="!editingId"><el-input v-model="form.password" type="password" placeholder="默认123456" /></el-form-item>
        <el-form-item label="部门">
          <el-tree-select v-model="form.dept_id" :data="deptTree" :props="{ label: 'name', value: 'id', children: 'children' }" check-strictly clearable style="width:100%" />
        </el-form-item>
        <el-form-item label="岗位">
          <el-select v-model="form.post_id" clearable style="width:100%">
            <el-option v-for="p in postList" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role_ids" multiple style="width:100%">
            <el-option v-for="r in roleList" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { userApi, deptApi, postApi, roleApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 20, keyword: '', status: undefined as number | undefined, dept_id: undefined as number | undefined })
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ username: '', nickname: '', password: '', dept_id: undefined as number | undefined, post_id: undefined as number | undefined, role_ids: [] as number[], status: 1 })
const deptTree = ref<any[]>([])
const postList = ref<any[]>([])
const roleList = ref<any[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res: any = await userApi.list(query)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally { loading.value = false }
}

async function fetchDepts() {
  const res: any = await deptApi.list()
  deptTree.value = res.data || []
}
async function fetchPosts() {
  const res: any = await postApi.list({ page: 1, page_size: 100 })
  postList.value = res.data.list || res.data || []
}
async function fetchRoles() {
  const res: any = await roleApi.list({ page: 1, page_size: 100 })
  roleList.value = res.data.list || res.data || []
}

function handleDeptClick(data: any) {
  query.dept_id = data.id
  fetchData()
}

function handleAdd() {
  editingId.value = null
  Object.assign(form, { username: '', nickname: '', password: '', dept_id: undefined, post_id: undefined, role_ids: [], status: 1 })
  dialogVisible.value = true
}
function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, { username: row.username, nickname: row.nickname, password: '', dept_id: row.dept_id || undefined, post_id: row.post_id || undefined, role_ids: (row.roles || []).map((r: any) => r.id), status: row.status ?? 1 })
  dialogVisible.value = true
}
async function handleSubmit() {
  try {
    if (editingId.value) {
      await userApi.update(editingId.value, form)
    } else {
      await userApi.create({ ...form, password: form.password || '123456' })
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}
async function handleDelete(id: number) {
  await userApi.delete(id)
  ElMessage.success('删除成功')
  fetchData()
}
async function handleStatusChange(row: any, val: boolean) {
  await userApi.updateStatus(row.id, { status: val ? 1 : 0 })
  ElMessage.success('操作成功')
  fetchData()
}
async function handleResetPwd(row: any) {
  await userApi.resetPassword(row.id, { password: '123456' })
  ElMessage.success('密码已重置为 123456')
}

onMounted(() => { fetchData(); fetchDepts(); fetchPosts(); fetchRoles() })
</script>

<style scoped>
.dept-card :deep(.el-card__header) {
  padding: 12px 16px;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--tech-text-primary);
}
</style>
