<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button type="primary" @click="handleAdd(0)"><el-icon><Plus /></el-icon>新增部门</el-button>
    </div>
    <el-table :data="deptTree" row-key="id" default-expand-all>
      <el-table-column prop="name" label="部门名称" width="200" />
      <el-table-column prop="leader" label="负责人" width="120" />
      <el-table-column prop="phone" label="联系电话" width="130" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="sort" label="排序" width="70" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '正常' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button size="small" @click="handleAdd(row.id)">新增</el-button>
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
            <template #reference><el-button size="small" type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑部门' : '新增部门'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="上级部门">
          <el-tree-select v-model="form.parent_id" :data="deptTreeWithRoot" :props="{ label: 'name', value: 'id', children: 'children' }" check-strictly clearable style="width:100%" />
        </el-form-item>
        <el-form-item label="部门名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="负责人"><el-input v-model="form.leader" /></el-form-item>
        <el-form-item label="联系电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { deptApi } from '../../api'

const deptTree = ref<any[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ parent_id: 0, name: '', leader: '', phone: '', email: '', sort: 0, status: 1 })

const deptTreeWithRoot = computed(() => [{ id: 0, name: '根部门', children: deptTree.value }])

async function fetchData() {
  const res: any = await deptApi.list()
  deptTree.value = res.data || []
}

function handleAdd(parentId: number) {
  editingId.value = null
  Object.assign(form, { parent_id: parentId, name: '', leader: '', phone: '', email: '', sort: 0, status: 1 })
  dialogVisible.value = true
}
function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, { parent_id: row.parent_id || 0, name: row.name, leader: row.leader || '', phone: row.phone || '', email: row.email || '', sort: row.sort || 0, status: row.status })
  dialogVisible.value = true
}
async function handleSubmit() {
  try {
    if (editingId.value) {
      await deptApi.update(editingId.value, form)
    } else {
      await deptApi.create(form)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}
async function handleDelete(id: number) {
  await deptApi.delete(id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>


