<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="taskForm">
        <el-form-item label="目标主机">
          <el-select v-model="taskForm.hostId" placeholder="请选择主机" class="w-[180px]" clearable>
            <el-option
              v-for="item in hostOptions"
              :key="item.ID"
              :label="`${item.name || item.hostname} (${item.ip || '-'})`"
              :value="item.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="任务类型">
          <el-select v-model="taskForm.taskType" class="w-[160px]">
            <el-option label="安装" value="install" />
            <el-option label="升级" value="upgrade" />
            <el-option label="回滚" value="rollback" />
            <el-option label="重载" value="reload" />
            <el-option label="重启" value="restart" />
          </el-select>
        </el-form-item>
        <el-form-item label="Falco 版本">
          <el-input v-model="taskForm.falcoVersion" placeholder="如 0.40.0" clearable />
        </el-form-item>
        <el-form-item label="规则包版本">
          <el-input v-model="taskForm.rulePackageVersion" placeholder="如 rules-20260521" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="plus" @click="createTask">创建任务</el-button>
          <el-button icon="refresh" @click="loadData">刷新</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="任务ID" prop="ID" min-width="90" />
        <el-table-column label="请求ID" prop="requestId" min-width="200" show-overflow-tooltip />
        <el-table-column label="主机ID" prop="hostId" min-width="100" />
        <el-table-column label="任务类型" prop="taskType" min-width="160" />
        <el-table-column label="动作" prop="action" min-width="180" />
        <el-table-column label="状态" prop="status" min-width="120" />
        <el-table-column label="阶段" prop="stage" min-width="140" />
        <el-table-column label="操作人" prop="operator" min-width="120" />
        <el-table-column label="负载" prop="payload" min-width="260" show-overflow-tooltip />
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import {
  createFalcoInstallTask,
  createFalcoReloadTask,
  createFalcoRestartTask,
  createFalcoRollbackTask,
  createFalcoUpgradeTask,
  getFalcoHostList,
  getFalcoTaskList
} from '@/api/falco'

defineOptions({
  name: 'FalcoTasks'
})

const tableData = ref([])
const hostOptions = ref([])
const taskForm = reactive({
  hostId: undefined,
  taskType: 'install',
  falcoVersion: '0.40.0',
  rulePackageVersion: 'default'
})

const taskCreators = {
  install: createFalcoInstallTask,
  upgrade: createFalcoUpgradeTask,
  rollback: createFalcoRollbackTask,
  reload: createFalcoReloadTask,
  restart: createFalcoRestartTask
}

const loadData = async () => {
  const res = await getFalcoTaskList({
    page: 1,
    pageSize: 50
  })
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const loadHosts = async () => {
  const res = await getFalcoHostList({
    page: 1,
    pageSize: 100
  })
  if (res.code === 0) {
    hostOptions.value = res.data.list || []
    if (!taskForm.hostId && hostOptions.value.length > 0) {
      taskForm.hostId = hostOptions.value[0].ID
    }
  }
}

const createTask = async () => {
  if (!taskForm.hostId) {
    ElMessage.warning('请先选择主机')
    return
  }
  const request = taskCreators[taskForm.taskType]
  const res = await request({
    hostId: taskForm.hostId,
    falcoVersion: taskForm.falcoVersion,
    rulePackageVersion: taskForm.rulePackageVersion
  })
  if (res.code === 0) {
    ElMessage.success('创建成功')
    loadData()
  }
}

onMounted(() => {
  loadHosts()
  loadData()
})
</script>
