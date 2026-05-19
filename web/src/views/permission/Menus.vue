<template>
  <el-card shadow="never">
    <div class="page-toolbar">
      <el-button type="primary" @click="handleAdd(0)"><el-icon><Plus /></el-icon>新增菜单</el-button>
    </div>
    <el-table :data="menuTree" row-key="id" default-expand-all>
      <el-table-column prop="name" label="菜单名称" width="200" />
      <el-table-column prop="menu_type" label="类型" width="80">
        <template #default="{ row }">
          <el-tag v-if="row.menu_type === 'M'" type="primary">目录</el-tag>
          <el-tag v-else-if="row.menu_type === 'C'" type="success">菜单</el-tag>
          <el-tag v-else type="info">按钮</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="icon" label="图标" width="80" />
      <el-table-column prop="path" label="路径" />
      <el-table-column prop="perms" label="权限标识" width="150" />
      <el-table-column prop="sort" label="排序" width="70" />
      <el-table-column prop="visible" label="显示" width="70">
        <template #default="{ row }">
          <el-tag :type="row.visible === 1 ? 'success' : 'info'">{{ row.visible === 1 ? '显示' : '隐藏' }}</el-tag>
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

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑菜单' : '新增菜单'" width="550px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="上级菜单">
          <el-tree-select v-model="form.parent_id" :data="menuTreeWithRoot" :props="{ label: 'name', value: 'id', children: 'children' }" check-strictly clearable style="width:100%" />
        </el-form-item>
        <el-form-item label="菜单类型">
          <el-radio-group v-model="form.menu_type">
            <el-radio value="M">目录</el-radio>
            <el-radio value="C">菜单</el-radio>
            <el-radio value="F">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="图标"><el-input v-model="form.icon" /></el-form-item>
        <el-form-item label="路由路径"><el-input v-model="form.path" /></el-form-item>
        <el-form-item label="权限标识"><el-input v-model="form.perms" placeholder="如：system:user:list" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="是否显示">
            <el-radio-group v-model="form.visible">
              <el-radio :value="1">显示</el-radio>
              <el-radio :value="0">隐藏</el-radio>
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
import { menuApi } from '../../api'

const menuTree = ref<any[]>([])
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const form = reactive({ parent_id: 0, menu_type: 'C', name: '', icon: '', path: '', perms: '', sort: 0, visible: 1 })

const menuTreeWithRoot = computed(() => [{ id: 0, name: '根目录', children: menuTree.value }])

async function fetchData() {
  const res: any = await menuApi.list()
  menuTree.value = res.data || []
}

function handleAdd(parentId: number) {
  editingId.value = null
  Object.assign(form, { parent_id: parentId, menu_type: 'C', name: '', icon: '', path: '', perms: '', sort: 0, visible: 1 })
  dialogVisible.value = true
}
function handleEdit(row: any) {
  editingId.value = row.id
  Object.assign(form, { parent_id: row.parent_id || 0, menu_type: row.menu_type || 'C', name: row.name, icon: row.icon || '', path: row.path || '', perms: row.perms || '', sort: row.sort || 0, visible: row.visible })
  dialogVisible.value = true
}
async function handleSubmit() {
  try {
    if (editingId.value) {
      await menuApi.update(editingId.value, form)
    } else {
      await menuApi.create(form)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchData()
  } catch { /* */ }
}
async function handleDelete(id: number) {
  await menuApi.delete(id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>


