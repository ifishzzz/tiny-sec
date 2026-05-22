<template>
  <div class="gva-table-box">
    <el-table :data="tableData" row-key="ID">
      <el-table-column label="事件时间" prop="eventTime" min-width="180" />
      <el-table-column label="Agent" prop="agentId" min-width="180" />
      <el-table-column label="规则" prop="rule" min-width="200" />
      <el-table-column label="优先级" prop="priority" min-width="120" />
      <el-table-column label="来源" prop="source" min-width="120" />
      <el-table-column label="输出" prop="output" min-width="360" show-overflow-tooltip />
    </el-table>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { getFalcoEventList } from '@/api/falco'

defineOptions({
  name: 'FalcoEvents'
})

const tableData = ref([])

const loadData = async () => {
  const res = await getFalcoEventList({
    page: 1,
    pageSize: 50
  })
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

onMounted(() => {
  loadData()
})
</script>
