<template>
  <div class="gva-table-box">
    <el-alert
      title="说明：install-falco-agent.sh 安装的是我们平台的 Agent；falco-agent.sh 是这个 Agent 的运行脚本。真正安装或升级 Falco 软件包，是 Agent 注册后再执行平台下发的 install/upgrade 任务。"
      type="info"
      :closable="false"
      class="mb-4"
    />

    <el-form :model="form" label-width="120px" class="max-w-[760px]">
      <el-form-item label="Enroll Key">
        <el-input v-model="form.enrollKey" />
      </el-form-item>
      <el-form-item label="事件保留天数">
        <el-input-number v-model="form.eventKeepDays" :min="1" />
      </el-form-item>
      <el-form-item label="规则同步模式">
        <el-select v-model="form.ruleSyncMode" class="w-full">
          <el-option label="手动" value="manual" />
          <el-option label="自动" value="auto" />
        </el-select>
      </el-form-item>
      <el-form-item label="安装命令">
        <div class="w-full">
          <el-input
            :model-value="installCommand"
            type="textarea"
            :rows="4"
            readonly
          />
          <div class="mt-3 flex gap-2">
            <el-button type="primary" @click="copyInstallCommand">复制安装命令</el-button>
            <el-button @click="saveSettings">保存设置</el-button>
          </div>
        </div>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { computed, onMounted, reactive } from 'vue'
import {
  buildFalcoAgentInstallCommand,
  getFalcoSettings,
  updateFalcoSettings
} from '@/api/falco'

defineOptions({
  name: 'FalcoSettings'
})

const form = reactive({
  enrollKey: '',
  eventKeepDays: 7,
  ruleSyncMode: 'manual'
})

const installCommand = computed(() =>
  buildFalcoAgentInstallCommand({
    enrollKey: form.enrollKey
  })
)

const loadSettings = async () => {
  const res = await getFalcoSettings()
  if (res.code === 0 && res.data) {
    Object.assign(form, res.data)
  }
}

const saveSettings = async () => {
  const res = await updateFalcoSettings(form)
  if (res.code === 0) {
    ElMessage.success('保存成功')
    loadSettings()
  }
}

const copyInstallCommand = async () => {
  await navigator.clipboard.writeText(installCommand.value)
  ElMessage.success('安装命令已复制')
}

onMounted(() => {
  loadSettings()
})
</script>
