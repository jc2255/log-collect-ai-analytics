<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button type="primary" @click="handleAdd"><el-icon><Plus /></el-icon>新增岗位</el-button>
    </div>
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="岗位名称" width="150" />
      <el-table-column prop="code" label="岗位编码" width="150" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '正常' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
            <template #reference><el-button size="small" type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑岗位' : '新增岗位'" width="450px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="岗位名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="岗位编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">正常</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { postApi } from '../../api'

const loading = ref(false)
const tableData = ref<any[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ name: '', code: '', sort: 0, status: 1 })

async function fetchData() {
  loading.value = true
  try {
    const res: any = await postApi.list({ page: 1, page_size: 100 })
    tableData.value = res.data.list || res.data || []
  } finally { loading.value = false }
}

function handleAdd() {
  editingId.value = null
  Object.assign(form, { name: '', code: '', sort: 0, status: 1 })
  dialogVisible.value = true
}
function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, { name: row.name, code: row.code, sort: row.sort || 0, status: row.status })
  dialogVisible.value = true
}
async function handleSubmit() {
  try {
    if (editingId.value) {
      await postApi.update(editingId.value, form)
    } else {
      await postApi.create(form)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}
async function handleDelete(id: number) {
  await postApi.delete(id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>


