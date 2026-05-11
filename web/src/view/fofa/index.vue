<template>
  <div>
    <WarningBar title="请先在参数管理中配置 `fofa_email` 和 `fofa_key`，查询请求会由后端代理发送到 FOFA。" />

    <div class="gva-search-box">
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="searchInfo"
        :rules="searchRules"
        class="demo-form-inline"
        @keyup.enter="onSubmit"
      >
        <el-form-item label="查询语法" prop="query" class="!mr-4">
          <el-input
            v-model="searchInfo.query"
            class="!w-[460px]"
            clearable
            placeholder='例如：app="nginx" && country="CN"'
          />
        </el-form-item>
        <el-form-item label="全量搜索" prop="full">
          <el-switch v-model="searchInfo.full" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" :loading="loading" @click="onSubmit">
            查询
          </el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <span class="text-sm text-gray-500">
          支持标准 FOFA 查询语法，默认显示常用字段并按分页加载结果。
        </span>
      </div>

      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        row-key="host"
      >
        <el-table-column align="left" label="Host" min-width="240">
          <template #default="{ row }">
            <a
              v-if="buildLink(row)"
              :href="buildLink(row)"
              target="_blank"
              rel="noopener noreferrer"
              class="text-primary hover:underline"
            >
              {{ row.host || '-' }}
            </a>
            <span v-else>{{ row.host || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="IP" prop="ip" min-width="140" />
        <el-table-column align="left" label="端口" prop="port" width="90" />
        <el-table-column align="left" label="协议" prop="protocol" width="100" />
        <el-table-column align="left" label="标题" prop="title" min-width="220" show-overflow-tooltip />
        <el-table-column align="left" label="域名" prop="domain" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="服务" prop="server" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="地区" min-width="140">
          <template #default="{ row }">
            {{ formatRegion(row) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import WarningBar from '@/components/warningBar/warningBar.vue'
import { fofaSearch } from '@/api/fofa'

defineOptions({
  name: 'FofaSearch'
})

const searchFormRef = ref()
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref([])

const searchInfo = reactive({
  query: '',
  full: false
})

const searchRules = reactive({
  query: [
    {
      required: true,
      message: '请输入 FOFA 查询语法',
      trigger: ['blur', 'change']
    },
    {
      validator: (_rule, value, callback) => {
        if (!value || !value.trim()) {
          callback(new Error('请输入 FOFA 查询语法'))
          return
        }
        callback()
      },
      trigger: ['blur', 'change']
    }
  ]
})

const getTableData = async () => {
  if (!searchInfo.query.trim()) {
    tableData.value = []
    total.value = 0
    return
  }

  loading.value = true
  try {
    const res = await fofaSearch({
      query: searchInfo.query.trim(),
      full: searchInfo.full,
      page: page.value,
      pageSize: pageSize.value
    })

    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
      page.value = res.data.page || page.value
      pageSize.value = res.data.pageSize || pageSize.value
      return
    }

    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const onSubmit = () => {
  searchFormRef.value?.validate(async (valid) => {
    if (!valid) return
    page.value = 1
    await getTableData()
  })
}

const onReset = () => {
  searchInfo.query = ''
  searchInfo.full = false
  page.value = 1
  pageSize.value = 10
  total.value = 0
  tableData.value = []
  searchFormRef.value?.clearValidate()
}

const handleCurrentChange = async (value) => {
  page.value = value
  await getTableData()
}

const handleSizeChange = async (value) => {
  pageSize.value = value
  page.value = 1
  await getTableData()
}

const buildLink = (row) => {
  if (!row.host) {
    return ''
  }

  if (row.host.startsWith('http://') || row.host.startsWith('https://')) {
    return row.host
  }

  if (row.protocol) {
    return `${row.protocol}://${row.host}`
  }

  return ''
}

const formatRegion = (row) => {
  return [row.country, row.city].filter(Boolean).join(' / ') || '-'
}
</script>
