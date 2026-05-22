<template>
  <div>
    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
      <el-card v-for="item in cards" :key="item.label" shadow="hover">
        <div class="text-sm text-gray-500">{{ item.label }}</div>
        <div class="mt-3 text-2xl font-semibold text-gray-800">{{ item.value }}</div>
      </el-card>
    </div>

    <div class="gva-table-box">
      <el-alert
        title="当前为主机一期方案，围绕主机、Agent、规则、任务和事件中心建设，K8s 暂不纳入本期。"
        type="info"
        :closable="false"
        show-icon
      />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive } from 'vue'
import { getFalcoDashboard } from '@/api/falco'

defineOptions({
  name: 'FalcoDashboard'
})

const dashboard = reactive({
  hostTotal: 0,
  onlineHostTotal: 0,
  agentTotal: 0,
  pendingTaskTotal: 0,
  rulePackageTotal: 0,
  eventTotal: 0,
  criticalEventTotal: 0
})

const cards = computed(() => [
  { label: '主机总数', value: dashboard.hostTotal },
  { label: '在线主机', value: dashboard.onlineHostTotal },
  { label: 'Agent 总数', value: dashboard.agentTotal },
  { label: '待处理任务', value: dashboard.pendingTaskTotal },
  { label: '规则包数', value: dashboard.rulePackageTotal },
  { label: '事件总数', value: dashboard.eventTotal },
  { label: '高危事件', value: dashboard.criticalEventTotal }
])

const loadDashboard = async () => {
  const res = await getFalcoDashboard()
  if (res.code === 0 && res.data) {
    Object.assign(dashboard, res.data)
  }
}

onMounted(() => {
  loadDashboard()
})
</script>
