<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="plus" @click="createDefaultRule">创建默认规则包</el-button>
      </div>

      <el-table :data="tableData" row-key="ID">
        <el-table-column label="名称" prop="name" min-width="180" />
        <el-table-column label="版本" prop="version" min-width="120" />
        <el-table-column label="状态" prop="status" min-width="120" />
        <el-table-column label="说明" prop="description" min-width="260" show-overflow-tooltip />
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { onMounted, ref } from 'vue'
import { createFalcoRulePackage, getFalcoRulePackageList } from '@/api/falco'

defineOptions({
  name: 'FalcoRules'
})

const tableData = ref([])

const loadData = async () => {
  const res = await getFalcoRulePackageList({
    page: 1,
    pageSize: 50
  })
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const createDefaultRule = async () => {
  const res = await createFalcoRulePackage({
    name: 'falco-host-default',
    version: 'v1',
    status: 'draft',
    description: '主机一期默认规则包',
    content: '- rule: Placeholder rule',
    checksum: ''
  })
  if (res.code === 0) {
    ElMessage.success('创建成功')
    loadData()
  }
}

onMounted(() => {
  loadData()
})
</script>
