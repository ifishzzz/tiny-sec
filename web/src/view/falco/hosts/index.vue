<template>
  <div>
    <el-alert
      title="Agent 安装说明：这里的“复制安装命令”是先把我们平台 Agent 装到主机上；Agent 在线后，再用“安装/升级/回滚/重载/重启”给目标主机下发 Falco 任务。"
      type="info"
      :closable="false"
      class="mb-4"
    />

    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="关键字">
          <el-input v-model="searchInfo.keyword" placeholder="主机名/IP/实例ID" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" placeholder="全部" clearable>
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="主机名称" prop="name" min-width="160" />
        <el-table-column label="Hostname" prop="hostname" min-width="160" />
        <el-table-column label="IP" prop="ip" min-width="140" />
        <el-table-column label="云厂商" prop="provider" min-width="120" />
        <el-table-column label="区域" prop="region" min-width="120" />
        <el-table-column label="实例ID" prop="instanceId" min-width="180" />
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'info'">
              {{ row.status || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="快捷操作" min-width="360" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-2">
              <el-button link type="info" @click="copyInstallCommand(row)">复制安装命令</el-button>
              <el-button link type="primary" @click="dispatchTask(row, 'install')">安装</el-button>
              <el-button link type="primary" @click="dispatchTask(row, 'upgrade')">升级</el-button>
              <el-button link type="warning" @click="dispatchTask(row, 'rollback')">回滚</el-button>
              <el-button link type="success" @click="dispatchTask(row, 'reload')">重载</el-button>
              <el-button link type="danger" @click="dispatchTask(row, 'restart')">重启</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import {
  buildFalcoAgentInstallCommand,
  createFalcoInstallTask,
  createFalcoReloadTask,
  createFalcoRestartTask,
  createFalcoRollbackTask,
  createFalcoUpgradeTask,
  getFalcoSettings,
  getFalcoHostList
} from '@/api/falco'

defineOptions({
  name: 'FalcoHosts'
})

const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref([])
const falcoSettings = ref({
  enrollKey: ''
})
const searchInfo = reactive({
  keyword: '',
  status: ''
})

const taskCreators = {
  install: createFalcoInstallTask,
  upgrade: createFalcoUpgradeTask,
  rollback: createFalcoRollbackTask,
  reload: createFalcoReloadTask,
  restart: createFalcoRestartTask
}

const getTableData = async () => {
  const res = await getFalcoHostList({
    page: page.value,
    pageSize: pageSize.value,
    ...searchInfo
  })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const loadSettings = async () => {
  const res = await getFalcoSettings()
  if (res.code === 0 && res.data) {
    falcoSettings.value = res.data
  }
}

const onSubmit = () => {
  page.value = 1
  getTableData()
}

const onReset = () => {
  searchInfo.keyword = ''
  searchInfo.status = ''
  page.value = 1
  getTableData()
}

const handleCurrentChange = (value) => {
  page.value = value
  getTableData()
}

const handleSizeChange = (value) => {
  pageSize.value = value
  page.value = 1
  getTableData()
}

const dispatchTask = async (row, taskType) => {
  const request = taskCreators[taskType]
  if (!request) {
    return
  }
  const res = await request({
    hostId: row.ID,
    falcoVersion: row.agentVersion || '0.40.0',
    rulePackageVersion: 'default'
  })
  if (res.code === 0) {
    ElMessage.success(`${taskType} 任务已创建`)
  }
}

const copyInstallCommand = async (row) => {
  const command = buildFalcoAgentInstallCommand({
    enrollKey: falcoSettings.value.enrollKey,
    provider: row.provider || 'aws',
    region: row.region || '',
    agentId: row.agentId || row.hostname || row.name || ''
  })
  await navigator.clipboard.writeText(command)
  ElMessage.success(
    falcoSettings.value.enrollKey
      ? '安装命令已复制'
      : '安装命令已复制，但请先到系统设置保存 Enroll Key'
  )
}

onMounted(() => {
  loadSettings()
  getTableData()
})
</script>
