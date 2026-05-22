<template>
  <div class="gva-table-box">
    <el-table :data="tableData" row-key="ID">
      <el-table-column label="Agent ID" prop="agentId" min-width="220" />
      <el-table-column label="主机ID" prop="hostId" min-width="100" />
      <el-table-column label="版本" prop="version" min-width="120" />
      <el-table-column label="状态" min-width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'online' ? 'success' : 'info'">
            {{ row.status || '-' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最后心跳" prop="lastHeartbeatAt" min-width="180" />
      <el-table-column label="最后状态上报" prop="lastReportedAt" min-width="180" />
    </el-table>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { getFalcoAgentList } from '@/api/falco'

defineOptions({
  name: 'FalcoAgents'
})

const tableData = ref([])

const loadData = async () => {
  const res = await getFalcoAgentList({
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
